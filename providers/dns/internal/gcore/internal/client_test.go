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

const (
	testToken         = "secret"
	testRecordContent = "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY"
	testTTL           = 120
)

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client := NewClient(testToken)
			client.BaseURL, _ = url.Parse(server.URL)
			client.HTTPClient = server.Client()

			return client, nil
		},
		servermock.CheckHeader().
			WithJSONHeaders().
			WithAuthorization("APIKey secret"),
	)
}

func TestClient_GetZone(t *testing.T) {
	client := mockBuilder().
		Route("GET /v2/zones/example.com",
			servermock.ResponseFromFixture("get_zones.json"),
		).
		Build(t)

	zone, err := client.GetZone(t.Context(), "example.com")
	require.NoError(t, err)

	expected := Zone{Name: "example.com"}

	assert.Equal(t, expected, zone)
}

func TestClient_GetZone_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /v2/zones/example.com",
			servermock.JSONEncode(APIError{Message: "oops"}).
				WithStatusCode(http.StatusInternalServerError),
		).
		Build(t)

	_, err := client.GetZone(t.Context(), "example.com")
	require.EqualError(t, err, "get zone example.com: 500: oops")
}

func TestClient_GetRRSet(t *testing.T) {
	client := mockBuilder().
		Route("GET /v2/zones/example.com/foo.example.com/TXT",
			servermock.ResponseFromFixture("get_rrset.json"),
		).
		Build(t)

	rrSet, err := client.GetRRSet(t.Context(), "example.com", "foo.example.com")
	require.NoError(t, err)

	expected := RRSet{
		TTL: 10,
		Records: []Records{
			{Content: []string{"foo"}},
		},
	}

	assert.Equal(t, expected, rrSet)
}

func TestClient_GetRRSet_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /v2/zones/example.com/foo.example.com/TXT",
			servermock.JSONEncode(APIError{Message: "oops"}).
				WithStatusCode(http.StatusInternalServerError),
		).
		Build(t)

	_, err := client.GetRRSet(t.Context(), "example.com", "foo.example.com")
	require.EqualError(t, err, "get txt records example.com -> foo.example.com: 500: oops")
}

func TestClient_DeleteRRSet(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /v2/zones/test.example.com/my.test.example.com/TXT",
			servermock.Noop(),
		).
		Build(t)

	err := client.DeleteRRSet(t.Context(), "test.example.com", "my.test.example.com.")
	require.NoError(t, err)
}

func TestClient_DeleteRRSet_error(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /v2/zones/test.example.com/my.test.example.com/TXT",
			servermock.JSONEncode(APIError{Message: "oops"}).
				WithStatusCode(http.StatusInternalServerError),
		).
		Build(t)

	err := client.DeleteRRSet(t.Context(), "test.example.com", "my.test.example.com.")
	require.NoError(t, err)
}

func TestClient_AddRRSet_add(t *testing.T) {
	client := mockBuilder().
		// GetRRSet
		Route("GET /v2/zones/example.com/_acme-challenge.example.com/TXT",
			servermock.JSONEncode(APIError{Message: "not found"}).
				WithStatusCode(http.StatusBadRequest),
		).
		// createRRSet
		Route("POST /v2/zones/example.com/_acme-challenge.example.com/TXT",
			servermock.ResponseFromFixture("create_rrset.json"),
			servermock.CheckRequestJSONBodyFromFixture("create_rrset-request.json"),
		).
		Build(t)

	err := client.AddRRSet(t.Context(), "example.com", "_acme-challenge.example.com", testRecordContent, testTTL)
	require.NoError(t, err)
}

func TestClient_AddRRSet_add_error(t *testing.T) {
	client := mockBuilder().
		// GetRRSet
		Route("GET /v2/zones/test.example.com/my.test.example.com/TXT",
			servermock.JSONEncode(APIError{Message: "not found"}).
				WithStatusCode(http.StatusBadRequest),
		).
		// createRRSet
		Route("POST /v2/zones/test.example.com/my.test.example.com/TXT",
			servermock.JSONEncode(APIError{Message: "oops"}).
				WithStatusCode(http.StatusBadRequest),
		).
		Build(t)

	err := client.AddRRSet(t.Context(), "test.example.com", "my.test.example.com", testRecordContent, testTTL)
	require.EqualError(t, err, "400: oops")
}

func TestClient_AddRRSet_update(t *testing.T) {
	client := mockBuilder().
		// GetRRSet
		Route("GET /v2/zones/example.com/_acme-challenge.example.com/TXT",
			servermock.ResponseFromFixture("get_rrset.json"),
		).
		// updateRRSet
		Route("PUT /v2/zones/example.com/_acme-challenge.example.com/TXT",
			servermock.ResponseFromFixture("update_rrset.json"),
			servermock.CheckRequestJSONBodyFromFixture("update_rrset-request.json"),
		).
		Build(t)

	err := client.AddRRSet(t.Context(), "example.com", "_acme-challenge.example.com", testRecordContent, testTTL)
	require.NoError(t, err)
}

func TestClient_AddRRSet_update_error(t *testing.T) {
	client := mockBuilder().
		// GetRRSet
		Route("GET /v2/zones/example.com/_acme-challenge.example.com/TXT",
			servermock.ResponseFromFixture("get_rrset.json"),
		).
		// updateRRSet
		Route("PUT /v2/zones/example.com/_acme-challenge.example.com/TXT",
			servermock.JSONEncode(APIError{Message: "oops"}).
				WithStatusCode(http.StatusBadRequest),
		).
		Build(t)

	err := client.AddRRSet(t.Context(), "example.com", "_acme-challenge.example.com", testRecordContent, testTTL)
	require.EqualError(t, err, "400: oops")
}
