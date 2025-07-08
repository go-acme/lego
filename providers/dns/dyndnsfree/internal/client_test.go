package internal

import (
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/clientmock"
	"github.com/stretchr/testify/require"
)

func setupClient(server *httptest.Server) (*Client, error) {
	client, err := NewClient("user", "secret")
	if err != nil {
		return nil, err
	}

	client.baseURL = server.URL
	client.HTTPClient = server.Client()

	return client, nil
}

func TestAddTXTRecord(t *testing.T) {
	client := clientmock.NewBuilder[*Client](setupClient).
		Route("GET /", clientmock.RawStringResponse("success"),
			clientmock.CheckQueryParameter().Strict().
				With("add_hostname", "sub.example.com").
				With("hostname", "example.com").
				With("password", "secret").
				With("txt", "value").
				With("username", "user")).
		Build(t)

	err := client.AddTXTRecord(t.Context(), "example.com", "sub.example.com", "value")
	require.NoError(t, err)
}

func TestAddTXTRecord_error(t *testing.T) {
	client := clientmock.NewBuilder[*Client](setupClient).
		Route("GET /", clientmock.RawStringResponse("error: authentification failed")).
		Build(t)

	err := client.AddTXTRecord(t.Context(), "example.com", "sub.example.com", "value")
	require.EqualError(t, err, "error: authentification failed")
}
