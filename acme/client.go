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
	"strings"
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

type solver interface {
	CanSolve() bool
	Solve(challenge challenge, domain string) error
}

// Client is the user-friendy way to ACME
type Client struct {
	regURL  string
	user    User
	jws     *jws
	keyBits int
	Solvers map[string]solver
}

// NewClient creates a new client for the set user.
func NewClient(caURL string, usr User, keyBits int, optPort string) *Client {
	if err := usr.GetPrivateKey().Validate(); err != nil {
		logger().Fatalf("Could not validate the private account key of %s -> %v", usr.GetEmail(), err)
	}

	jws := &jws{privKey: usr.GetPrivateKey()}

	// REVIEW: best possibility?
	solvers := make(map[string]solver)
	solvers["simpleHttps"] = &simpleHTTPChallenge{jws: jws, optPort: optPort}

	return &Client{regURL: caURL, user: usr, jws: jws, keyBits: keyBits, Solvers: solvers}
}

// Register the current account to the ACME server.
func (c *Client) Register() (*RegistrationResource, error) {
	logger().Print("Registering account ... ")
	jsonBytes, err := json.Marshal(registrationMessage{Contact: []string{"mailto:" + c.user.GetEmail()}})
	if err != nil {
		return nil, err
	}

	resp, err := c.jws.post(c.regURL, jsonBytes)
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
// DER encoded byte slices.
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

// Looks through the challenge combinations to find a solvable match.
// Then solves the challenges in series and returns.
func (c *Client) solveChallenges(challenges []*authorizationResource) error {
	// loop through the resources, basically through the domains.
	for _, authz := range challenges {
		// no solvers - no solving
		if solvers := c.chooseSolvers(authz.Body); solvers != nil {
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
func (c *Client) chooseSolvers(auth authorization) map[int]solver {
	for _, combination := range auth.Combinations {
		solvers := make(map[int]solver)
		for _, idx := range combination {
			if solver, ok := c.Solvers[auth.Challenges[idx].Type]; ok {
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
			jsonBytes, err := json.Marshal(authorization{Identifier: identifier{Type: "dns", Value: domain}})
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

func (c *Client) requestCertificates(challenges []*authorizationResource) ([]CertificateResource, error) {
	var certs []CertificateResource
	for _, authz := range challenges {
		privKey, err := generatePrivateKey(c.keyBits)
		if err != nil {
			return nil, err
		}

		csr, err := generateCsr(privKey, authz.Domain)
		if err != nil {
			return nil, err
		}
		csrString := base64.URLEncoding.EncodeToString(csr)
		jsonBytes, err := json.Marshal(csrMessage{Csr: csrString, Authorizations: []string{authz.AuthURL}})
		if err != nil {
			return nil, err
		}

		resp, err := c.jws.post(authz.NewCertURL, jsonBytes)
		if err != nil {
			return nil, err
		}

		logResponseHeaders(resp)

		if resp.Header.Get("Content-Type") != "application/pkix-cert" {
			return nil, fmt.Errorf("The server returned an unexpected content-type header: %s - expected %s", resp.Header.Get("Content-Type"), "application/pkix-cert")
		}

		cert, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		privateKeyPem := pemEncode(privKey)

		certs = append(certs, CertificateResource{Domain: authz.Domain, CertURL: resp.Header.Get("Location"), PrivateKey: privateKeyPem, Certificate: cert})
	}
	return certs, nil
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
