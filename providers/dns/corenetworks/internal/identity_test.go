package internal

import (
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/clientmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_CreateAuthenticationToken(t *testing.T) {
	client := mockBuilder().
		Route("POST /auth/token", clientmock.ResponseFromFixture("auth.json")).
		Build(t)

	token, err := client.CreateAuthenticationToken(t.Context())
	require.NoError(t, err)

	expected := &Token{
		Token:   "authsecret",
		Expires: 123,
	}
	assert.Equal(t, expected, token)
}
