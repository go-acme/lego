package acme

import (
	"bytes"
	"crypto/rsa"
	"encoding/json"
	"errors"
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
