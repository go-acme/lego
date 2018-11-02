package registration

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"

	"github.com/xenolf/lego/le"
	"github.com/xenolf/lego/le/api"
	"github.com/xenolf/lego/log"
)

// Resource represents all important information about a registration
// of which the client needs to keep track itself.
type Resource struct {
	Body le.AccountMessage `json:"body,omitempty"`
	URI  string            `json:"uri,omitempty"`
}

type Registrar struct {
	core *api.Core
	user User
}

func NewRegistrar(core *api.Core, user User) *Registrar {
	return &Registrar{
		core: core,
		user: user,
	}
}

// Register the current account to the ACME server.
func (r *Registrar) Register(tosAgreed bool) (*Resource, error) {
	return r.register(tosAgreed)
}

// RegisterWithExternalAccountBinding Register the current account to the ACME server.
func (r *Registrar) RegisterWithExternalAccountBinding(tosAgreed bool, kid string, hmacEncoded string) (*Resource, error) {
	eab := func(accMsg *le.AccountMessage) error {
		hmac, err := base64.RawURLEncoding.DecodeString(hmacEncoded)
		if err != nil {
			return fmt.Errorf("acme: could not decode hmac key: %v", err)
		}

		eabJWS, err := r.core.SignEABContent(r.core.GetDirectory().NewAccountURL, kid, hmac)
		if err != nil {
			return fmt.Errorf("acme: error signing eab content: %v", err)
		}

		accMsg.ExternalAccountBinding = eabJWS

		return nil
	}

	return r.register(tosAgreed, eab)
}

// register the current account to the ACME server.
func (r *Registrar) register(tosAgreed bool, opts ...func(*le.AccountMessage) error) (*Resource, error) {
	if r == nil || r.user == nil {
		return nil, errors.New("acme: cannot register a nil client or user")
	}

	log.Infof("acme: Registering account for %s", r.user.GetEmail())

	accMsg := &le.AccountMessage{
		TermsOfServiceAgreed: tosAgreed,
		Contact:              []string{},
	}

	if r.user.GetEmail() != "" {
		accMsg.Contact = []string{"mailto:" + r.user.GetEmail()}
	}

	for _, opt := range opts {
		err := opt(accMsg)
		if err != nil {
			return nil, err
		}
	}

	var serverReg le.AccountMessage
	resp, err := r.core.Post(r.core.GetDirectory().NewAccountURL, accMsg, &serverReg)
	if err != nil {
		errorDetails, ok := err.(le.ProblemDetails)
		if !ok || errorDetails.HTTPStatus != http.StatusConflict {
			return nil, err
		}
	}

	r.core.UpdateKID(resp.Header.Get("Location"))

	return &Resource{URI: resp.Header.Get("Location"), Body: serverReg}, nil
}

// QueryRegistration runs a POST request on the client's registration and returns the result.
//
// This is similar to the Register function,
// but acting on an existing registration link and resource.
func (r *Registrar) QueryRegistration() (*Resource, error) {
	if r == nil || r.user == nil {
		return nil, errors.New("acme: cannot query the registration of a nil client or user")
	}

	// Log the URL here instead of the email as the email may not be set
	log.Infof("acme: Querying account for %s", r.user.GetRegistration().URI)

	accMsg := le.AccountMessage{}

	var serverReg le.AccountMessage
	_, err := r.core.Post(r.user.GetRegistration().URI, accMsg, &serverReg)
	if err != nil {
		return nil, err
	}

	return &Resource{
		Body: serverReg,
		// Location: header is not returned so this needs to be populated off of existing URI
		URI: r.user.GetRegistration().URI,
	}, nil
}

// DeleteRegistration deletes the client's user registration from the ACME server.
func (r *Registrar) DeleteRegistration() error {
	if r == nil || r.user == nil {
		return errors.New("acme: cannot unregister a nil client or user")
	}

	log.Infof("acme: Deleting account for %s", r.user.GetEmail())

	accMsg := le.AccountMessage{Status: le.StatusDeactivated}

	_, err := r.core.Post(r.user.GetRegistration().URI, accMsg, nil)
	return err
}

// ResolveAccountByKey will attempt to look up an account using the given account key
// and return its registration resource.
func (r *Registrar) ResolveAccountByKey() (*Resource, error) {
	log.Infof("acme: Trying to resolve account by key")

	acc := le.AccountMessage{OnlyReturnExisting: true}
	resp, err := r.core.Post(r.core.GetDirectory().NewAccountURL, acc, nil)
	if err != nil {
		return nil, err
	}

	accountLink := resp.Header.Get("Location")
	if accountLink == "" {
		return nil, errors.New("server did not return the account link")
	}

	r.core.UpdateKID(accountLink)

	var retAccount le.AccountMessage
	_, err = r.core.Post(accountLink, le.AccountMessage{}, &retAccount)
	if err != nil {
		return nil, err
	}

	return &Resource{URI: accountLink, Body: retAccount}, nil
}
