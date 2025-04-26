package internal

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	identifier, err := NewIdentifier("tyo1")
	require.NoError(t, err)

	identifier.HTTPClient = server.Client()
	identifier.baseURL, _ = url.Parse(server.URL)

	mux.HandleFunc("/v2.0/tokens", writeFixtureHandler(http.MethodPost, "tokens_POST.json"))

	auth := Auth{
		TenantID: "487727e3921d44e3bfe7ebb337bf085e",
		PasswordCredentials: PasswordCredentials{
			Username: "ConoHa",
			Password: "paSSword123456#$%",
		},
	}

	token, err := identifier.GetToken(context.Background(), auth)
	require.NoError(t, err)

	expected := &IdentityResponse{Access: Access{Token: Token{ID: "sample00d88246078f2bexample788f7"}}}

	assert.Equal(t, expected, token)
}
