package registration

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"

	"github.com/xenolf/lego/emca/internal/secure"
	"github.com/xenolf/lego/emca/le"
	"github.com/xenolf/lego/log"
)

type Registrar struct {
	jws       *secure.JWS
	user      User
	directory le.Directory
}

func NewRegistrar(jws *secure.JWS, user User, directory le.Directory) *Registrar {
	return &Registrar{
		jws:       jws,
		user:      user,
		directory: directory,
	}
}

// Register the current account to the ACME server.
func (r *Registrar) Register(tosAgreed bool) (*le.RegistrationResource, error) {
	return r.register(tosAgreed)
}

// RegisterWithExternalAccountBinding Register the current account to the ACME server.
func (r *Registrar) RegisterWithExternalAccountBinding(tosAgreed bool, kid string, hmacEncoded string) (*le.RegistrationResource, error) {
	eab := func(accMsg *le.AccountMessage) error {
		hmac, err := base64.RawURLEncoding.DecodeString(hmacEncoded)
		if err != nil {
			return fmt.Errorf("acme: could not decode hmac key: %v", err)
		}

		eabJWS, err := r.jws.SignEABContent(r.directory.NewAccountURL, kid, hmac)
		if err != nil {
			return fmt.Errorf("acme: error signing eab content: %v", err)
		}

		accMsg.ExternalAccountBinding = []byte(eabJWS.FullSerialize())

		return nil
	}

	return r.register(tosAgreed, eab)
}

// register the current account to the ACME server.
func (r *Registrar) register(tosAgreed bool, opts ...func(*le.AccountMessage) error) (*le.RegistrationResource, error) {
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
	hdr, err := r.jws.PostJSON(r.directory.NewAccountURL, accMsg, &serverReg)
	if err != nil {
		errorDetails, ok := err.(le.ErrorDetails)
		if !ok || errorDetails.HTTPStatus != http.StatusConflict {
			return nil, err
		}
	}

	reg := &le.RegistrationResource{URI: hdr.Get("Location"), Body: serverReg}

	r.jws.SetKid(reg.URI)

	return reg, nil
}

// QueryRegistration runs a POST request on the client's registration and returns the result.
//
// This is similar to the Register function,
// but acting on an existing registration link and resource.
func (r *Registrar) QueryRegistration() (*le.RegistrationResource, error) {
	if r == nil || r.user == nil {
		return nil, errors.New("acme: cannot query the registration of a nil client or user")
	}

	// Log the URL here instead of the email as the email may not be set
	log.Infof("acme: Querying account for %s", r.user.GetRegistration().URI)

	accMsg := le.AccountMessage{}

	var serverReg le.AccountMessage
	_, err := r.jws.PostJSON(r.user.GetRegistration().URI, accMsg, &serverReg)
	if err != nil {
		return nil, err
	}

	return &le.RegistrationResource{
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

	accMsg := le.AccountMessage{Status: "deactivated"}

	_, err := r.jws.PostJSON(r.user.GetRegistration().URI, accMsg, nil)
	return err
}

// ResolveAccountByKey will attempt to look up an account using the given account key
// and return its registration resource.
func (r *Registrar) ResolveAccountByKey() (*le.RegistrationResource, error) {
	log.Infof("acme: Trying to resolve account by key")

	acc := le.AccountMessage{OnlyReturnExisting: true}
	hdr, err := r.jws.PostJSON(r.directory.NewAccountURL, acc, nil)
	if err != nil {
		return nil, err
	}

	accountLink := hdr.Get("Location")
	if accountLink == "" {
		return nil, errors.New("server did not return the account link")
	}

	r.jws.SetKid(accountLink)

	var retAccount le.AccountMessage
	_, err = r.jws.PostJSON(accountLink, le.AccountMessage{}, &retAccount)
	if err != nil {
		return nil, err
	}

	return &le.RegistrationResource{URI: accountLink, Body: retAccount}, nil
}
