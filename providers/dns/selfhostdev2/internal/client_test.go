package internal

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/require"
)

func setupClient(server *httptest.Server) (*Client, error) {
	client := NewClient()
	client.BaseURL, _ = url.Parse(server.URL)
	client.HTTPClient = server.Client()

	return client, nil
}

func TestClient_UpdateTXTRecord(t *testing.T) {
	client := servermock.NewBuilder[*Client](setupClient).
		Route("POST /",
			servermock.ResponseFromFixture("present.json"),
		).
		Build(t)

	payload := Payload{
		APIKey:   "42.base64",
		Action:   "present",
		RecordID: 123456,
		Content:  "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
	}

	err := client.UpdateTXTRecord(t.Context(), payload)
	require.NoError(t, err)
}

func TestClient_UpdateTXTRecord_error(t *testing.T) {
	client := servermock.NewBuilder[*Client](setupClient).
		Route("POST /",
			servermock.Noop().WithStatusCode(http.StatusBadRequest),
		).
		Build(t)

	payload := Payload{
		APIKey:   "42.base64",
		Action:   "present",
		RecordID: 123456,
		Content:  "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
	}

	err := client.UpdateTXTRecord(t.Context(), payload)
	require.EqualError(t, err, "unexpected status code: [status code: 400] body: ")
}
