package acme

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xenolf/lego/platform/tester"
	"github.com/xenolf/lego/registration"
)

func TestNewClient(t *testing.T) {
	_, apiURL, tearDown := tester.SetupFakeAPI()
	defer tearDown()

	keyBits := 32 // small value keeps test fast
	key, err := rsa.GenerateKey(rand.Reader, keyBits)
	require.NoError(t, err, "Could not generate test key")

	user := mockUser{
		email:      "test@test.com",
		regres:     new(registration.Resource),
		privatekey: key,
	}

	config := NewDefaultConfig(user).WithCADirURL(apiURL)

	client, err := NewClient(config)
	require.NoError(t, err, "Could not create client")

	assert.NotNil(t, client)
}

type mockUser struct {
	email      string
	regres     *registration.Resource
	privatekey *rsa.PrivateKey
}

func (u mockUser) GetEmail() string                        { return u.email }
func (u mockUser) GetRegistration() *registration.Resource { return u.regres }
func (u mockUser) GetPrivateKey() crypto.PrivateKey        { return u.privatekey }
