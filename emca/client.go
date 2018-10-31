package emca

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/xenolf/lego/log"

	"github.com/xenolf/lego/emca/certificate"

	"github.com/xenolf/lego/emca/challenge"
	"github.com/xenolf/lego/emca/challenge/http01"
	"github.com/xenolf/lego/emca/challenge/tlsalpn01"
	"github.com/xenolf/lego/emca/internal/secure"
	"github.com/xenolf/lego/emca/internal/sender"
	"github.com/xenolf/lego/emca/le"
)

const (
	statusValid   = "valid"
	statusInvalid = "invalid"
)

const (
	// maxBodySize is the maximum size of body that we will read.
	maxBodySize = 1024 * 1024

	// overallRequestLimit is the overall number of request per second
	// limited on the "new-reg", "new-authz" and "new-cert" endpoints.
	// From the documentation the limitation is 20 requests per second,
	// but using 20 as value doesn't work but 18 do
	overallRequestLimit = 18
)

// Client is the user-friendly way to ACME
type Client struct {
	directory le.Directory
	user      User
	jws       *secure.JWS
	keyType   certificate.KeyType
	solvers   map[challenge.Type]solver

	do *sender.Do
}

// NewClient creates a new ACME client on behalf of the user.
// The client will depend on the ACME directory located at caDirURL for the rest of its actions.
// A private key of type keyType (see KeyType constants) will be generated when requesting a new certificate if one isn't provided.
func NewClient(config *Config) (*Client, error) {
	if config == nil {
		return nil, errors.New("a configuration must be provided")
	}

	privKey := config.user.GetPrivateKey()
	if privKey == nil {
		return nil, errors.New("private key was nil")
	}

	do := sender.NewDo(config.HTTPClient, config.userAgent)

	var dir le.Directory
	if _, err := do.Get(config.caDirURL, &dir); err != nil {
		return nil, fmt.Errorf("get directory at '%s': %v", config.caDirURL, err)
	}

	if dir.NewAccountURL == "" {
		return nil, errors.New("directory missing new registration URL")
	}
	if dir.NewOrderURL == "" {
		return nil, errors.New("directory missing new order URL")
	}

	jws := secure.NewJWS(do, privKey, dir.NewNonceURL)
	if reg := config.user.GetRegistration(); reg != nil {
		jws.SetKid(reg.URI)
	}

	// REVIEW: best possibility?
	// Add all available solvers with the right index as per ACME spec to this map.
	// Otherwise they won't be found.
	solvers := map[challenge.Type]solver{
		challenge.HTTP01:    http01.NewChallenge(jws, validate(do), &http01.ProviderServer{}),
		challenge.TLSALPN01: tlsalpn01.NewChallenge(jws, validate(do), &tlsalpn01.ProviderServer{}),
	}

	return &Client{
		directory: dir,
		user:      config.user,
		jws:       jws,
		keyType:   config.keyType,
		solvers:   solvers,
		do:        do,
	}, nil
}

// GetToSURL returns the current ToS URL from the Directory
func (c *Client) GetToSURL() string {
	return c.directory.Meta.TermsOfService
}

// GetExternalAccountRequired returns the External Account Binding requirement of the Directory
func (c *Client) GetExternalAccountRequired() bool {
	return c.directory.Meta.ExternalAccountRequired
}

func validate(do *sender.Do) func(*secure.JWS, string, string, le.Challenge) error {
	return func(j *secure.JWS, domain, uri string, _ le.Challenge) error {
		var chlng le.Challenge

		// Challenge initiation is done by sending a JWS payload containing the
		// trivial JSON object `{}`. We use an empty struct instance as the postJSON
		// payload here to achieve this result.
		hdr, err := j.PostJSON(uri, struct{}{}, &chlng)
		if err != nil {
			return err
		}

		// After the path is sent, the ACME server will access our server.
		// Repeatedly check the server for an updated status on our request.
		for {
			switch chlng.Status {
			case statusValid:
				log.Infof("[%s] The server validated our request", domain)
				return nil
			case "pending":
			case "processing":
			case statusInvalid:
				return handleChallengeError(chlng)
			default:
				return errors.New("the server returned an unexpected state")
			}

			ra, err := strconv.Atoi(hdr.Get("Retry-After"))
			if err != nil {
				// The ACME server MUST return a Retry-After.
				// If it doesn't, we'll just poll hard.
				ra = 5
			}

			time.Sleep(time.Duration(ra) * time.Second)

			resp, err := do.Get(uri, &chlng)
			if resp != nil {
				hdr = resp.Header
			}
			if err != nil {
				return err
			}

		}
	}
}

func handleChallengeError(chlng le.Challenge) error {
	return chlng.Error
}
