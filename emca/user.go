package emca

import (
	"crypto"

	"github.com/xenolf/lego/emca/le"
)

// User interface is to be implemented by users of this library.
// It is used by the client type to get user specific information.
type User interface {
	GetEmail() string
	GetRegistration() *le.RegistrationResource
	GetPrivateKey() crypto.PrivateKey
}
