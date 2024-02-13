package internal

import (
	"context"
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
			client := NewClient("secret")
			client.HTTPClient = server.Client()
			client.baseURL, _ = url.Parse(server.URL)

			return client, nil
		},
		servermock.CheckHeader().
			WithJSONHeaders().
			With(apiKeyHeader, "secret"),
	)
}

func TestUpdateZone(t *testing.T) {
	domain := "example.com"

	client := mockBuilder().
		Route("PATCH /v1/zones/"+domain,
			servermock.ResponseFromFixture("update.json"),
			servermock.CheckRequestJSONBodyFromFixture("update-request.json")).
		Build(t)

	patch := ZoneRequest{
		Data: Zone{
			Type: "zone",
			ID:   domain,
			Attributes: Attributes{
				Records: map[string]map[string][]Record{
					"_acme-challenge.test": {
						"TXT": []Record{
							{Data: "test"},
							{Data: "test1"},
							{Data: "test2"},
						},
					},
				},
			},
		},
	}

	zone, err := client.UpdateZone(context.Background(), domain, patch)
	require.NoError(t, err)

	expected := &APIResponse[*Zone]{
		Meta: &Metadata{
			InvocationID: "95cdcc21-b9c3-4b21-8bd1-b05c34c56147",
		},
		Data: &Zone{
			Type: "zone",
			ID:   "dipcon.com",
			Attributes: Attributes{
				OrganisationID:          "10154",
				OrganisationDescription: "My Company AB",
				DNSTypeDescription:      "Anycast",
				Slave:                   false,
				Pending:                 false,
				Deleted:                 false,
				Settings: &Settings{
					MName:   "dns01.dipcon.com.",
					Refresh: 3600,
					Expire:  604800,
					TTL:     600,
				},
				Records: map[string]map[string][]Record{
					"@": {
						"NS": {
							{
								TTL:      3600,
								Data:     "193.14.90.194",
								Comments: "this is a comment",
							},
						},
					},
				},
				Redirects: map[string][]Redirect{
					"<name>": {
						{
							Path:        "/x/y",
							Destination: "https://abion.com/?ref=dipcon",
							Status:      301,
							Slugs:       true,
							Certificate: true,
						},
					},
				},
			},
		},
	}

	assert.Equal(t, expected, zone)
}

func TestUpdateZone_error(t *testing.T) {
	domain := "example.com"

	client := mockBuilder().
		Route("PATCH /v1/zones/"+domain,
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized)).
		Build(t)

	patch := ZoneRequest{
		Data: Zone{
			Type: "zone",
			ID:   domain,
			Attributes: Attributes{
				Records: map[string]map[string][]Record{
					"_acme-challenge.test": {
						"TXT": []Record{
							{Data: "test"},
							{Data: "test1"},
							{Data: "test2"},
						},
					},
				},
			},
		},
	}

	_, err := client.UpdateZone(context.Background(), domain, patch)
	require.EqualError(t, err, "could not update zone example.com: api error: status=401, message=Authentication Error")
}

func TestGetZones(t *testing.T) {
	client := mockBuilder().
		Route("GET /v1/zones/",
			servermock.ResponseFromFixture("zones.json")).
		Build(t)

	zones, err := client.GetZones(context.Background(), nil)
	require.NoError(t, err)

	expected := &APIResponse[[]Zone]{
		Meta: &Metadata{
			InvocationID: "95cdcc21-b9c3-4b21-8bd1-b05c34c56147",
			Pagination: &Pagination{
				Offset: 0,
				Limit:  1,
				Total:  1,
			},
		},
		Data: []Zone{
			{
				Type: "zone",
				ID:   "dipcon.com",
				Attributes: Attributes{
					OrganisationID:          "10154",
					OrganisationDescription: "My Company AB",
					DNSTypeDescription:      "Anycast",
					Slave:                   true,
					Pending:                 true,
					Deleted:                 true,
				},
			},
		},
	}

	assert.Equal(t, expected, zones)
}

func TestGetZones_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /v1/zones/",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized)).
		Build(t)

	_, err := client.GetZones(context.Background(), nil)
	require.EqualError(t, err, "could not get zones: api error: status=401, message=Authentication Error")
}

func TestGetZone(t *testing.T) {
	client := mockBuilder().
		Route("GET /v1/zones/example.com",
			servermock.ResponseFromFixture("zone.json")).
		Build(t)

	zones, err := client.GetZone(context.Background(), "example.com")
	require.NoError(t, err)

	expected := &APIResponse[*Zone]{
		Meta: &Metadata{
			InvocationID: "95cdcc21-b9c3-4b21-8bd1-b05c34c56147",
		},
		Data: &Zone{
			Type: "zone",
			ID:   "dipcon.com",
			Attributes: Attributes{
				OrganisationID:          "10154",
				OrganisationDescription: "My Company AB",
				DNSTypeDescription:      "Anycast",
				Slave:                   false,
				Pending:                 false,
				Deleted:                 false,
				Settings: &Settings{
					MName:   "dns01.dipcon.com.",
					Refresh: 3600,
					Expire:  604800,
					TTL:     600,
				},
				Records: map[string]map[string][]Record{
					"@": {
						"NS": {
							{
								TTL:      3600,
								Data:     "193.14.90.194",
								Comments: "this is a comment",
							},
						},
					},
				},
				Redirects: map[string][]Redirect{
					"<name>": {
						{
							Path:        "/x/y",
							Destination: "https://abion.com/?ref=dipcon",
							Status:      301,
							Slugs:       true,
							Certificate: true,
						},
					},
				},
			},
		},
	}

	assert.Equal(t, expected, zones)
}

func TestGetZone_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /v1/zones/example.com",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized)).
		Build(t)

	_, err := client.GetZone(context.Background(), "example.com")
	require.EqualError(t, err, "could not get zone example.com: api error: status=401, message=Authentication Error")
}
