package cmd

import (
	"crypto"

	"github.com/go-acme/lego/v4/registration"
)

// Account represents a users local saved credentials.
type Account struct {
	Email        string                 `json:"email"`
	Registration *registration.Resource `json:"registration"`
	key          crypto.PrivateKey
}

/** Implementation of the registration.User interface **/

// GetEmail returns the email address for the account.
func (a *Account) GetEmail() string {
	return a.Email
}

// GetPrivateKey returns the private RSA account key.
func (a *Account) GetPrivateKey() crypto.PrivateKey {
	return a.key
}

// GetRegistration returns the server registration.
func (a *Account) GetRegistration() *registration.Resource {
	return a.Registration
}

/** End **/
