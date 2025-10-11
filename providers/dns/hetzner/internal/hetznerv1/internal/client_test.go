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
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client, err := NewClient(OAuthStaticAccessToken(server.Client(), "secret"))
			if err != nil {
				return nil, err
			}

			client.BaseURL, _ = url.Parse(server.URL)

			return client, nil
		},
		servermock.CheckHeader().
			WithJSONHeaders().
			WithAuthorization("Bearer secret"),
	)
}

func TestClient_AddRRSetRecords(t *testing.T) {
	client := mockBuilder().
		Route("POST /zones/example.com/rrsets/www/TXT/actions/add_records",
			servermock.ResponseFromFixture("add_rrset_records.json"),
			servermock.CheckRequestJSONBodyFromFixture("add_rrset_records-request.json")).
		Build(t)

	records := []Record{{
		Value:   "198.51.100.1",
		Comment: "My web server at Hetzner Cloud.",
	}}

	result, err := client.AddRRSetRecords(t.Context(), "example.com", "TXT", "www", 3600, records)
	require.NoError(t, err)

	expected := &Action{
		ID:        1,
		Command:   "add_rrset_records",
		Status:    "running",
		Progress:  50,
		Resources: []Resources{{ID: 42, Type: "zone"}},
	}

	assert.Equal(t, expected, result)
}

func TestClient_AddRRSetRecords_error_invalid_input(t *testing.T) {
	client := mockBuilder().
		Route("POST /zones/example.com/rrsets/www/TXT/actions/add_records",
			servermock.ResponseFromFixture("error-invalid_input.json").
				WithStatusCode(http.StatusBadRequest)).
		Build(t)

	records := []Record{{
		Value:   "198.51.100.1",
		Comment: "My web server at Hetzner Cloud.",
	}}

	_, err := client.AddRRSetRecords(t.Context(), "example.com", "TXT", "www", 0, records)
	require.EqualError(t, err, "invalid_input: invalid input in field 'broken_field': is too longfield: broken_field: is too long")
}

func TestClient_AddRRSetRecords_error_resource_limit_exceeded(t *testing.T) {
	client := mockBuilder().
		Route("POST /zones/example.com/rrsets/www/TXT/actions/add_records",
			servermock.ResponseFromFixture("error-resource_limit_exceeded.json").
				WithStatusCode(http.StatusBadRequest)).
		Build(t)

	records := []Record{{
		Value:   "198.51.100.1",
		Comment: "My web server at Hetzner Cloud.",
	}}

	_, err := client.AddRRSetRecords(t.Context(), "example.com", "TXT", "www", 0, records)
	require.EqualError(t, err, "resource_limit_exceeded: project limit exceededlimit: project_limit")
}

func TestClient_AddRRSetRecords_error_deprecated_api_endpoint(t *testing.T) {
	client := mockBuilder().
		Route("POST /zones/example.com/rrsets/www/TXT/actions/add_records",
			servermock.ResponseFromFixture("error-deprecated_api_endpoint.json").
				WithStatusCode(http.StatusBadRequest)).
		Build(t)

	records := []Record{{
		Value:   "198.51.100.1",
		Comment: "My web server at Hetzner Cloud.",
	}}

	_, err := client.AddRRSetRecords(t.Context(), "example.com", "TXT", "www", 0, records)
	require.EqualError(t, err, "deprecated_api_endpoint: API functionality was removed: https://docs.hetzner.cloud/changelog#2023-07-20-foo-endpoint-is-deprecated")
}

func TestClient_RemoveRRSetRecords(t *testing.T) {
	client := mockBuilder().
		Route("POST /zones/example.com/rrsets/www/TXT/actions/remove_records",
			servermock.ResponseFromFixture("remove_rrset_records.json"),
			servermock.CheckRequestJSONBodyFromFixture("remove_rrset_records-request.json")).
		Build(t)

	records := []Record{{
		Value:   "198.51.100.1",
		Comment: "My web server at Hetzner Cloud.",
	}}

	result, err := client.RemoveRRSetRecords(t.Context(), "example.com", "TXT", "www", records)
	require.NoError(t, err)

	expected := &Action{
		ID:        1,
		Command:   "remove_rrset_records",
		Status:    "running",
		Progress:  50,
		Resources: []Resources{{ID: 42, Type: "zone"}},
	}

	assert.Equal(t, expected, result)
}

func TestClient_GetAction(t *testing.T) {
	client := mockBuilder().
		Route("GET /actions/123", servermock.ResponseFromFixture("get_action.json")).
		Route("/", servermock.DumpRequest()).
		Build(t)

	result, err := client.GetAction(t.Context(), 123)
	require.NoError(t, err)

	expected := &Action{
		ID:        42,
		Command:   "start_resource",
		Status:    "running",
		Progress:  100,
		Resources: []Resources{{ID: 42, Type: "server"}},
		ErrorInfo: &ErrorInfo{
			Code:    "action_failed",
			Message: "Action failed",
		},
	}

	assert.Equal(t, expected, result)
}
