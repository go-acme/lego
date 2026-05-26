package internal

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockBuilder(apiKey string) *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client := NewClient(apiKey)
			client.BaseURL, _ = url.Parse(server.URL)
			client.HTTPClient = server.Client()

			return client, nil
		},
		servermock.CheckHeader().
			WithJSONHeaders().
			WithAuthorization(apiKey),
	)
}

func TestClient_GetTxtRecord(t *testing.T) {
	client := mockBuilder("myKeyA").
		Route("GET /cdn/4.0/domains/example.com/dns-records",
			servermock.ResponseFromFixture("get_txt_record.json"),
			servermock.CheckQueryParameter().
				With("search", "acme-challenge"),
		).
		Build(t)

	_, err := client.GetTxtRecord(t.Context(), "example.com", "_acme-challenge", "txtxtxt")
	require.NoError(t, err)
}

func TestClient_CreateRecord(t *testing.T) {
	client := mockBuilder("myKeyB").
		Route("POST /cdn/4.0/domains/example.com/dns-records",
			servermock.ResponseFromFixture("create_txt_record.json").
				WithStatusCode(http.StatusCreated),
			servermock.CheckRequestJSONBodyFromFixture("create_record-request.json"),
		).
		Build(t)

	record := DNSRecord{
		Name:          "_acme-challenge",
		Type:          "txt",
		Value:         &TXTRecordValue{Text: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY"},
		TTL:           600,
		UpstreamHTTPS: "default",
		IPFilterMode: &IPFilterMode{
			Count:     "single",
			Order:     "none",
			GeoFilter: "none",
		},
	}

	newRecord, err := client.CreateRecord(t.Context(), "example.com", record)
	require.NoError(t, err)

	expected := &DNSRecord{
		ID:            "fe93c082-70d9-4d53-a121-66928adc40c8",
		Type:          "txt",
		Value:         map[string]any{"text": "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY"},
		Name:          "_acme-challenge",
		TTL:           600,
		UpstreamHTTPS: "default",
		IPFilterMode: &IPFilterMode{
			Count:     "single",
			Order:     "none",
			GeoFilter: "none",
		},
	}

	assert.Equal(t, expected, newRecord)
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder("myKeyC").
		Route("DELETE /cdn/4.0/domains/example.com/dns-records/recordId",
			servermock.Noop(),
		).
		Build(t)

	err := client.DeleteRecord(t.Context(), "example.com", "recordId")
	require.NoError(t, err)
}
