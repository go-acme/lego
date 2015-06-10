package acme

import (
	"bytes"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/square/go-jose"
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

type challengeHandler interface {
	CanSolve() bool
	Solve()
}

// Client is the user-friendy way to ACME
type Client struct {
	regURL string
	user   User
}

// NewClient creates a new client for the set user.
func NewClient(caURL string, usr User) *Client {
	if err := usr.GetPrivateKey().Validate(); err != nil {
		logger().Fatalf("Could not validate the private account key of %s -> %v", usr.GetEmail(), err)
	}

	return &Client{regURL: caURL, user: usr}
}

// Posts a JWS signed message to the specified URL
func (c *Client) jwsPost(url string, content []byte) (*http.Response, error) {
	signer, err := jose.NewSigner(jose.RS256, c.user.GetPrivateKey())
	if err != nil {
		return nil, err
	}

	signed, err := signer.Sign(content)
	if err != nil {
		return nil, err
	}
	signedContent := signed.FullSerialize()

	resp, err := http.Post(url, "application/json", bytes.NewBuffer([]byte(signedContent)))
	if err != nil {
		return nil, err
	}

	return resp, err
}

// Register the current account to the ACME server.
func (c *Client) Register() (*RegistrationResource, error) {
	logger().Print("Registering account ... ")
	jsonBytes, err := json.Marshal(registrationMessage{Contact: []string{"mailto:" + c.user.GetEmail()}})
	if err != nil {
		return nil, err
	}

	resp, err := c.jwsPost(c.regURL, jsonBytes)
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

	logger().Printf("Agreement: %s", string(jsonBytes))

	resp, err := c.jwsPost(c.user.GetRegistration().URI, jsonBytes)
	if err != nil {
		return err
	}

	logResponseBody(resp)

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("The server returned %d but we expected %d", resp.StatusCode, http.StatusAccepted)
	}

	logResponseHeaders(resp)
	logResponseBody(resp)

	return nil
}

// ObtainCertificates tries to obtain certificates from the CA server
// using the challenges it has configured. It also tries to do multiple
// certificate processings at the same time in parallel.
func (c *Client) ObtainCertificates(domains []string) error {

	challenges := c.getChallenges(domains)
	c.doChallenges(challenges)
	return nil
}

func (c *Client) doChallenges(challenges []*authorizationResource) {
	for _, auth := range challenges {
	}
}

func (c *Client) getChallenges(domains []string) []*authorizationResource {
	resc, errc := make(chan *authorizationResource), make(chan error)

	for _, domain := range domains {
		go func(domain string) {
			jsonBytes, err := json.Marshal(authorization{Identifier: identifier{Type: "dns", Value: domain}})
			if err != nil {
				errc <- err
				return
			}

			resp, err := c.jwsPost(c.user.GetRegistration().NewAuthzURL, jsonBytes)
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

			resc <- &authorizationResource{Body: authz, NewCertURL: links["next"], Domain: domain}

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
