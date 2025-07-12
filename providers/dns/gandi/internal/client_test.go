package internal

import (
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/stubrouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *stubrouter.Builder[*Client] {
	return stubrouter.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client := NewClient("secret")
			client.BaseURL = server.URL
			client.HTTPClient = server.Client()

			return client, nil
		},
		stubrouter.CheckHeader().WithContentType("text/xml"),
	)
}

func TestClient_GetZoneID(t *testing.T) {
	client := mockBuilder().
		Route("POST /", stubrouter.ResponseFromFixture("get_zone_id.xml"),
			stubrouter.CheckRequestBodyFromFile("get_zone_id-request.xml").IgnoreWhitespace()).
		Build(t)

	zoneID, err := client.GetZoneID(t.Context(), "example.com")
	require.NoError(t, err)

	assert.Equal(t, 1, zoneID)
}

func TestClient_CloneZone(t *testing.T) {
	client := mockBuilder().
		Route("POST /", stubrouter.ResponseFromFixture("clone_zone.xml"),
			stubrouter.CheckRequestBodyFromFile("clone_zone-request.xml").IgnoreWhitespace()).
		Build(t)

	zoneID, err := client.CloneZone(t.Context(), 6, "foo")
	require.NoError(t, err)

	assert.Equal(t, 1, zoneID)
}

func TestClient_NewZoneVersion(t *testing.T) {
	client := mockBuilder().
		Route("POST /", stubrouter.ResponseFromFixture("new_zone_version.xml"),
			stubrouter.CheckRequestBodyFromFile("new_zone_version-request.xml").IgnoreWhitespace()).
		Build(t)

	zoneID, err := client.NewZoneVersion(t.Context(), 6)
	require.NoError(t, err)

	assert.Equal(t, 1, zoneID)
}

func TestClient_AddTXTRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /", stubrouter.ResponseFromFixture("empty.xml"),
			stubrouter.CheckRequestBodyFromFile("add_txt_record-request.xml").IgnoreWhitespace()).
		Build(t)

	err := client.AddTXTRecord(t.Context(), 1, 123, "foo", "content", 120)
	require.NoError(t, err)
}

func TestClient_SetZoneVersion(t *testing.T) {
	client := mockBuilder().
		Route("POST /", stubrouter.ResponseFromFixture("set_zone_version.xml"),
			stubrouter.CheckRequestBodyFromFile("set_zone_version-request.xml").IgnoreWhitespace()).
		Build(t)

	err := client.SetZoneVersion(t.Context(), 1, 123)
	require.NoError(t, err)
}

func TestClient_SetZone(t *testing.T) {
	client := mockBuilder().
		Route("POST /", stubrouter.ResponseFromFixture("set_zone.xml"),
			stubrouter.CheckRequestBodyFromFile("set_zone-request.xml").IgnoreWhitespace()).
		Build(t)

	err := client.SetZone(t.Context(), "example.com", 1)
	require.NoError(t, err)
}

func TestClient_DeleteZone(t *testing.T) {
	client := mockBuilder().
		Route("POST /", stubrouter.ResponseFromFixture("delete_zone.xml"),
			stubrouter.CheckRequestBodyFromFile("delete_zone-request.xml").IgnoreWhitespace()).
		Build(t)

	err := client.DeleteZone(t.Context(), 1)
	require.NoError(t, err)
}
