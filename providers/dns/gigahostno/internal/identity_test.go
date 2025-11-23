package internal

import (
	"context"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupIdentifierClient(server *httptest.Server) (*Identifier, error) {
	client, err := NewIdentifier("user", "secret", "")
	if err != nil {
		return nil, err
	}

	client.BaseURL, _ = url.Parse(server.URL)
	client.HTTPClient = server.Client()

	return client, nil
}

func mockContext(t *testing.T) context.Context {
	t.Helper()

	return context.WithValue(t.Context(), tokenKey, "secret")
}

func TestIdentifier_Authenticate(t *testing.T) {
	identifier := servermock.NewBuilder[*Identifier](setupIdentifierClient).
		Route("POST /authenticate",
			servermock.ResponseFromFixture("authenticate.json"),
			servermock.CheckRequestJSONBodyFromFixture("authenticate-request.json")).
		Build(t)

	token, err := identifier.Authenticate(context.Background())
	require.NoError(t, err)

	assert.Equal(t, "secrettoken", token)
}
