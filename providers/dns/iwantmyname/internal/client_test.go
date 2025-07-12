package internal

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/stubrouter"
	"github.com/stretchr/testify/require"
)

func setupClient(server *httptest.Server) (*Client, error) {
	client := NewClient("user", "secret")
	client.HTTPClient = server.Client()
	client.baseURL, _ = url.Parse(server.URL)

	return client, nil
}

func TestClient_Do(t *testing.T) {
	client := stubrouter.NewBuilder[*Client](setupClient,
		stubrouter.CheckHeader().
			WithBasicAuth("user", "secret"),
	).
		Route("POST /", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			fmt.Println(req)
		}),
			stubrouter.CheckQueryParameter().Strict().
				With("hostname", "example.com").
				With("ttl", "120").
				With("type", "TXT").
				With("value", "data")).
		Build(t)

	record := Record{
		Hostname: "example.com",
		Type:     "TXT",
		Value:    "data",
		TTL:      120,
	}

	err := client.SendRequest(t.Context(), record)
	require.NoError(t, err)
}
