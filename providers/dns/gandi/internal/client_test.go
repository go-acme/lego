package internal

import (
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client := NewClient("secret")
			client.BaseURL = server.URL
			client.HTTPClient = server.Client()

			return client, nil
		},
		servermock.CheckHeader().WithContentType("text/xml"),
	)
}

func TestClient_GetZoneID(t *testing.T) {
	client := mockBuilder().
		Route("POST /", servermock.ResponseFromFixture("get_zone_id.xml"),
			servermock.CheckRequestBodyFromFixture("get_zone_id-request.xml").IgnoreWhitespace()).
		Build(t)

	zoneID, err := client.GetZoneID(t.Context(), "example.com")
	require.NoError(t, err)

	assert.Equal(t, 1, zoneID)
}

func TestClient_CloneZone(t *testing.T) {
	client := mockBuilder().
		Route("POST /", servermock.ResponseFromFixture("clone_zone.xml"),
			servermock.CheckRequestBodyFromFixture("clone_zone-request.xml").IgnoreWhitespace()).
		Build(t)

	zoneID, err := client.CloneZone(t.Context(), 6, "foo")
	require.NoError(t, err)

	assert.Equal(t, 1, zoneID)
}

func TestClient_NewZoneVersion(t *testing.T) {
	client := mockBuilder().
		Route("POST /", servermock.ResponseFromFixture("new_zone_version.xml"),
			servermock.CheckRequestBodyFromFixture("new_zone_version-request.xml").IgnoreWhitespace()).
		Build(t)

	zoneID, err := client.NewZoneVersion(t.Context(), 6)
	require.NoError(t, err)

	assert.Equal(t, 1, zoneID)
}

func TestClient_AddTXTRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /", servermock.ResponseFromFixture("empty.xml"),
			servermock.CheckRequestBodyFromFixture("add_txt_record-request.xml").IgnoreWhitespace()).
		Build(t)

	err := client.AddTXTRecord(t.Context(), 1, 123, "foo", "content", 120)
	require.NoError(t, err)
}

func TestClient_SetZoneVersion(t *testing.T) {
	client := mockBuilder().
		Route("POST /", servermock.ResponseFromFixture("set_zone_version.xml"),
			servermock.CheckRequestBodyFromFixture("set_zone_version-request.xml").IgnoreWhitespace()).
		Build(t)

	err := client.SetZoneVersion(t.Context(), 1, 123)
	require.NoError(t, err)
}

func TestClient_SetZone(t *testing.T) {
	client := mockBuilder().
		Route("POST /", servermock.ResponseFromFixture("set_zone.xml"),
			servermock.CheckRequestBodyFromFixture("set_zone-request.xml").IgnoreWhitespace()).
		Build(t)

	err := client.SetZone(t.Context(), "example.com", 1)
	require.NoError(t, err)
}

func TestClient_DeleteZone(t *testing.T) {
	client := mockBuilder().
		Route("POST /", servermock.ResponseFromFixture("delete_zone.xml"),
			servermock.CheckRequestBodyFromFixture("delete_zone-request.xml").IgnoreWhitespace()).
		Build(t)

	err := client.DeleteZone(t.Context(), 1)
	require.NoError(t, err)
}
