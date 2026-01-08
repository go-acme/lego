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
			WithAuthorization("Bearer secret"),
	)
}

func TestClient_CreateTXTRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /dns_zones/dns_zone/records",
			servermock.ResponseFromFixture("create_record.json"),
			servermock.CheckRequestJSONBodyFromFixture("create_record-request.json"),
		).
		Build(t)

	record := RecordProperties{
		Name: "_acme-challenge",
		Type: "TXT",
		TTL:  120,
		Content: &RecordContent{
			TXT: &RecordTXT{
				Content: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
			},
		},
	}

	response, err := client.CreateTXTRecord(t.Context(), "example.com", record)
	require.NoError(t, err)

	expected := &DNSRecordResponse{
		ID:       "abc123",
		Name:     "_acme-challenge",
		FullName: "_acme-challenge.example.com",
		TTL:      120,
		Type:     "TXT",
		Content:  "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		ContentAttributes: &RecordContent{
			TXT: &RecordTXT{Content: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY"},
		},
	}

	assert.Equal(t, expected, response)
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /dns_zones/dns_record",
			servermock.ResponseFromFixture("delete_record.json"),
			servermock.CheckRequestJSONBodyFromFixture("delete_record-request.json"),
		).
		Build(t)

	err := client.DeleteRecord(t.Context(), "abc123")
	require.NoError(t, err)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /dns_zones/dns_record",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusBadRequest),
		).
		Build(t)

	err := client.DeleteRecord(t.Context(), "abc123")
	require.EqualError(t, err, "missing_api_token: string: {}")
}
