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

func mockBuilderAuthenticated() *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client, err := NewClient(server.URL, "userA", "secret")
			if err != nil {
				return nil, err
			}

			client.baseURL, _ = url.Parse(server.URL)
			client.HTTPClient = server.Client()

			return client, nil
		},
		servermock.CheckHeader().
			WithJSONHeaders(),
		servermock.CheckHeader().
			WithAuthorization("Basic secretToken"),
	)
}

func TestClient_RetrieveConfigurations(t *testing.T) {
	client := mockBuilderAuthenticated().
		Route("GET /api/v2/configurations",
			servermock.ResponseFromFixture("configurations.json"),
			servermock.CheckQueryParameter().Strict().
				With("filter", "name:eq('myConfiguration')"),
		).
		Build(t)

	opts := &CollectionOptions{
		Filter: Eq("name", "myConfiguration").String(),
	}

	result, err := client.RetrieveConfigurations(mockToken(t.Context()), opts)
	require.NoError(t, err)

	expected := []CommonResource{
		{ID: 12345, Type: "Configuration", Name: "myConfiguration"},
	}

	assert.Equal(t, expected, result)
}

func TestClient_RetrieveConfigurations_error(t *testing.T) {
	client := mockBuilderAuthenticated().
		Route("GET /api/v2/configurations",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized),
		).
		Build(t)

	opts := &CollectionOptions{
		Filter: Eq("name", "myConfiguration").String(),
	}

	_, err := client.RetrieveConfigurations(mockToken(t.Context()), opts)
	require.EqualError(t, err, "401: Unauthorized: InvalidAuthorizationToken: The provided authorization token is invalid")
}

func TestClient_RetrieveConfigurationViews(t *testing.T) {
	client := mockBuilderAuthenticated().
		Route("GET /api/v2/configurations/12345/views",
			servermock.ResponseFromFixture("views.json"),
			servermock.CheckQueryParameter().Strict().
				With("filter", "name:eq('myView')"),
		).
		Build(t)

	opts := &CollectionOptions{
		Filter: Eq("name", "myView").String(),
	}

	result, err := client.RetrieveConfigurationViews(mockToken(t.Context()), 12345, opts)
	require.NoError(t, err)

	expected := []CommonResource{
		{ID: 12345, Type: "View", Name: "myView"},
	}

	assert.Equal(t, expected, result)
}

func TestClient_RetrieveViewZones(t *testing.T) {
	client := mockBuilderAuthenticated().
		Route("GET /api/v2/views/12345/zones",
			servermock.ResponseFromFixture("zones.json"),
			servermock.CheckQueryParameter().Strict().
				With("filter", "absoluteName:eq('example.com')"),
		).
		Build(t)

	opts := &CollectionOptions{
		Filter: Eq("absoluteName", "example.com").String(),
	}

	result, err := client.RetrieveViewZones(mockToken(t.Context()), 12345, opts)
	require.NoError(t, err)

	expected := []ZoneResource{
		{
			CommonResource: CommonResource{ID: 12345, Type: "ENUMZone", Name: "5678"},
			AbsoluteName:   "string",
		},
		{
			CommonResource: CommonResource{ID: 12345, Type: "ExternalHostsZone", Name: "name"},
		},
		{
			CommonResource: CommonResource{ID: 12345, Type: "InternalRootZone", Name: "name"},
		},
		{
			CommonResource: CommonResource{ID: 12345, Type: "ResponsePolicyZone", Name: "name"},
		},
		{
			CommonResource: CommonResource{ID: 12345, Type: "Zone", Name: "example.com"},
			AbsoluteName:   "example.com",
		},
	}

	assert.Equal(t, expected, result)
}

func TestClient_RetrieveZones(t *testing.T) {
	client := mockBuilderAuthenticated().
		Route("GET /api/v2/zones",
			servermock.ResponseFromFixture("zones.json"),
			servermock.CheckQueryParameter().Strict().
				With(
					"filter",
					"absoluteName:eq('example.com') and configuration.name:eq('myConfiguration') and view.name:eq('myView')",
				),
		).
		Build(t)

	// NOTE(ldez): I don't know if this approach with 3 "and" works in reality.
	opts := &CollectionOptions{
		Filter: And(
			Eq("absoluteName", "example.com"),
			Eq("configuration.name", "myConfiguration"),
			Eq("view.name", "myView"),
		).String(),
	}

	result, err := client.RetrieveZones(mockToken(t.Context()), opts)
	require.NoError(t, err)

	expected := []ZoneResource{
		{
			CommonResource: CommonResource{ID: 12345, Type: "ENUMZone", Name: "5678"},
			AbsoluteName:   "string",
		},
		{
			CommonResource: CommonResource{ID: 12345, Type: "ExternalHostsZone", Name: "name"},
		},
		{
			CommonResource: CommonResource{ID: 12345, Type: "InternalRootZone", Name: "name"},
		},
		{
			CommonResource: CommonResource{ID: 12345, Type: "ResponsePolicyZone", Name: "name"},
		},
		{
			CommonResource: CommonResource{ID: 12345, Type: "Zone", Name: "example.com"},
			AbsoluteName:   "example.com",
		},
	}

	assert.Equal(t, expected, result)
}

func TestClient_CreateZoneResourceRecord(t *testing.T) {
	client := mockBuilderAuthenticated().
		Route("POST /api/v2/zones/12345/resourceRecords",
			servermock.ResponseFromFixture("postZoneResourceRecord.json"),
			servermock.CheckRequestJSONBodyFromFixture("postZoneResourceRecord-request.json"),
		).
		Build(t)

	record := RecordTXT{
		CommonResource: CommonResource{
			Type: "TXTRecord",
			Name: "_acme-challenge",
		},
		Text:       "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		TTL:        120,
		RecordType: "TXT",
	}

	result, err := client.CreateZoneResourceRecord(mockToken(t.Context()), 12345, record)
	require.NoError(t, err)

	expected := &RecordTXT{
		CommonResource: CommonResource{
			ID:   12345,
			Type: "ResourceRecord",
			Name: "name",
		},
		TTL:          3600,
		AbsoluteName: "host1.example.com",
		Comment:      "Sample comment.",
		Dynamic:      true,
		RecordType:   "CNAME",
		Text:         "",
	}

	assert.Equal(t, expected, result)
}

func TestClient_DeleteResourceRecord(t *testing.T) {
	client := mockBuilderAuthenticated().
		Route("DELETE /api/v2/resourceRecords/12345",
			servermock.ResponseFromFixture("deleteResourceRecord.json"),
		).
		Build(t)

	err := client.DeleteResourceRecord(mockToken(t.Context()), 12345)
	require.NoError(t, err)
}
