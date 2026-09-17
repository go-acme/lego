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
			client, err := NewClient()
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

func TestClient_GetDomainDetails(t *testing.T) {
	client := mockBuilder().
		Route("GET /domains/example.com",
			servermock.ResponseFromFixture("domain_details.json"),
		).
		Build(t)

	result, err := client.GetDomainDetails(mockContext(t), "example.com")
	require.NoError(t, err)

	expected := &Domain{
		DomainID:    12345,
		Domain:      "example.com",
		Status:      "hotovo",
		ObjectType:  "DOMAIN",
		HostingOnly: 0,
		State:       "OK",
		TLD:         "com",
	}

	assert.Equal(t, expected, result)
}

func TestClient_CreateRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /12345/dns",
			servermock.ResponseFromFixture("create_record.json"),
			servermock.CheckRequestJSONBodyFromFixture("create_record-request.json"),
		).
		Build(t)

	record := Record{
		Name: "_acme-challenge",
		Type: "TXT",
		Data: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		TTL:  120,
	}

	result, err := client.CreateRecord(mockContext(t), 12345, record)
	require.NoError(t, err)

	expected := &Record{
		ID:   456789,
		Zone: 45,
		Name: "_acme-challenge",
		Type: "TXT",
		Data: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		TTL:  120,
	}

	assert.Equal(t, expected, result)
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /12345/dns/456789",
			servermock.Noop(),
		).
		Build(t)

	err := client.DeleteRecord(mockContext(t), 12345, 456789)
	require.NoError(t, err)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /12345/dns/456789",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusBadRequest),
		).
		Build(t)

	err := client.DeleteRecord(mockContext(t), 12345, 456789)

	require.EqualError(t, err, "1: Unknown error - problem has been reported and will be fixed soon.")
}
