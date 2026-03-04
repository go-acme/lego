package internal

import (
	"crypto"

	"github.com/go-acme/lego/v5/acme"
)

type FakeUser struct {
	Email        string
	PrivateKey   crypto.Signer
	Registration *acme.ExtendedAccount
}

func (f *FakeUser) GetEmail() string                       { return f.Email }
func (f *FakeUser) GetRegistration() *acme.ExtendedAccount { return f.Registration }
func (f *FakeUser) GetPrivateKey() crypto.Signer           { return f.PrivateKey }
