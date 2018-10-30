package acme

import (
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/xenolf/lego/log"
)

// Register the current account to the ACME server.
func (c *Client) Register(tosAgreed bool) (*RegistrationResource, error) {
	if c == nil || c.user == nil {
		return nil, errors.New("acme: cannot register a nil client or user")
	}
	log.Infof("acme: Registering account for %s", c.user.GetEmail())

	accMsg := accountMessage{}
	if c.user.GetEmail() != "" {
		accMsg.Contact = []string{"mailto:" + c.user.GetEmail()}
	} else {
		accMsg.Contact = []string{}
	}
	accMsg.TermsOfServiceAgreed = tosAgreed

	var serverReg accountMessage
	hdr, err := c.jws.postJSON(c.directory.NewAccountURL, accMsg, &serverReg)
	if err != nil {
		remoteErr, ok := err.(RemoteError)
		if ok && remoteErr.StatusCode == 409 {
		} else {
			return nil, err
		}
	}

	reg := &RegistrationResource{
		URI:  hdr.Get("Location"),
		Body: serverReg,
	}
	c.jws.setKid(reg.URI)

	return reg, nil
}

// RegisterWithExternalAccountBinding Register the current account to the ACME server.
func (c *Client) RegisterWithExternalAccountBinding(tosAgreed bool, kid string, hmacEncoded string) (*RegistrationResource, error) {
	if c == nil || c.user == nil {
		return nil, errors.New("acme: cannot register a nil client or user")
	}
	log.Infof("acme: Registering account (EAB) for %s", c.user.GetEmail())

	accMsg := accountMessage{}
	if c.user.GetEmail() != "" {
		accMsg.Contact = []string{"mailto:" + c.user.GetEmail()}
	} else {
		accMsg.Contact = []string{}
	}

	accMsg.TermsOfServiceAgreed = tosAgreed

	hmac, err := base64.RawURLEncoding.DecodeString(hmacEncoded)
	if err != nil {
		return nil, fmt.Errorf("acme: could not decode hmac key: %v", err)
	}

	eabJWS, err := c.jws.signEABContent(c.directory.NewAccountURL, kid, hmac)
	if err != nil {
		return nil, fmt.Errorf("acme: error signing eab content: %v", err)
	}
	accMsg.ExternalAccountBinding = []byte(eabJWS.FullSerialize())

	var serverReg accountMessage
	hdr, err := c.jws.postJSON(c.directory.NewAccountURL, accMsg, &serverReg)
	if err != nil {
		remoteErr, ok := err.(RemoteError)
		if ok && remoteErr.StatusCode == 409 {
		} else {
			return nil, err
		}
	}

	reg := &RegistrationResource{URI: hdr.Get("Location"), Body: serverReg}
	c.jws.setKid(reg.URI)

	return reg, nil
}

// DeleteRegistration deletes the client's user registration from the ACME
// server.
func (c *Client) DeleteRegistration() error {
	if c == nil || c.user == nil {
		return errors.New("acme: cannot unregister a nil client or user")
	}
	log.Infof("acme: Deleting account for %s", c.user.GetEmail())

	accMsg := accountMessage{
		Status: "deactivated",
	}

	_, err := c.jws.postJSON(c.user.GetRegistration().URI, accMsg, nil)
	return err
}

// QueryRegistration runs a POST request on the client's registration and
// returns the result.
//
// This is similar to the Register function, but acting on an existing
// registration link and resource.
func (c *Client) QueryRegistration() (*RegistrationResource, error) {
	if c == nil || c.user == nil {
		return nil, errors.New("acme: cannot query the registration of a nil client or user")
	}
	// Log the URL here instead of the email as the email may not be set
	log.Infof("acme: Querying account for %s", c.user.GetRegistration().URI)

	accMsg := accountMessage{}

	var serverReg accountMessage
	_, err := c.jws.postJSON(c.user.GetRegistration().URI, accMsg, &serverReg)
	if err != nil {
		return nil, err
	}

	reg := &RegistrationResource{Body: serverReg}

	// Location: header is not returned so this needs to be populated off of
	// existing URI
	reg.URI = c.user.GetRegistration().URI

	return reg, nil
}

// ResolveAccountByKey will attempt to look up an account using the given account key
// and return its registration resource.
func (c *Client) ResolveAccountByKey() (*RegistrationResource, error) {
	log.Infof("acme: Trying to resolve account by key")

	acc := accountMessage{OnlyReturnExisting: true}
	hdr, err := c.jws.postJSON(c.directory.NewAccountURL, acc, nil)
	if err != nil {
		return nil, err
	}

	accountLink := hdr.Get("Location")
	if accountLink == "" {
		return nil, errors.New("Server did not return the account link")
	}

	var retAccount accountMessage
	c.jws.setKid(accountLink)
	_, err = c.jws.postJSON(accountLink, accountMessage{}, &retAccount)
	if err != nil {
		return nil, err
	}

	return &RegistrationResource{URI: accountLink, Body: retAccount}, nil
}
