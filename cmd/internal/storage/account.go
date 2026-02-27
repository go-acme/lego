package storage

import (
	"crypto"

	"github.com/go-acme/lego/v5/acme"
	"github.com/go-acme/lego/v5/certcrypto"
)

const AccountIDPlaceholder = "noemail@example.com"

// Account represents a users local saved credentials.
type Account struct {
	ID           string                `json:"id"`
	Email        string                `json:"email"`
	KeyType      certcrypto.KeyType    `json:"keyType"`
	Registration *acme.ExtendedAccount `json:"registration"`

	key crypto.PrivateKey
}

func NewAccount(email, id string, keyType certcrypto.KeyType, key crypto.PrivateKey) *Account {
	return &Account{Email: email, ID: id, KeyType: keyType, key: key}
}

/** Implementation of the registration.User interface **/

// GetID returns the effective account ID.
func (a *Account) GetID() string {
	return getEffectiveAccountID(a.Email, a.ID)
}

// GetEmail returns the email address for the account.
func (a *Account) GetEmail() string {
	return a.Email
}

// GetKeyType returns the key type of the account.
func (a *Account) GetKeyType() certcrypto.KeyType {
	return a.KeyType
}

// GetPrivateKey returns the private account key.
func (a *Account) GetPrivateKey() crypto.PrivateKey {
	return a.key
}

// GetRegistration returns the server registration.
func (a *Account) GetRegistration() *acme.ExtendedAccount {
	return a.Registration
}

func getEffectiveAccountID(email, id string) string {
	if id != "" {
		return id
	}

	if email != "" {
		return email
	}

	return AccountIDPlaceholder
}
