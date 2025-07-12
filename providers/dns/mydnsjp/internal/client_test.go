package internal

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/stubrouter"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *stubrouter.Builder[*Client] {
	return stubrouter.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client := NewClient("xxx", "secret")
			client.HTTPClient = server.Client()
			client.baseURL, _ = url.Parse(server.URL)

			return client, nil
		},
		stubrouter.CheckHeader().
			WithContentTypeFromURLEncoded().
			WithBasicAuth("xxx", "secret"))
}

func TestClient_AddTXTRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /", nil,
			stubrouter.CheckForm().Strict().
				With("CERTBOT_DOMAIN", "example.com").
				With("CERTBOT_VALIDATION", "txt").
				With("EDIT_CMD", "REGIST")).
		Build(t)

	err := client.AddTXTRecord(t.Context(), "example.com", "txt")
	require.NoError(t, err)
}

func TestClient_DeleteTXTRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /", nil,
			stubrouter.CheckForm().Strict().
				With("CERTBOT_DOMAIN", "example.com").
				With("CERTBOT_VALIDATION", "txt").
				With("EDIT_CMD", "DELETE")).
		Build(t)

	err := client.DeleteTXTRecord(t.Context(), "example.com", "txt")
	require.NoError(t, err)
}
