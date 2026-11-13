package internal

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client, err := NewClient(server.URL, "secret", "p123")
			if err != nil {
				return nil, err
			}

			client.HTTPClient = server.Client()

			return client, nil
		},
		servermock.CheckHeader().
			WithJSONHeaders().
			WithAuthorization("ApiKey secret"),
	)
}

func TestClient_ListZones(t *testing.T) {
	client := mockBuilder().
		Route("GET /project/p123/zones",
			servermock.ResponseFromFixture("zones.json"),
		).
		Build(t)

	zones, err := client.ListZones(t.Context())
	require.NoError(t, err)

	expected := []Zone{
		{
			Metadata: Metadata{
				Name:            "z123",
				ResourceVersion: "rv001",
			},
			Spec: ZoneSpec{
				Zone: "example.com",
			},
		},
	}

	assert.Equal(t, expected, zones)
}

func TestClient_ListDNSEndpoints(t *testing.T) {
	client := mockBuilder().
		Route("GET /project/p123/zones/z123/dns-endpoints",
			servermock.ResponseFromFixture("dns_endpoints.json"),
		).
		Build(t)

	des, err := client.ListDNSEndpoints(t.Context(), "z123")
	require.NoError(t, err)

	expected := []Record{
		{
			Metadata: Metadata{
				Name:            "r123",
				ResourceVersion: "mrv101",
			},
			Spec: RecordSpec{
				DNSName:       "_acme-challenge.example.com",
				TTL:           120,
				Type:          "TXT",
				RoutingPolicy: RoutingPolicySimple,
				Targets:       []string{"foo"},
			},
		},
	}

	assert.Equal(t, expected, des)
}

func TestClient_CreateDNSEndpoint(t *testing.T) {
	client := mockBuilder().
		Route("POST /project/p123/zones/z123/dns-endpoints",
			servermock.ResponseFromFixture("create_dns_endpoint.json"),
			servermock.CheckRequestJSONBodyFromFixture("create_dns_endpoint-request.json"),
		).
		Build(t)

	record := RecordSpec{
		DNSName:       "_acme-challenge.example.com",
		TTL:           120,
		Type:          "TXT",
		RoutingPolicy: RoutingPolicySimple,
		Targets:       []string{"ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY"},
	}

	err := client.CreateDNSEndpoint(t.Context(), "z123", record)
	require.NoError(t, err)
}

func TestClient_UpdateDNSEndpoint(t *testing.T) {
	client := mockBuilder().
		Route("PATCH /project/p123/zones/z123/dns-endpoints/e123",
			servermock.ResponseFromFixture("update_dns_endpoint.json"),
			servermock.CheckRequestJSONBodyFromFixture("update_dns_endpoint-request.json"),
		).
		Build(t)

	record := RecordSpec{
		Targets: []string{"foo", "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY"},
	}

	err := client.UpdateDNSEndpoint(t.Context(), "z123", "e123", record)
	require.NoError(t, err)
}

func TestClient_DeleteDNSEndpoint(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /project/p123/zones/z123/dns-endpoints/e123",
			servermock.Noop().
				WithStatusCode(http.StatusNoContent),
			servermock.CheckQueryParameter().Strict().
				With("resourceVersion", "mrv101"),
		).
		Build(t)

	err := client.DeleteDNSEndpoint(t.Context(), "z123", "e123", "mrv101")
	require.NoError(t, err)
}

func TestClient_DeleteDNSEndpoint_error(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /project/p123/zones/z123/dns-endpoints/e123",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized),
		).
		Build(t)

	err := client.DeleteDNSEndpoint(t.Context(), "z123", "e123", "mrv101")
	require.EqualError(t, err, "[status code 401]: invalid API key")
}
