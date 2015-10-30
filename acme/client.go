package acme

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Logger is used to log errors; if nil, the default log.Logger is used.
var Logger *log.Logger

// logger is an helper function to retrieve the available logger
func logger() *log.Logger {
	if Logger == nil {
		Logger = log.New(os.Stderr, "", log.LstdFlags)
	}
	return Logger
}

// User interface is to be implemented by users of this library.
// It is used by the client type to get user specific information.
type User interface {
	GetEmail() string
	GetRegistration() *RegistrationResource
	GetPrivateKey() *rsa.PrivateKey
}

// Interface for all challenge solvers to implement.
type solver interface {
	Solve(challenge challenge, domain string) error
}

// Client is the user-friendy way to ACME
type Client struct {
	directory  directory
	user       User
	jws        *jws
	keyBits    int
	issuerCert []byte
	solvers    map[string]solver
}

// NewClient creates a new client for the set user.
// caURL - The root url to the boulder instance you want certificates from
// usr - A filled in user struct
// keyBits - Size of the key in bits
// optPort - The alternative port to listen on for challenges.
func NewClient(caURL string, usr User, keyBits int, optPort string) (*Client, error) {
	privKey := usr.GetPrivateKey()
	if privKey == nil {
		return nil, errors.New("private key was nil")
	}

	if err := privKey.Validate(); err != nil {
		return nil, fmt.Errorf("invalid private key: %v", err)
	}

	jws := &jws{privKey: privKey}

	// REVIEW: best possibility?
	// Add all available solvers with the right index as per ACME
	// spec to this map. Otherwise they won`t be found.
	solvers := make(map[string]solver)
	solvers["simpleHttp"] = &simpleHTTPChallenge{jws: jws, optPort: optPort}

	dirURL := caURL + "/directory"
	dirResp, err := http.Get(dirURL)
	if err != nil {
		return nil, fmt.Errorf("get directory at '%s': %v", dirURL, err)
	}
	defer dirResp.Body.Close()

	var dir directory
	err = json.NewDecoder(dirResp.Body).Decode(&dir)
	if err != nil {
		return nil, fmt.Errorf("decode directory: %v", err)
	}

	if dir.NewRegURL == "" {
		return nil, errors.New("directory missing new registration URL")
	}
	if dir.NewAuthzURL == "" {
		return nil, errors.New("directory missing new authz URL")
	}
	if dir.NewCertURL == "" {
		return nil, errors.New("directory missing new certificate URL")
	}
	if dir.RevokeCertURL == "" {
		return nil, errors.New("directory missing revoke certificate URL")
	}

	return &Client{directory: dir, user: usr, jws: jws, keyBits: keyBits, solvers: solvers}, nil
}

// Register the current account to the ACME server.
func (c *Client) Register() (*RegistrationResource, error) {
	logger().Print("Registering account ... ")

	regMsg := registrationMessage{
		Resource: "new-reg",
	}
	if c.user.GetEmail() != "" {
		regMsg.Contact = []string{"mailto:" + c.user.GetEmail()}
	} else {
		regMsg.Contact = []string{}
	}

	jsonBytes, err := json.Marshal(regMsg)
	if err != nil {
		return nil, err
	}

	resp, err := c.jws.post(c.directory.NewRegURL, jsonBytes)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return nil, handleHTTPError(resp)
	}

	var serverReg Registration
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&serverReg)
	if err != nil {
		return nil, err
	}

	reg := &RegistrationResource{Body: serverReg}

	links := parseLinks(resp.Header["Link"])
	reg.URI = resp.Header.Get("Location")
	if links["terms-of-service"] != "" {
		reg.TosURL = links["terms-of-service"]
	}

	if links["next"] != "" {
		reg.NewAuthzURL = links["next"]
	} else {
		return nil, errors.New("The server did not return enough information to proceed...")
	}

	return reg, nil
}

// AgreeToTOS updates the Client registration and sends the agreement to
// the server.
func (c *Client) AgreeToTOS() error {
	c.user.GetRegistration().Body.Agreement = c.user.GetRegistration().TosURL
	c.user.GetRegistration().Body.Resource = "reg"
	jsonBytes, err := json.Marshal(&c.user.GetRegistration().Body)
	if err != nil {
		return err
	}

	resp, err := c.jws.post(c.user.GetRegistration().URI, jsonBytes)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return handleHTTPError(resp)
	}

	return nil
}

// ObtainCertificates tries to obtain certificates from the CA server
// using the challenges it has configured. The returned certificates are
// PEM encoded byte slices.
// If bundle is true, the []byte contains both the issuer certificate and
// your issued certificate as a bundle.
func (c *Client) ObtainCertificates(domains []string, bundle bool) ([]CertificateResource, error) {
	logger().Print("Obtaining certificates...")
	challenges := c.getChallenges(domains)
	err := c.solveChallenges(challenges)
	if err != nil {
		return nil, err
	}

	logger().Print("Validations succeeded. Getting certificates")

	return c.requestCertificates(challenges, bundle)
}

// RevokeCertificate takes a PEM encoded certificate or bundle and tries to revoke it at the CA.
func (c *Client) RevokeCertificate(certificate []byte) error {
	certificates, err := parsePEMBundle(certificate)
	if err != nil {
		return err
	}

	x509Cert := certificates[0]
	if x509Cert.IsCA {
		return fmt.Errorf("Certificate bundle starts with a CA certificate")
	}

	encodedCert := base64.URLEncoding.EncodeToString(x509Cert.Raw)

	jsonBytes, err := json.Marshal(revokeCertMessage{Resource: "revoke-cert", Certificate: encodedCert})
	if err != nil {
		return err
	}

	resp, err := c.jws.post(c.directory.RevokeCertURL, jsonBytes)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return handleHTTPError(resp)
	}

	return nil
}

// RenewCertificate takes a CertificateResource and tries to renew the certificate.
// If the renewal process succeeds, the new certificate will ge returned in a new CertResource.
// Please be aware that this function will return a new certificate in ANY case that is not an error.
// If the server does not provide us with a new cert on a GET request to the CertURL
// this function will start a new-cert flow where a new certificate gets generated.
// If bundle is true, the []byte contains both the issuer certificate and
// your issued certificate as a bundle.
func (c *Client) RenewCertificate(cert CertificateResource, revokeOld bool, bundle bool) (CertificateResource, error) {
	// Input certificate is PEM encoded. Decode it here as we may need the decoded
	// cert later on in the renewal process. The input may be a bundle or a single certificate.
	certificates, err := parsePEMBundle(cert.Certificate)
	if err != nil {
		return CertificateResource{}, err
	}

	x509Cert := certificates[0]
	if x509Cert.IsCA {
		return CertificateResource{}, fmt.Errorf("[%s] Certificate bundle starts with a CA certificate", cert.Domain)
	}

	// This is just meant to be informal for the user.
	timeLeft := x509Cert.NotAfter.Sub(time.Now().UTC())
	logger().Printf("[%s] Trying to renew certificate with %d hours remaining.", cert.Domain, int(timeLeft.Hours()))

	// The first step of renewal is to check if we get a renewed cert
	// directly from the cert URL.
	resp, err := http.Get(cert.CertURL)
	if err != nil {
		return CertificateResource{}, err
	}
	defer resp.Body.Close()
	serverCertBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return CertificateResource{}, err
	}

	serverCert, err := x509.ParseCertificate(serverCertBytes)
	if err != nil {
		return CertificateResource{}, err
	}

	// If the server responds with a different certificate we are effectively renewed.
	// TODO: Further test if we can actually use the new certificate (Our private key works)
	if !x509Cert.Equal(serverCert) {
		logger().Printf("[%s] The server responded with a renewed certificate.", cert.Domain)
		if revokeOld {
			c.RevokeCertificate(cert.Certificate)
		}
		issuedCert := pemEncode(derCertificateBytes(serverCertBytes))
		// If bundle is true, we want to return a certificate bundle.
		// To do this, we need the issuer certificate.
		if bundle {
			// The issuer certificate link is always supplied via an "up" link
			// in the response headers of a new certificate.
			links := parseLinks(resp.Header["Link"])
			issuerCert, err := c.getIssuerCertificate(links["up"])
			if err != nil {
				// If we fail to aquire the issuer cert, return the issued certificate - do not fail.
				logger().Printf("[%s] Could not bundle issuer certificate.\n%v", cert.Domain, err)
			} else {
				// Success - append the issuer cert to the issued cert.
				issuerCert = pemEncode(derCertificateBytes(issuerCert))
				issuedCert = append(issuedCert, issuerCert...)
				cert.Certificate = issuedCert
			}
		}

		cert.Certificate = issuedCert
		return cert, nil
	}

	newCerts, err := c.ObtainCertificates([]string{cert.Domain}, bundle)
	if err != nil {
		return CertificateResource{}, err
	}

	if revokeOld {
		c.RevokeCertificate(cert.Certificate)
	}

	return newCerts[0], nil
}

// Looks through the challenge combinations to find a solvable match.
// Then solves the challenges in series and returns.
func (c *Client) solveChallenges(challenges []*authorizationResource) error {
	// loop through the resources, basically through the domains.
	for _, authz := range challenges {
		// no solvers - no solving
		if solvers := c.chooseSolvers(authz.Body, authz.Domain); solvers != nil {
			for i, solver := range solvers {
				// TODO: do not immediately fail if one domain fails to validate.
				err := solver.Solve(authz.Body.Challenges[i], authz.Domain)
				if err != nil {
					return err
				}
			}
		} else {
			return fmt.Errorf("Could not determine solvers for %s", authz.Domain)
		}
	}

	return nil
}

// Checks all combinations from the server and returns an array of
// solvers which should get executed in series.
func (c *Client) chooseSolvers(auth authorization, domain string) map[int]solver {
	for _, combination := range auth.Combinations {
		solvers := make(map[int]solver)
		for _, idx := range combination {
			if solver, ok := c.solvers[auth.Challenges[idx].Type]; ok {
				solvers[idx] = solver
			} else {
				logger().Printf("Could not find solver for: %s", auth.Challenges[idx].Type)
			}
		}

		// If we can solve the whole combination, return the solvers
		if len(solvers) == len(combination) {
			return solvers
		}
	}
	return nil
}

// Get the challenges needed to proof our identifier to the ACME server.
func (c *Client) getChallenges(domains []string) []*authorizationResource {
	resc, errc := make(chan *authorizationResource), make(chan error)

	for _, domain := range domains {
		go func(domain string) {
			jsonBytes, err := json.Marshal(authorization{Resource: "new-authz", Identifier: identifier{Type: "dns", Value: domain}})
			if err != nil {
				errc <- err
				return
			}

			resp, err := c.jws.post(c.user.GetRegistration().NewAuthzURL, jsonBytes)
			if err != nil {
				errc <- err
				return
			}

			if resp.StatusCode != http.StatusCreated {
				errc <- handleHTTPError(resp)
			}

			links := parseLinks(resp.Header["Link"])
			if links["next"] == "" {
				logger().Println("The server did not provide enough information to proceed.")
				return
			}

			var authz authorization
			decoder := json.NewDecoder(resp.Body)
			err = decoder.Decode(&authz)
			if err != nil {
				errc <- err
			}
			resp.Body.Close()

			resc <- &authorizationResource{Body: authz, NewCertURL: links["next"], AuthURL: resp.Header.Get("Location"), Domain: domain}
		}(domain)
	}

	var responses []*authorizationResource
	for i := 0; i < len(domains); i++ {
		select {
		case res := <-resc:
			responses = append(responses, res)
		case err := <-errc:
			logger().Printf("%v", err)
		}
	}

	close(resc)
	close(errc)

	return responses
}

// requestCertificates iterates all granted authorizations, creates RSA private keys and CSRs.
// It then uses these to request a certificate from the CA and returns the list of successfully
// granted certificates.
func (c *Client) requestCertificates(challenges []*authorizationResource, bundle bool) ([]CertificateResource, error) {
	resc, errc := make(chan CertificateResource), make(chan error)
	for _, authz := range challenges {
		go c.requestCertificate(authz, resc, errc, bundle)
	}

	var certs []CertificateResource
	for i := 0; i < len(challenges); i++ {
		select {
		case res := <-resc:
			certs = append(certs, res)
		case err := <-errc:
			logger().Printf("%v", err)
		}
	}

	close(resc)
	close(errc)

	return certs, nil
}

func (c *Client) requestCertificate(authz *authorizationResource, result chan CertificateResource, errc chan error, bundle bool) {
	privKey, err := generatePrivateKey(rsakey, c.keyBits)
	if err != nil {
		errc <- err
		return
	}

	// TODO: should the CSR be customizable?
	csr, err := generateCsr(privKey.(*rsa.PrivateKey), authz.Domain)
	if err != nil {
		errc <- err
		return
	}

	csrString := base64.URLEncoding.EncodeToString(csr)
	jsonBytes, err := json.Marshal(csrMessage{Resource: "new-cert", Csr: csrString, Authorizations: []string{authz.AuthURL}})
	if err != nil {
		errc <- err
		return
	}

	resp, err := c.jws.post(authz.NewCertURL, jsonBytes)
	if err != nil {
		errc <- err
		return
	}

	privateKeyPem := pemEncode(privKey)
	cerRes := CertificateResource{
		Domain:     authz.Domain,
		CertURL:    resp.Header.Get("Location"),
		PrivateKey: privateKeyPem}

	for {

		switch resp.StatusCode {
		case 202:
		case 201:

			cert, err := ioutil.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				errc <- err
				return
			}

			// The server returns a body with a length of zero if the
			// certificate was not ready at the time this request completed.
			// Otherwise the body is the certificate.
			if len(cert) > 0 {

				cerRes.CertStableURL = resp.Header.Get("Content-Location")

				issuedCert := pemEncode(derCertificateBytes(cert))
				// If bundle is true, we want to return a certificate bundle.
				// To do this, we need the issuer certificate.
				if bundle {
					// The issuer certificate link is always supplied via an "up" link
					// in the response headers of a new certificate.
					links := parseLinks(resp.Header["Link"])
					issuerCert, err := c.getIssuerCertificate(links["up"])
					if err != nil {
						// If we fail to aquire the issuer cert, return the issued certificate - do not fail.
						logger().Printf("[%s] Could not bundle issuer certificate.\n%v", authz.Domain, err)
					} else {
						// Success - append the issuer cert to the issued cert.
						issuerCert = pemEncode(derCertificateBytes(issuerCert))
						issuedCert = append(issuedCert, issuerCert...)
					}
				}

				cerRes.Certificate = issuedCert
				logger().Printf("[%s] Server responded with a certificate.", authz.Domain)
				result <- cerRes
				return
			}

			// The certificate was granted but is not yet issued.
			// Check retry-after and loop.
			ra := resp.Header.Get("Retry-After")
			retryAfter, err := strconv.Atoi(ra)
			if err != nil {
				errc <- err
				return
			}

			logger().Printf("[%s] Server responded with status 202. Respecting retry-after of: %d", authz.Domain, retryAfter)
			time.Sleep(time.Duration(retryAfter) * time.Second)

			break
		default:
			errc <- handleHTTPError(resp)
			return
		}

		resp, err = http.Get(cerRes.CertURL)
		if err != nil {
			errc <- err
			return
		}
	}
}

// getIssuerCertificate requests the issuer certificate and caches it for
// subsequent requests.
func (c *Client) getIssuerCertificate(url string) ([]byte, error) {
	logger().Printf("Requesting issuer cert from: %s", url)
	if c.issuerCert != nil {
		return c.issuerCert, nil
	}

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	issuerBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	_, err = x509.ParseCertificate(issuerBytes)
	if err != nil {
		return nil, err
	}

	c.issuerCert = issuerBytes
	return issuerBytes, err
}

func logResponseHeaders(resp *http.Response) {
	logger().Println(resp.Status)
	for k, v := range resp.Header {
		logger().Printf("-- %s: %s", k, v)
	}
}

func logResponseBody(resp *http.Response) {
	body, _ := ioutil.ReadAll(resp.Body)
	logger().Printf("Returned json data: \n%s", body)
}

func parseLinks(links []string) map[string]string {
	aBrkt := regexp.MustCompile("[<>]")
	slver := regexp.MustCompile("(.+) *= *\"(.+)\"")
	linkMap := make(map[string]string)

	for _, link := range links {

		link = aBrkt.ReplaceAllString(link, "")
		parts := strings.Split(link, ";")

		matches := slver.FindStringSubmatch(parts[1])
		if len(matches) > 0 {
			linkMap[matches[2]] = parts[0]
		}
	}

	return linkMap
}
