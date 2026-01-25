package zerossl

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v5/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client := NewClient()

			client.baseURL, _ = url.Parse(server.URL)
			client.httpClient = server.Client()

			return client, nil
		},
		servermock.CheckHeader().
			With("Accept", "application/json"),
	)
}

func TestClient_GenerateEAB(t *testing.T) {
	client := mockBuilder().
		Route("POST /acme/eab-credentials",
			servermock.ResponseFromFixture("success.json"),
			servermock.CheckQueryParameter().Strict().
				With("access_key", "secret"),
		).
		Build(t)

	eab, err := client.GenerateEAB(t.Context(), "secret")
	require.NoError(t, err)

	expected := &APIResponse{
		Success: true,
		Kid:     "GD-VvWydSVFuss_GhBwYQQ",
		HmacKey: "MjXU3MH-Z0WQ7piMAnVsCpD1shgMiWx6ggPWiTmydgUaj7dWWWfQfA",
	}

	assert.Equal(t, expected, eab)
}

func TestClient_GenerateEAB_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /acme/eab-credentials",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized),
		).
		Build(t)

	_, err := client.GenerateEAB(t.Context(), "foo")
	require.EqualError(t, err, "101: invalid_access_key: You have not supplied a valid API Access Key.")
}

func TestClient_GenerateEABFromEmail(t *testing.T) {
	client := mockBuilder().
		Route("POST /acme/eab-credentials-email",
			servermock.ResponseFromFixture("success.json"),
			servermock.CheckHeader().
				WithContentTypeFromURLEncoded(),
			servermock.CheckForm().
				With("email", "test@exmample.com"),
		).
		Build(t)

	eab, err := client.GenerateEABFromEmail(t.Context(), "test@exmample.com")
	require.NoError(t, err)

	expected := &APIResponse{
		Success: true,
		Kid:     "GD-VvWydSVFuss_GhBwYQQ",
		HmacKey: "MjXU3MH-Z0WQ7piMAnVsCpD1shgMiWx6ggPWiTmydgUaj7dWWWfQfA",
	}

	assert.Equal(t, expected, eab)
}

func TestClient_GenerateEABFromEmail_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /acme/eab-credentials-email",
			// NOTE: with this endpoint the server always returns a 200.
			servermock.ResponseFromFixture("error_email.json"),
		).
		Build(t)

	_, err := client.GenerateEABFromEmail(t.Context(), "test@exmample.com")
	require.EqualError(t, err, "2900: missing_email")
}
