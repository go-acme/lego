package internal

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

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

func TestClient_RetrieveZones_error(t *testing.T) {
	client := mockBuilderAuthenticated().
		Route("GET /api/v2/zones",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized),
		).
		Build(t)

	opts := &CollectionOptions{
		Filter: And(
			Eq("absoluteName", "example.com"),
			Eq("configuration.name", "myConfiguration"),
			Eq("view.name", "myView"),
		).String(),
	}

	_, err := client.RetrieveZones(mockToken(t.Context()), opts)
	require.EqualError(t, err, "401: Unauthorized: InvalidAuthorizationToken: The provided authorization token is invalid")
}

func TestClient_RetrieveZoneDeployments(t *testing.T) {
	client := mockBuilderAuthenticated().
		Route("GET /api/v2/zones/456789/deployments",
			servermock.ResponseFromFixture("getZoneDeployments.json"),
			servermock.CheckQueryParameter().Strict().
				With("filter", "id:eq('12345')"),
		).
		Build(t)

	opts := &CollectionOptions{
		Filter: Eq("id", "12345").String(),
	}

	result, err := client.RetrieveZoneDeployments(mockToken(t.Context()), 456789, opts)
	require.NoError(t, err)

	expected := []QuickDeployment{
		{
			CommonResource:     CommonResource{ID: 12345, Type: "QuickDeployment", Name: ""},
			State:              "PENDING",
			Status:             "CANCEL",
			Message:            "string",
			PercentComplete:    50,
			CreationDateTime:   time.Date(2022, time.November, 23, 2, 53, 0, 0, time.UTC),
			StartDateTime:      time.Date(2022, time.November, 23, 2, 53, 3, 0, time.UTC),
			CompletionDateTime: time.Date(2022, time.November, 23, 2, 54, 5, 0, time.UTC),
			Method:             "SCHEDULED",
		},
	}

	assert.Equal(t, expected, result)
}

func TestClient_CreateZoneDeployment(t *testing.T) {
	client := mockBuilderAuthenticated().
		Route("POST /api/v2/zones/12345/deployments",
			servermock.ResponseFromFixture("postZoneDeployment.json").
				WithStatusCode(http.StatusCreated),
			servermock.CheckRequestJSONBodyFromFixture("postZoneDeployment-request.json"),
		).
		Build(t)

	quickDeployment, err := client.CreateZoneDeployment(mockToken(t.Context()), 12345)
	require.NoError(t, err)

	expected := &QuickDeployment{
		CommonResource:     CommonResource{ID: 12345, Type: "QuickDeployment"},
		State:              "PENDING",
		Status:             "CANCEL",
		Message:            "string",
		PercentComplete:    50,
		CreationDateTime:   time.Date(2022, time.November, 23, 2, 53, 0, 0, time.UTC),
		StartDateTime:      time.Date(2022, time.November, 23, 2, 53, 3, 0, time.UTC),
		CompletionDateTime: time.Date(2022, time.November, 23, 2, 54, 5, 0, time.UTC),
		Method:             "SCHEDULED",
	}

	assert.Equal(t, expected, quickDeployment)
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
