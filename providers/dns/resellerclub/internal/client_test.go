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
			client, err := NewClient("secret", "user123")
			if err != nil {
				return nil, err
			}

			client.BaseURL, _ = url.Parse(server.URL)
			client.HTTPClient = server.Client()

			return client, nil
		},
		servermock.CheckHeader().
			WithJSONHeaders().
			WithAuthorization("ApiKey secret").
			With(UserIDHeader, "user123"),
	)
}

func TestClient_CreateDNSRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /dns/zones/example.com/records",
			servermock.ResponseFromFixture("create_dns_record.json"),
			servermock.CheckRequestJSONBodyFromFixture("create_dns_record-request.json"),
		).
		Build(t)

	record := Record{
		Type:  "TXT",
		Host:  "_acme-challenge",
		Value: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		TTL:   120,
	}

	result, err := client.CreateDNSRecord(t.Context(), "example.com", record)
	require.NoError(t, err)

	expected := &APIResponse{
		Status:        "status",
		Message:       "message",
		TransactionID: "transactionId",
	}

	assert.Equal(t, expected, result)
}

func TestClient_CreateDNSRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /dns/zones/example.com/records",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusBadRequest),
		).
		Build(t)

	record := Record{
		Type:  "TXT",
		Host:  "_acme-challenge",
		Value: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		TTL:   120,
	}

	_, err := client.CreateDNSRecord(t.Context(), "example.com", record)
	require.EqualError(t, err, "401: /customers/33: Please add your user id an api key in the header.")
}

func TestClient_DeleteDNSRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /dns/zones/example.com/records",
			servermock.ResponseFromFixture("delete_dns_record.json"),
			servermock.CheckRequestJSONBodyFromFixture("delete_dns_record-request.json"),
		).
		Build(t)

	record := Record{
		Type:  "TXT",
		Host:  "_acme-challenge",
		Value: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
	}

	result, err := client.DeleteDNSRecord(t.Context(), "example.com", record)
	require.NoError(t, err)

	expected := &APIResponse{
		Status:        "status",
		Message:       "message",
		TransactionID: "transactionId",
	}

	assert.Equal(t, expected, result)
}
