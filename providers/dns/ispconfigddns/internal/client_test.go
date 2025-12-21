package internal

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/require"
)

func setupClient(server *httptest.Server) (*Client, error) {
	client, err := NewClient(server.URL, "secret")
	if err != nil {
		return nil, err
	}

	client.HTTPClient = server.Client()

	return client, nil
}

func TestClient_AddTXTRecord(t *testing.T) {
	client := servermock.NewBuilder[*Client](setupClient).
		Route("POST /ddns/update.php",
			servermock.Noop(),
			servermock.CheckHeader().
				WithBasicAuth("anonymous", "secret"),
			servermock.CheckQueryParameter().Strict().
				With("action", "add").
				With("zone", "example.com").
				With("type", "TXT").
				With("record", "_acme-challenge.example.com.").
				With("data", "token"),
		).
		Build(t)

	err := client.AddTXTRecord(t.Context(), "example.com", "_acme-challenge.example.com.", "token")
	require.NoError(t, err)
}

func TestClient_AddTXTRecord_error(t *testing.T) {
	client := servermock.NewBuilder[*Client](setupClient).
		Route("POST /ddns/update.php",
			servermock.RawStringResponse("Missing or invalid token.").
				WithStatusCode(http.StatusUnauthorized),
		).
		Build(t)

	err := client.AddTXTRecord(t.Context(), "example.com", "_acme-challenge.example.com.", "token")
	require.EqualError(t, err, "unexpected status code: [status code: 401] body: Missing or invalid token.")
}

func TestClient_DeleteTXTRecord(t *testing.T) {
	client := servermock.NewBuilder[*Client](setupClient).
		Route("DELETE /ddns/update.php",
			servermock.Noop(),
			servermock.CheckHeader().
				WithBasicAuth("anonymous", "secret"),
			servermock.CheckQueryParameter().Strict().
				With("action", "delete").
				With("zone", "example.com").
				With("type", "TXT").
				With("record", "_acme-challenge.example.com.").
				With("data", "token"),
		).
		Build(t)

	err := client.DeleteTXTRecord(t.Context(), "example.com", "_acme-challenge.example.com.", "token")
	require.NoError(t, err)
}

func TestClient_DeleteTXTRecord_error(t *testing.T) {
	client := servermock.NewBuilder[*Client](setupClient).
		Route("DELETE /ddns/update.php",
			servermock.RawStringResponse("Missing or invalid token.").
				WithStatusCode(http.StatusUnauthorized),
		).
		Build(t)

	err := client.DeleteTXTRecord(t.Context(), "example.com", "_acme-challenge.example.com.", "token")
	require.EqualError(t, err, "unexpected status code: [status code: 401] body: Missing or invalid token.")
}
