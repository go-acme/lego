package internal

import (
	"net/http/httptest"
	"testing"

	servermock2 "github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock2.Builder[*Client] {
	return servermock2.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client := NewClient("secret")
			client.BaseURL = server.URL
			client.HTTPClient = server.Client()

			return client, nil
		},
		servermock2.CheckHeader().WithContentType("text/xml"),
	)
}

func TestClient_GetZoneID(t *testing.T) {
	client := mockBuilder().
		Route("POST /", servermock2.ResponseFromFixture("get_zone_id.xml"),
			servermock2.CheckRequestBodyFromFixture("get_zone_id-request.xml").IgnoreWhitespace()).
		Build(t)

	zoneID, err := client.GetZoneID(t.Context(), "example.com")
	require.NoError(t, err)

	assert.Equal(t, 1, zoneID)
}

func TestClient_CloneZone(t *testing.T) {
	client := mockBuilder().
		Route("POST /", servermock2.ResponseFromFixture("clone_zone.xml"),
			servermock2.CheckRequestBodyFromFixture("clone_zone-request.xml").IgnoreWhitespace()).
		Build(t)

	zoneID, err := client.CloneZone(t.Context(), 6, "foo")
	require.NoError(t, err)

	assert.Equal(t, 1, zoneID)
}

func TestClient_NewZoneVersion(t *testing.T) {
	client := mockBuilder().
		Route("POST /", servermock2.ResponseFromFixture("new_zone_version.xml"),
			servermock2.CheckRequestBodyFromFixture("new_zone_version-request.xml").IgnoreWhitespace()).
		Build(t)

	zoneID, err := client.NewZoneVersion(t.Context(), 6)
	require.NoError(t, err)

	assert.Equal(t, 1, zoneID)
}

func TestClient_AddTXTRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /", servermock2.ResponseFromFixture("empty.xml"),
			servermock2.CheckRequestBodyFromFixture("add_txt_record-request.xml").IgnoreWhitespace()).
		Build(t)

	err := client.AddTXTRecord(t.Context(), 1, 123, "foo", "content", 120)
	require.NoError(t, err)
}

func TestClient_SetZoneVersion(t *testing.T) {
	client := mockBuilder().
		Route("POST /", servermock2.ResponseFromFixture("set_zone_version.xml"),
			servermock2.CheckRequestBodyFromFixture("set_zone_version-request.xml").IgnoreWhitespace()).
		Build(t)

	err := client.SetZoneVersion(t.Context(), 1, 123)
	require.NoError(t, err)
}

func TestClient_SetZone(t *testing.T) {
	client := mockBuilder().
		Route("POST /", servermock2.ResponseFromFixture("set_zone.xml"),
			servermock2.CheckRequestBodyFromFixture("set_zone-request.xml").IgnoreWhitespace()).
		Build(t)

	err := client.SetZone(t.Context(), "example.com", 1)
	require.NoError(t, err)
}

func TestClient_DeleteZone(t *testing.T) {
	client := mockBuilder().
		Route("POST /", servermock2.ResponseFromFixture("delete_zone.xml"),
			servermock2.CheckRequestBodyFromFixture("delete_zone-request.xml").IgnoreWhitespace()).
		Build(t)

	err := client.DeleteZone(t.Context(), 1)
	require.NoError(t, err)
}
