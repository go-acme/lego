package internal

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client, err := NewClient("secret")
			if err != nil {
				return nil, err
			}

			client.BaseURL, _ = url.Parse(server.URL)
			client.HTTPClient = server.Client()

			return client, nil
		},
		servermock.CheckHeader().
			WithJSONHeaders().
			With(authorizationTokenHeader, "secret"),
	)
}

func TestClient_AddTXTRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /DNS/example.com/TXT",
			servermock.Noop(),
			servermock.CheckRequestJSONBodyFromFixture("add_txt_record-request.json"),
		).
		Build(t)

	record := TXTRecord{
		Hostname:    "_acme-challenge",
		Text:        "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		TTL:         120,
		PublishZone: 1,
	}

	err := client.AddTXTRecord(t.Context(), "example.com", record)
	require.NoError(t, err)
}

func TestClient_DeleteTXTRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /DNS/example.com/TXT",
			servermock.Noop(),
			servermock.CheckRequestJSONBodyFromFixture("add_txt_record-request.json"),
		).
		Build(t)

	record := TXTRecord{
		Hostname:    "_acme-challenge",
		Text:        "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		TTL:         120,
		PublishZone: 1,
	}

	err := client.DeleteTXTRecord(t.Context(), "example.com", record)
	require.NoError(t, err)
}
