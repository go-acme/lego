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

const (
	testToken         = "test"
	testRecordContent = "acme"
	testTTL           = 10
)

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client := NewClient(testToken)
			client.baseURL, _ = url.Parse(server.URL)
			client.HTTPClient = server.Client()

			return client, nil
		},
		servermock.CheckHeader().WithJSONHeaders())
}

func TestClient_GetZone(t *testing.T) {
	expected := Zone{Name: "example.com"}

	client := mockBuilder().
		Route("GET /v2/zones/example.com",
			servermock.JSONEncode(expected)).
		Build(t)

	zone, err := client.GetZone(t.Context(), "example.com")
	require.NoError(t, err)

	assert.Equal(t, expected, zone)
}

func TestClient_GetZone_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /v2/zones/example.com",
			servermock.JSONEncode(APIError{Message: "oops"}).WithStatusCode(http.StatusInternalServerError)).
		Build(t)

	_, err := client.GetZone(t.Context(), "example.com")
	require.EqualError(t, err, "get zone example.com: 500: oops")
}

func TestClient_GetRRSet(t *testing.T) {
	expected := RRSet{
		TTL: testTTL,
		Records: []Records{
			{Content: []string{testRecordContent}},
		},
	}

	client := mockBuilder().
		Route("GET /v2/zones/example.com/foo.example.com/TXT",
			servermock.JSONEncode(expected)).
		Build(t)

	rrSet, err := client.GetRRSet(t.Context(), "example.com", "foo.example.com")
	require.NoError(t, err)

	assert.Equal(t, expected, rrSet)
}

func TestClient_GetRRSet_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /v2/zones/example.com/foo.example.com/TXT",
			servermock.JSONEncode(APIError{Message: "oops"}).WithStatusCode(http.StatusInternalServerError)).
		Build(t)

	_, err := client.GetRRSet(t.Context(), "example.com", "foo.example.com")
	require.EqualError(t, err, "get txt records example.com -> foo.example.com: 500: oops")
}

func TestClient_DeleteRRSet(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /v2/zones/test.example.com/my.test.example.com/TXT", nil).
		Build(t)

	err := client.DeleteRRSet(t.Context(), "test.example.com", "my.test.example.com.")
	require.NoError(t, err)
}

func TestClient_DeleteRRSet_error(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /v2/zones/test.example.com/my.test.example.com/TXT",
			servermock.JSONEncode(APIError{Message: "oops"}).WithStatusCode(http.StatusInternalServerError)).
		Build(t)

	err := client.DeleteRRSet(t.Context(), "test.example.com", "my.test.example.com.")
	require.NoError(t, err)
}

func TestClient_AddRRSet_add(t *testing.T) {
	client := mockBuilder().
		// GetRRSet
		Route("GET /v2/zones/test.example.com/my.test.example.com/TXT",
			servermock.JSONEncode(APIError{Message: "not found"}).WithStatusCode(http.StatusBadRequest)).
		// createRRSet
		Route("POST /v2/zones/test.example.com/my.test.example.com/TXT",
			servermock.JSONEncode([]Records{{Content: []string{testRecordContent}}}),
			servermock.CheckRequestJSONBody(`{"ttl":10,"resource_records":[{"content":["acme"]}]}`)).
		Build(t)

	err := client.AddRRSet(t.Context(), "test.example.com", "my.test.example.com", testRecordContent, testTTL)
	require.NoError(t, err)
}

func TestClient_AddRRSet_add_error(t *testing.T) {
	client := mockBuilder().
		// GetRRSet
		Route("GET /v2/zones/test.example.com/my.test.example.com/TXT",
			servermock.JSONEncode(APIError{Message: "not found"}).WithStatusCode(http.StatusBadRequest)).
		// createRRSet
		Route("POST /v2/zones/test.example.com/my.test.example.com/TXT",
			servermock.JSONEncode(APIError{Message: "oops"}).WithStatusCode(http.StatusBadRequest)).
		Build(t)

	err := client.AddRRSet(t.Context(), "test.example.com", "my.test.example.com", testRecordContent, testTTL)
	require.EqualError(t, err, "400: oops")
}

func TestClient_AddRRSet_update(t *testing.T) {
	client := mockBuilder().
		// GetRRSet
		Route("GET /v2/zones/test.example.com/my.test.example.com/TXT",
			servermock.JSONEncode(RRSet{
				TTL:     testTTL,
				Records: []Records{{Content: []string{"foo"}}},
			})).
		// updateRRSet
		Route("PUT /v2/zones/test.example.com/my.test.example.com/TXT", nil,
			servermock.CheckRequestJSONBody(`{"ttl":10,"resource_records":[{"content":["acme"]},{"content":["foo"]}]}`)).
		Build(t)

	err := client.AddRRSet(t.Context(), "test.example.com", "my.test.example.com", testRecordContent, testTTL)
	require.NoError(t, err)
}

func TestClient_AddRRSet_update_error(t *testing.T) {
	client := mockBuilder().
		// GetRRSet
		Route("GET /v2/zones/test.example.com/my.test.example.com/TXT",
			servermock.JSONEncode(RRSet{
				TTL:     testTTL,
				Records: []Records{{Content: []string{"foo"}}},
			})).
		// updateRRSet
		Route("PUT /v2/zones/test.example.com/my.test.example.com/TXT",
			servermock.JSONEncode(APIError{Message: "oops"}).WithStatusCode(http.StatusBadRequest)).
		Build(t)

	err := client.AddRRSet(t.Context(), "test.example.com", "my.test.example.com", testRecordContent, testTTL)
	require.EqualError(t, err, "400: oops")
}
