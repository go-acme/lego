package registration

import (
	"crypto"
	"crypto/rsa"

	"github.com/go-acme/lego/v5/acme"
)

type mockUser struct {
	email      string
	regres     *acme.ExtendedAccount
	privatekey *rsa.PrivateKey
}

func (u mockUser) GetEmail() string                       { return u.email }
func (u mockUser) GetRegistration() *acme.ExtendedAccount { return u.regres }
func (u mockUser) GetPrivateKey() crypto.PrivateKey       { return u.privatekey }
