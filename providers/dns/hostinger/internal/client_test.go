package internal

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder(
		func(server *httptest.Server) (*Client, error) {
			client, err := NewClient("secret")
			if err != nil {
				return nil, err
			}

			client.BaseURL, _ = url.Parse(server.URL)
			client.HTTPClient = server.Client()

			return client, nil
		},
		servermock.CheckHeader().WithJSONHeaders().
			With("Authorization", "Bearer secret"),
	)
}

func TestClient_GetDNSRecords(t *testing.T) {
	client := mockBuilder().
		Route("GET /api/dns/v1/zones/example.com",
			servermock.ResponseFromFixture("get_dns_records.json")).
		Build(t)

	records, err := client.GetDNSRecords(t.Context(), "example.com")
	require.NoError(t, err)

	expected := []RecordSet{
		{
			Name: "_acme-challenge",
			Records: []Record{{
				Content: "aaa",
			}},
			TTL:  14400,
			Type: "TXT",
		},
		{
			Name: "_acme-challenge",
			Records: []Record{{
				Content: "example.com.",
			}},
			TTL:  14400,
			Type: "A",
		},
	}

	assert.Equal(t, expected, records)
}

func TestClient_GetDNSRecords_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /api/dns/v1/zones/example.com",
			servermock.ResponseFromFixture("error_401.json").
				WithStatusCode(http.StatusUnauthorized)).
		Build(t)

	_, err := client.GetDNSRecords(t.Context(), "example.com")

	require.EqualError(t, err, "26a91bd9-f8c8-4a83-9df9-83e23d696fe3: Unauthenticated")
}

func TestClient_UpdateDNSRecords(t *testing.T) {
	client := mockBuilder().
		Route("PUT /api/dns/v1/zones/example.com",
			servermock.ResponseFromFixture("update_dns_records.json"),
			servermock.CheckRequestJSONBodyFromFixture("update_dns_records-request.json")).
		Build(t)

	zone := ZoneRequest{
		Overwrite: false,
		Zone: []RecordSet{
			{
				Name: "_acme-challenge",
				Records: []Record{
					{Content: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY"},
				},
				TTL:  120,
				Type: "TXT",
			},
		},
	}

	err := client.UpdateDNSRecords(t.Context(), "example.com", zone)
	require.NoError(t, err)
}

func TestClient_UpdateDNSRecords_error(t *testing.T) {
	client := mockBuilder().
		Route("PUT /api/dns/v1/zones/example.com",
			servermock.ResponseFromFixture("error_422.json").
				WithStatusCode(http.StatusBadRequest)).
		Build(t)

	zone := ZoneRequest{
		Zone: []RecordSet{{
			Name: "_acme-challenge",
			Records: []Record{{
				Content: "aaa",
			}},
			TTL:  14400,
			Type: "TXT",
		}},
	}

	err := client.UpdateDNSRecords(t.Context(), "example.com", zone)

	require.EqualError(t, err, "26a91bd9-f8c8-4a83-9df9-83e23d696fe3: The name field is required. (and 1 more error): field_1: The field_1 field is required., The field_1 must be a number.")
}

func TestClient_DeleteDNSRecords(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /api/dns/v1/zones/example.com",
			servermock.ResponseFromFixture("delete_dns_records.json"),
			servermock.CheckRequestJSONBody(`{"filters":[{"name":"_acme-challenge","type":"TXT"}]}`)).
		Build(t)

	filters := []Filter{{
		Name: "_acme-challenge",
		Type: "TXT",
	}}

	err := client.DeleteDNSRecords(t.Context(), "example.com", filters)
	require.NoError(t, err)
}

func TestClient_DeleteDNSRecords_error(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /api/dns/v1/zones/example.com",
			servermock.ResponseFromFixture("error_401.json").
				WithStatusCode(http.StatusUnauthorized)).
		Build(t)

	filters := []Filter{{
		Name: "_acme-challenge",
		Type: "TXT",
	}}

	err := client.DeleteDNSRecords(t.Context(), "example.com", filters)

	require.EqualError(t, err, "26a91bd9-f8c8-4a83-9df9-83e23d696fe3: Unauthenticated")
}
