package storage

import (
	"crypto"
	"encoding/json"
	"fmt"
	"os"

	"github.com/go-acme/lego/v5/acme"
	"github.com/go-acme/lego/v5/certcrypto"
	"github.com/go-acme/lego/v5/cmd/internal/configuration"
)

const (
	OriginConfiguration = "configuration"
	OriginCommand       = "command"
	OriginMigration     = "migration"
)

// Account represents a users local saved credentials.
type Account struct {
	ID      string             `json:"id"`
	Email   string             `json:"email"`
	KeyType certcrypto.KeyType `json:"keyType"`
	Server  string             `json:"server"`

	Origin string `json:"origin,omitempty"`

	Registration  *acme.ExtendedAccount `json:"registration"`
	NeedsRecovery bool                  `json:"-"`

	key crypto.Signer
}

func NewRawAccount(id, email string, key crypto.Signer) (*Account, error) {
	keyType, err := certcrypto.GetKeyType(key)
	if err != nil {
		return nil, fmt.Errorf("get the key type: %w", err)
	}

	return &Account{
		ID:      GetEffectiveAccountID(email, id),
		Email:   email,
		KeyType: keyType,
		key:     key,
	}, nil
}

/** Implementation of the registration.User interface **/

// GetID returns the effective account ID.
func (a *Account) GetID() string {
	return GetEffectiveAccountID(a.Email, a.ID)
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
func (a *Account) GetPrivateKey() crypto.Signer {
	return a.key
}

// SetPrivateKey sets the private account key.
func (a *Account) SetPrivateKey(key crypto.Signer) {
	a.key = key
}

// GetRegistration returns the server registration.
func (a *Account) GetRegistration() *acme.ExtendedAccount {
	return a.Registration
}

// GetEffectiveAccountID returns the effective account ID.
func GetEffectiveAccountID(email, id string) string {
	if id != "" {
		return id
	}

	if email != "" {
		return email
	}

	return configuration.DefaultAccountID
}

func ReadJSONFile[T any](filename string) (*T, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	result := new(T)

	err = json.Unmarshal(data, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
