package internal

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/clientmock"
	"github.com/stretchr/testify/require"
)

func setupClient(server *httptest.Server) (*Client, error) {
	client := NewClient("user", "secret")
	client.baseURL = server.URL
	client.HTTPClient = server.Client()

	return client, nil
}

func TestClient_UpdateTXTRecord(t *testing.T) {
	client := clientmock.NewBuilder[*Client](setupClient).
		Route("GET /", nil, clientmock.CheckQueryParameter().Strict().
			With("rid", "123456").
			With("content", "txt").
			With("username", "user").
			With("password", "secret"),
		).
		Build(t)

	err := client.UpdateTXTRecord(t.Context(), "123456", "txt")
	require.NoError(t, err)
}

func TestClient_UpdateTXTRecord_error(t *testing.T) {
	client := clientmock.NewBuilder[*Client](setupClient).
		Route("GET /", clientmock.Noop().WithStatusCode(http.StatusBadRequest)).
		Build(t)

	err := client.UpdateTXTRecord(t.Context(), "123456", "txt")
	require.EqualError(t, err, "unexpected status code: [status code: 400] body: ")
}
