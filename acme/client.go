package acme

import (
	"crypto/rsa"
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
	CanSolve(domain string) bool
	Solve(challenge challenge, domain string) error
}

// Client is the user-friendy way to ACME
type Client struct {
	directory directory
	user      User
	jws       *jws
	keyBits   int
	devMode   bool
	solvers   map[string]solver
}

// NewClient creates a new client for the set user.
// caURL - The root url to the boulder instance you want certificates from
// usr - A filled in user struct
// optPort - The alternative port to listen on for challenges.
// devMode - If set to true, all CanSolve() checks are skipped.
func NewClient(caURL string, usr User, keyBits int, optPort string, devMode bool) *Client {
	if err := usr.GetPrivateKey().Validate(); err != nil {
		logger().Fatalf("Could not validate the private account key of %s\n\t%v", usr.GetEmail(), err)
	}
	jws := &jws{privKey: usr.GetPrivateKey()}

	// REVIEW: best possibility?
	// Add all available solvers with the right index as per ACME
	// spec to this map. Otherwise they won`t be found.
	solvers := make(map[string]solver)
	solvers["simpleHttp"] = &simpleHTTPChallenge{jws: jws, optPort: optPort}

	dirResp, err := http.Get(caURL + "/directory")
	if err != nil {
		logger().Fatalf("Could not get directory from CA URL. Please check the URL.\n\t%v", err)
	}
	var dir directory
	decoder := json.NewDecoder(dirResp.Body)
	err = decoder.Decode(&dir)
	if err != nil {
		logger().Fatalf("Could not parse directory response from CA URL.\n\t%v", err)
	}
	if dir.NewRegURL == "" || dir.NewAuthzURL == "" || dir.NewCertURL == "" || dir.RevokeCertURL == "" {
		logger().Fatal("The directory returned by the server was invalid.")
	}

	return &Client{directory: dir, user: usr, jws: jws, keyBits: keyBits, devMode: devMode, solvers: solvers}
}

// Register the current account to the ACME server.
func (c *Client) Register() (*RegistrationResource, error) {
	logger().Print("Registering account ... ")
	jsonBytes, err := json.Marshal(registrationMessage{Resource: "new-reg", Contact: []string{"mailto:" + c.user.GetEmail()}})
	if err != nil {
		return nil, err
	}

	resp, err := c.jws.post(c.directory.NewRegURL, jsonBytes)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusConflict {
		// REVIEW: should this return an error?
		return nil, errors.New("This account is already registered with this CA.")
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

// AgreeToTos updates the Client registration and sends the agreement to
// the server.
func (c *Client) AgreeToTos() error {
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

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("The server returned %d but we expected %d", resp.StatusCode, http.StatusAccepted)
	}

	return nil
}

// ObtainCertificates tries to obtain certificates from the CA server
// using the challenges it has configured. The returned certificates are
// PEM encoded byte slices.
func (c *Client) ObtainCertificates(domains []string) ([]CertificateResource, error) {
	logger().Print("Obtaining certificates...")
	challenges := c.getChallenges(domains)
	err := c.solveChallenges(challenges)
	if err != nil {
		return nil, err
	}

	logger().Print("Validations succeeded. Getting certificates")

	return c.requestCertificates(challenges)
}

// RevokeCertificate takes a PEM encoded certificate and tries to revoke it at the CA.
func (c *Client) RevokeCertificate(certificate []byte) error {
	encodedCert := base64.URLEncoding.EncodeToString(certificate)

	jsonBytes, err := json.Marshal(revokeCertMessage{Resource: "revoke-cert", Certificate: encodedCert})
	if err != nil {
		return err
	}

	resp, err := c.jws.post(c.directory.RevokeCertURL, jsonBytes)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("The server returned an error while trying to revoke the certificate.\n%s", body)
	}

	return nil
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
			if solver, ok := c.solvers[auth.Challenges[idx].Type]; ok && (c.devMode || solver.CanSolve(domain)) {
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
				errc <- fmt.Errorf("Getting challenges for %s failed. Got status %d but expected %d",
					domain, resp.StatusCode, http.StatusCreated)
			}

			links := parseLinks(resp.Header["Link"])
			if links["next"] == "" {
				logger().Fatalln("The server did not provide enough information to proceed.")
			}

			var authz authorization
			decoder := json.NewDecoder(resp.Body)
			err = decoder.Decode(&authz)
			if err != nil {
				errc <- err
			}

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
func (c *Client) requestCertificates(challenges []*authorizationResource) ([]CertificateResource, error) {
	resc, errc := make(chan CertificateResource), make(chan error)
	for _, authz := range challenges {
		go c.requestCertificate(authz, resc, errc)
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

	return certs, nil
}

func (c *Client) requestCertificate(authz *authorizationResource, result chan CertificateResource, errc chan error) {
	privKey, err := generatePrivateKey(c.keyBits)
	if err != nil {
		errc <- err
		return
	}

	csr, err := generateCsr(privKey, authz.Domain)
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
			if err != nil {
				errc <- err
				return
			}

			// The server returns a body with a length of zero if the
			// certificate was not ready at the time this request completed.
			// Otherwise the body is the certificate.
			if len(cert) > 0 {
				cerRes.CertStableURL = resp.Header.Get("Content-Location")
				cerRes.Certificate = pemEncode(cert)
				result <- cerRes
			} else {
				// The certificate was granted but is not yet issued.
				// Check retry-after and loop.
				ra := resp.Header.Get("Retry-After")
				retryAfter, err := strconv.Atoi(ra)
				if err != nil {
					errc <- err
					return
				}

				logger().Printf("[%s] Server responded with status 202. Respecting retry-after of: %d", authz.Domain, retryAfter)
				time.Sleep(time.Duration(retryAfter) * time.Millisecond)
			}
			break
		default:
			logger().Fatalf("[%s] The server returned an unexpected status code %d.", authz.Domain, resp.StatusCode)
			return
		}

		resp, err = http.Get(cerRes.CertURL)
		if err != nil {
			errc <- err
			return
		}
	}
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
