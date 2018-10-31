package registration

import (
	"crypto"
	"crypto/rsa"

	"github.com/xenolf/lego/emca/le"
)

type mockUser struct {
	email      string
	regres     *le.RegistrationResource
	privatekey *rsa.PrivateKey
}

func (u mockUser) GetEmail() string                          { return u.email }
func (u mockUser) GetRegistration() *le.RegistrationResource { return u.regres }
func (u mockUser) GetPrivateKey() crypto.PrivateKey          { return u.privatekey }
