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
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Logger is an optional custom logger.
var Logger *log.Logger

// logf writes a log entry. It uses Logger if not
// nil, otherwise it uses the default log.Logger.
func logf(format string, args ...interface{}) {
	if Logger != nil {
		Logger.Printf(format, args...)
	} else {
		log.Printf(format, args...)
	}
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

// NewClient creates a new ACME client on behalf of user. The client will depend on
// the ACME directory located at caDirURL for the rest of its actions. It will
// generate private keys for certificates of size keyBits. And, if the challenge
// type requires it, the client will open a port at optPort to solve the challenge.
// If optPort is blank, the port required by the spec will be used, but you must
// forward the required port to optPort for the challenge to succeed.
func NewClient(caDirURL string, user User, keyBits int, optPort string) (*Client, error) {
	privKey := user.GetPrivateKey()
	if privKey == nil {
		return nil, errors.New("private key was nil")
	}

	if err := privKey.Validate(); err != nil {
		return nil, fmt.Errorf("invalid private key: %v", err)
	}

	dirResp, err := http.Get(caDirURL)
	if err != nil {
		return nil, fmt.Errorf("get directory at '%s': %v", caDirURL, err)
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

	jws := &jws{privKey: privKey, directoryURL: caDirURL}

	// REVIEW: best possibility?
	// Add all available solvers with the right index as per ACME
	// spec to this map. Otherwise they won`t be found.
	solvers := make(map[string]solver)
	solvers["http-01"] = &httpChallenge{jws: jws, optPort: optPort}
	solvers["tls-sni-01"] = &tlsSNIChallenge{jws: jws, optPort: optPort}

	return &Client{directory: dir, user: user, jws: jws, keyBits: keyBits, solvers: solvers}, nil
}

// Register the current account to the ACME server.
func (c *Client) Register() (*RegistrationResource, error) {
	if c == nil || c.user == nil {
		return nil, errors.New("acme: cannot register a nil client or user")
	}
	logf("[INFO] acme: Registering account for %s", c.user.GetEmail())

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
		return nil, errors.New("acme: The server did not return 'next' link to proceed")
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
func (c *Client) ObtainCertificates(domains []string, bundle bool) ([]CertificateResource, map[string]error) {
	if bundle {
		logf("[INFO] acme: Obtaining bundled certificates for %v", strings.Join(domains, ", "))
	} else {
		logf("[INFO] acme: Obtaining certificates for %v", strings.Join(domains, ", "))
	}

	challenges, failures := c.getChallenges(domains)
	if len(challenges) == 0 {
		return nil, failures
	}

	err := c.solveChallenges(challenges)
	for k, v := range err {
		failures[k] = v
	}

	if len(failures) == len(domains) {
		return nil, failures
	}

	// remove failed challenges from slice
	var succeededChallenges []authorizationResource
	for _, chln := range challenges {
		if failures[chln.Domain] == nil {
			succeededChallenges = append(succeededChallenges, chln)
		}
	}

	logf("[INFO] acme: Validations succeeded; requesting certificates")

	certs, err := c.requestCertificates(succeededChallenges, bundle)
	for k, v := range err {
		failures[k] = v
	}

	return certs, failures
}

// ObtainSANCertificate tries to obtain a single certificate using all domains passed into it.
// The first domain in domains is used for the CommonName field of the certificate, all other
// domains are added using the Subject Alternate Names extension.
// If bundle is true, the []byte contains both the issuer certificate and
// your issued certificate as a bundle.
func (c *Client) ObtainSANCertificate(domains []string, bundle bool) (CertificateResource, map[string]error) {
	if bundle {
		logf("[INFO] acme: Obtaining bundled SAN certificate for %v", strings.Join(domains, ", "))
	} else {
		logf("[INFO] acme: Obtaining SAN certificate for %v", strings.Join(domains, ", "))
	}

	challenges, failures := c.getChallenges(domains)
	// If any challenge fails - return. Do not generate partial SAN certificates.
	if len(failures) > 0 {
		return CertificateResource{}, failures
	}

	errs := c.solveChallenges(challenges)
	// If any challenge fails - return. Do not generate partial SAN certificates.
	if len(errs) > 0 {
		return CertificateResource{}, errs
	}

	logf("[INFO] acme: Validations succeeded; requesting certificates")

	cert, err := c.requestCertificate(challenges, bundle)
	if err != nil {
		for _, chln := range challenges {
			failures[chln.Domain] = err
		}
	}

	return cert, failures
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
	logf("[INFO] acme: [%s] Trying renewal with %d hours remaining", cert.Domain, int(timeLeft.Hours()))

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
		logf("[INFO] acme: [%s] Server responded with renewed certificate", cert.Domain)
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
				logf("[ERROR] acme: [%s] Could not bundle issuer certificate: %v", cert.Domain, err)
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

	newCerts, failures := c.ObtainCertificates([]string{cert.Domain}, bundle)
	if len(failures) > 0 {
		return CertificateResource{}, failures[cert.Domain]
	}

	if revokeOld {
		c.RevokeCertificate(cert.Certificate)
	}

	return newCerts[0], nil
}

// Looks through the challenge combinations to find a solvable match.
// Then solves the challenges in series and returns.
func (c *Client) solveChallenges(challenges []authorizationResource) map[string]error {
	// loop through the resources, basically through the domains.
	failures := make(map[string]error)
	for _, authz := range challenges {
		// no solvers - no solving
		if solvers := c.chooseSolvers(authz.Body, authz.Domain); solvers != nil {
			for i, solver := range solvers {
				// TODO: do not immediately fail if one domain fails to validate.
				err := solver.Solve(authz.Body.Challenges[i], authz.Domain)
				if err != nil {
					failures[authz.Domain] = err
				}
			}
		} else {
			failures[authz.Domain] = fmt.Errorf("acme: Could not determine solvers for %s", authz.Domain)
		}
	}

	return failures
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
				logf("[INFO] acme: Could not find solver for: %s", auth.Challenges[idx].Type)
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
func (c *Client) getChallenges(domains []string) ([]authorizationResource, map[string]error) {
	resc, errc := make(chan authorizationResource), make(chan domainError)

	for _, domain := range domains {
		go func(domain string) {
			jsonBytes, err := json.Marshal(authorization{Resource: "new-authz", Identifier: identifier{Type: "dns", Value: domain}})
			if err != nil {
				errc <- domainError{Domain: domain, Error: err}
				return
			}

			resp, err := c.jws.post(c.user.GetRegistration().NewAuthzURL, jsonBytes)
			if err != nil {
				errc <- domainError{Domain: domain, Error: err}
				return
			}

			if resp.StatusCode != http.StatusCreated {
				errc <- domainError{Domain: domain, Error: handleHTTPError(resp)}
			}

			links := parseLinks(resp.Header["Link"])
			if links["next"] == "" {
				logf("[ERROR] acme: Server did not provide next link to proceed")
				return
			}

			var authz authorization
			decoder := json.NewDecoder(resp.Body)
			err = decoder.Decode(&authz)
			if err != nil {
				errc <- domainError{Domain: domain, Error: err}
			}
			resp.Body.Close()

			resc <- authorizationResource{Body: authz, NewCertURL: links["next"], AuthURL: resp.Header.Get("Location"), Domain: domain}
		}(domain)
	}

	responses := make(map[string]authorizationResource)
	failures := make(map[string]error)
	for i := 0; i < len(domains); i++ {
		select {
		case res := <-resc:
			responses[res.Domain] = res
		case err := <-errc:
			failures[err.Domain] = err.Error
		}
	}

	challenges := make([]authorizationResource, 0, len(responses))
	for _, domain := range domains {
		if challenge, ok := responses[domain]; ok {
			challenges = append(challenges, challenge)
		}
	}

	close(resc)
	close(errc)

	return challenges, failures
}

// requestCertificates iterates all granted authorizations, creates RSA private keys and CSRs.
// It then uses these to request a certificate from the CA and returns the list of successfully
// granted certificates.
func (c *Client) requestCertificates(challenges []authorizationResource, bundle bool) ([]CertificateResource, map[string]error) {
	resc, errc := make(chan CertificateResource), make(chan domainError)
	for _, authz := range challenges {
		go func(authz authorizationResource, resc chan CertificateResource, errc chan domainError) {
			certRes, err := c.requestCertificate([]authorizationResource{authz}, bundle)
			if err != nil {
				errc <- domainError{Domain: authz.Domain, Error: err}
			} else {
				resc <- certRes
			}
		}(authz, resc, errc)
	}

	var certs []CertificateResource
	failures := make(map[string]error)
	for i := 0; i < len(challenges); i++ {
		select {
		case res := <-resc:
			certs = append(certs, res)
		case err := <-errc:
			failures[err.Domain] = err.Error
		}
	}

	close(resc)
	close(errc)

	return certs, failures
}

func (c *Client) requestCertificate(authz []authorizationResource, bundle bool) (CertificateResource, error) {
	if len(authz) == 0 {
		return CertificateResource{}, errors.New("Passed no authorizations to requestCertificate!")
	}

	commonName := authz[0]
	privKey, err := generatePrivateKey(rsakey, c.keyBits)
	if err != nil {
		return CertificateResource{}, err
	}

	var san []string
	var authURLs []string
	for _, auth := range authz[1:] {
		san = append(san, auth.Domain)
		authURLs = append(authURLs, auth.AuthURL)
	}

	// TODO: should the CSR be customizable?
	csr, err := generateCsr(privKey.(*rsa.PrivateKey), commonName.Domain, san)
	if err != nil {
		return CertificateResource{}, err
	}

	csrString := base64.URLEncoding.EncodeToString(csr)
	jsonBytes, err := json.Marshal(csrMessage{Resource: "new-cert", Csr: csrString, Authorizations: authURLs})
	if err != nil {
		return CertificateResource{}, err
	}

	resp, err := c.jws.post(commonName.NewCertURL, jsonBytes)
	if err != nil {
		return CertificateResource{}, err
	}

	privateKeyPem := pemEncode(privKey)
	cerRes := CertificateResource{
		Domain:     commonName.Domain,
		CertURL:    resp.Header.Get("Location"),
		PrivateKey: privateKeyPem}

	for {

		switch resp.StatusCode {
		case 202:
		case 201:

			cert, err := ioutil.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				return CertificateResource{}, err
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
						logf("[WARNING] acme: [%s] Could not bundle issuer certificate: %v", commonName.Domain, err)
					} else {
						// Success - append the issuer cert to the issued cert.
						issuerCert = pemEncode(derCertificateBytes(issuerCert))
						issuedCert = append(issuedCert, issuerCert...)
					}
				}

				cerRes.Certificate = issuedCert
				logf("[%s] Server responded with a certificate.", commonName.Domain)
				return cerRes, nil
			}

			// The certificate was granted but is not yet issued.
			// Check retry-after and loop.
			ra := resp.Header.Get("Retry-After")
			retryAfter, err := strconv.Atoi(ra)
			if err != nil {
				return CertificateResource{}, err
			}

			logf("[INFO] acme: [%s] Server responded with status 202; retrying after %ds", commonName.Domain, retryAfter)
			time.Sleep(time.Duration(retryAfter) * time.Second)

			break
		default:
			return CertificateResource{}, handleHTTPError(resp)
		}

		resp, err = http.Get(cerRes.CertURL)
		if err != nil {
			return CertificateResource{}, err
		}
	}
}

// getIssuerCertificate requests the issuer certificate and caches it for
// subsequent requests.
func (c *Client) getIssuerCertificate(url string) ([]byte, error) {
	logf("[INFO] acme: Requesting issuer cert from %s", url)
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
