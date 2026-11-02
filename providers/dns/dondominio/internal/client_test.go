package internal

import (
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
			client, err := NewClient("user123", "secret")
			if err != nil {
				return nil, err
			}

			client.BaseURL, _ = url.Parse(server.URL)
			client.HTTPClient = server.Client()

			return client, nil
		},
		servermock.CheckHeader().
			WithAccept("application/json").
			WithContentTypeFromURLEncoded(),
	)
}

func TestClient_CreateRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /service/dnscreate",
			servermock.ResponseFromFixture("create_record.json"),
			servermock.CheckForm().Strict().
				With("name", "_acme-challenge.example.com").
				With("serviceName", "example.com").
				With("ttl", "600").
				With("type", "TXT").
				With("value", "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY").
				With("apipasswd", "secret").
				With("apiuser", "user123"),
		).
		Build(t)

	payload := CreateRecordRequest{
		ServiceName: "example.com",
		Name:        "_acme-challenge.example.com",
		Type:        "TXT",
		Value:       "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		TTL:         "600",
	}

	records, err := client.CreateRecord(t.Context(), payload)
	require.NoError(t, err)

	expected := []Record{{
		EntityID: "SRV0000214277-000396072",
		Name:     "_acme-challenge.example.com",
		Type:     "TXT",
		TTL:      "600",
		Value:    "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
	}}

	assert.Equal(t, expected, records)
}

func TestClient_CreateRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /service/dnscreate",
			servermock.ResponseFromFixture("error.json"),
		).
		Build(t)

	payload := CreateRecordRequest{
		ServiceName: "example.com",
		Name:        "_acme-challenge.example.com",
		Type:        "TXT",
		Value:       "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		TTL:         "-",
	}

	_, err := client.CreateRecord(t.Context(), payload)
	require.EqualError(t, err, "102: tool/requestinvalid: Object action invalid: invalid object/action")
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /service/dnsdelete",
			servermock.ResponseFromFixture("delete_record.json"),
			servermock.CheckForm().Strict().
				With("serviceName", "example.com").
				With("entityID", "SRV0000214277-000396072").
				With("apipasswd", "secret").
				With("apiuser", "user123"),
		).
		Build(t)

	payload := DeleteRecordRequest{
		ServiceName: "example.com",
		EntityID:    "SRV0000214277-000396072",
	}

	err := client.DeleteRecord(t.Context(), payload)
	require.NoError(t, err)
}
