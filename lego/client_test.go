package lego

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/go-acme/lego/v3/platform/tester"
	"github.com/go-acme/lego/v3/registration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	config := NewConfig(user)
	config.CADirURL = apiURL + "/dir"

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
