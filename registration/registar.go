package registration

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/go-acme/lego/v4/acme"
	"github.com/go-acme/lego/v4/acme/api"
	"github.com/go-acme/lego/v4/log"
)

const mailTo = "mailto:"

// Resource represents all important information about a registration
// of which the client needs to keep track itself.
// WARNING: will be removed in the future (acme.ExtendedAccount), https://github.com/go-acme/lego/issues/855.
type Resource struct {
	Body acme.Account `json:"body,omitempty"`
	URI  string       `json:"uri,omitempty"`
}

type RegisterOptions struct {
	TermsOfServiceAgreed bool
}

type RegisterEABOptions struct {
	TermsOfServiceAgreed bool
	Kid                  string
	HmacEncoded          string
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
func (r *Registrar) Register(options RegisterOptions) (*Resource, error) {
	if r == nil || r.user == nil {
		return nil, errors.New("acme: cannot register a nil client or user")
	}

	accMsg := acme.Account{
		TermsOfServiceAgreed: options.TermsOfServiceAgreed,
		Contact:              []string{},
	}

	if r.user.GetEmail() != "" {
		log.Infof("acme: Registering account for %s", r.user.GetEmail())
		accMsg.Contact = []string{mailTo + r.user.GetEmail()}
	}

	account, err := r.core.Accounts.New(accMsg)
	if err != nil {
		// seems impossible
		var errorDetails acme.ProblemDetails
		if !errors.As(err, &errorDetails) || errorDetails.HTTPStatus != http.StatusConflict {
			return nil, err
		}
	}

	return &Resource{URI: account.Location, Body: account.Account}, nil
}

func createZeroSSLAccount(email string) (string, string, error) {
	newAccountURL := "https://api.zerossl.com/acme/eab-credentials-email"
	data := struct {
		Success bool   `json:"success"`
		KID     string `json:"eab_kid"`
		HMAC    string `json:"eab_hmac_key"`
	}{}

	resp, err := http.PostForm(newAccountURL, url.Values{"email": {email}})
	if err != nil {
		return "", "", fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	// ZeroSSL might return errors as plain-text messages instead of JSON,
	// so we buffer the response to be able to return it as error.
	var rawResp bytes.Buffer
	r := io.TeeReader(io.LimitReader(resp.Body, 10*1024), &rawResp) // Limit response to 10KB
	if err := json.NewDecoder(r).Decode(&data); err != nil {
		// It is likely not a JSON but a plain-text error message
		_, _ = io.ReadAll(r) // read the rest of the body
		return "", "", fmt.Errorf("parsing response: %w. Original response:\n%s", err, rawResp.String())
	}

	if !data.Success {
		return "", "", errors.New("received success=false")
	}
	return data.KID, data.HMAC, nil
}

// RegisterWithZeroSSL Register the current account to the ZeroSSL server.
func (r *Registrar) RegisterWithZeroSSL(options RegisterOptions) (*Resource, error) {
	if r.user.GetEmail() == "" {
		return nil, errors.New("acme: cannot register ZeroSSL account without email address")
	}

	kid, hmac, err := createZeroSSLAccount(r.user.GetEmail())
	if err != nil {
		return nil, fmt.Errorf("acme: error registering new ZeroSSL account: %w", err)
	}

	return r.RegisterWithExternalAccountBinding(RegisterEABOptions{
		TermsOfServiceAgreed: options.TermsOfServiceAgreed,
		Kid:                  kid,
		HmacEncoded:          hmac,
	})
}

// RegisterWithExternalAccountBinding Register the current account to the ACME server.
func (r *Registrar) RegisterWithExternalAccountBinding(options RegisterEABOptions) (*Resource, error) {
	accMsg := acme.Account{
		TermsOfServiceAgreed: options.TermsOfServiceAgreed,
		Contact:              []string{},
	}

	if r.user.GetEmail() != "" {
		log.Infof("acme: Registering account for %s", r.user.GetEmail())
		accMsg.Contact = []string{mailTo + r.user.GetEmail()}
	}

	account, err := r.core.Accounts.NewEAB(accMsg, options.Kid, options.HmacEncoded)
	if err != nil {
		// seems impossible
		var errorDetails acme.ProblemDetails
		if !errors.As(err, &errorDetails) || errorDetails.HTTPStatus != http.StatusConflict {
			return nil, err
		}
	}

	return &Resource{URI: account.Location, Body: account.Account}, nil
}

// QueryRegistration runs a POST request on the client's registration and returns the result.
//
// This is similar to the Register function,
// but acting on an existing registration link and resource.
func (r *Registrar) QueryRegistration() (*Resource, error) {
	if r == nil || r.user == nil || r.user.GetRegistration() == nil {
		return nil, errors.New("acme: cannot query the registration of a nil client or user")
	}

	// Log the URL here instead of the email as the email may not be set
	log.Infof("acme: Querying account for %s", r.user.GetRegistration().URI)

	account, err := r.core.Accounts.Get(r.user.GetRegistration().URI)
	if err != nil {
		return nil, err
	}

	return &Resource{
		Body: account,
		// Location: header is not returned so this needs to be populated off of existing URI
		URI: r.user.GetRegistration().URI,
	}, nil
}

// UpdateRegistration update the user registration on the ACME server.
func (r *Registrar) UpdateRegistration(options RegisterOptions) (*Resource, error) {
	if r == nil || r.user == nil {
		return nil, errors.New("acme: cannot update a nil client or user")
	}

	accMsg := acme.Account{
		TermsOfServiceAgreed: options.TermsOfServiceAgreed,
		Contact:              []string{},
	}

	if r.user.GetEmail() != "" {
		log.Infof("acme: Registering account for %s", r.user.GetEmail())
		accMsg.Contact = []string{mailTo + r.user.GetEmail()}
	}

	accountURL := r.user.GetRegistration().URI

	account, err := r.core.Accounts.Update(accountURL, accMsg)
	if err != nil {
		return nil, err
	}

	return &Resource{URI: accountURL, Body: account}, nil
}

// DeleteRegistration deletes the client's user registration from the ACME server.
func (r *Registrar) DeleteRegistration() error {
	if r == nil || r.user == nil {
		return errors.New("acme: cannot unregister a nil client or user")
	}

	log.Infof("acme: Deleting account for %s", r.user.GetEmail())

	return r.core.Accounts.Deactivate(r.user.GetRegistration().URI)
}

// ResolveAccountByKey will attempt to look up an account using the given account key
// and return its registration resource.
func (r *Registrar) ResolveAccountByKey() (*Resource, error) {
	log.Infof("acme: Trying to resolve account by key")

	accMsg := acme.Account{OnlyReturnExisting: true}
	account, err := r.core.Accounts.New(accMsg)
	if err != nil {
		return nil, err
	}

	return &Resource{URI: account.Location, Body: account.Account}, nil
}
