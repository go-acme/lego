package cpanel

import (
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/go-acme/lego/v4/providers/dns/cpanel/internal/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client, err := NewClient(server.URL, "user", "secret")
			if err != nil {
				return nil, err
			}

			client.HTTPClient = server.Client()

			return client, nil
		},
		servermock.CheckHeader().WithJSONHeaders().
			WithAuthorization("cpanel user:secret"))
}

func TestClient_FetchZoneInformation(t *testing.T) {
	client := mockBuilder().
		Route("GET /execute/DNS/parse_zone",
			servermock.ResponseFromFixture("zone-info.json"),
			servermock.CheckQueryParameter().Strict().
				With("zone", "example.com")).
		Build(t)

	zoneInfo, err := client.FetchZoneInformation(t.Context(), "example.com")
	require.NoError(t, err)

	expected := []shared.ZoneRecord{{
		LineIndex:  22,
		Type:       "record",
		DataB64:    []string{"dGV4YXMuY29tLg=="},
		DNameB64:   "dGV4YXMuY29tLg==",
		RecordType: "MX",
		TTL:        14400,
	}}

	assert.Equal(t, expected, zoneInfo)
}

func TestClient_FetchZoneInformation_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /execute/DNS/parse_zone",
			servermock.ResponseFromFixture("zone-info_error.json")).
		Build(t)

	zoneInfo, err := client.FetchZoneInformation(t.Context(), "example.com")
	require.EqualError(t, err, "error(0): You do not control a DNS zone named example.com.: a, b, c")

	assert.Nil(t, zoneInfo)
}

func TestClient_AddRecord(t *testing.T) {
	client := mockBuilder().
		Route("GET /execute/DNS/mass_edit_zone",
			servermock.ResponseFromFixture("update-zone.json"),
			servermock.CheckQueryParameter().Strict().
				With("zone", "example.com").
				With("add", `{"dname":"example","ttl":14400,"record_type":"TXT","data":["string1","string2"]}`).
				With("serial", "123456").
				With("zone", "example.com")).
		Build(t)

	record := shared.Record{
		DName:      "example",
		TTL:        14400,
		RecordType: "TXT",
		Data:       []string{"string1", "string2"},
	}

	zoneSerial, err := client.AddRecord(t.Context(), 123456, "example.com", record)
	require.NoError(t, err)

	expected := &shared.ZoneSerial{NewSerial: "2021031903"}

	assert.Equal(t, expected, zoneSerial)
}

func TestClient_AddRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /execute/DNS/mass_edit_zone",
			servermock.ResponseFromFixture("update-zone_error.json")).
		Build(t)

	record := shared.Record{
		DName:      "example",
		TTL:        14400,
		RecordType: "TXT",
		Data:       []string{"string1", "string2"},
	}

	zoneSerial, err := client.AddRecord(t.Context(), 123456, "example.com", record)
	require.Error(t, err)

	assert.Nil(t, zoneSerial)
}

func TestClient_EditRecord(t *testing.T) {
	client := mockBuilder().
		Route("GET /execute/DNS/mass_edit_zone",
			servermock.ResponseFromFixture("update-zone.json"),
			servermock.CheckQueryParameter().Strict().
				With("edit", `{"dname":"example","ttl":14400,"record_type":"TXT","data":["string1","string2"],"line_index":9}`).
				With("serial", "123456").
				With("zone", "example.com")).
		Build(t)

	record := shared.Record{
		LineIndex:  9,
		DName:      "example",
		TTL:        14400,
		RecordType: "TXT",
		Data:       []string{"string1", "string2"},
	}

	zoneSerial, err := client.EditRecord(t.Context(), 123456, "example.com", record)
	require.NoError(t, err)

	expected := &shared.ZoneSerial{NewSerial: "2021031903"}

	assert.Equal(t, expected, zoneSerial)
}

func TestClient_EditRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /execute/DNS/mass_edit_zone",
			servermock.ResponseFromFixture("update-zone_error.json")).
		Build(t)

	record := shared.Record{
		LineIndex:  9,
		DName:      "example",
		TTL:        14400,
		RecordType: "TXT",
		Data:       []string{"string1", "string2"},
	}

	zoneSerial, err := client.EditRecord(t.Context(), 123456, "example.com", record)
	require.Error(t, err)

	assert.Nil(t, zoneSerial)
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("GET /execute/DNS/mass_edit_zone",
			servermock.ResponseFromFixture("update-zone.json"),
			servermock.CheckQueryParameter().Strict().
				With("remove", "0").
				With("serial", "123456").
				With("zone", "example.com")).
		Build(t)

	zoneSerial, err := client.DeleteRecord(t.Context(), 123456, "example.com", 0)
	require.NoError(t, err)

	expected := &shared.ZoneSerial{NewSerial: "2021031903"}

	assert.Equal(t, expected, zoneSerial)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /execute/DNS/mass_edit_zone",
			servermock.ResponseFromFixture("update-zone_error.json")).
		Build(t)

	zoneSerial, err := client.DeleteRecord(t.Context(), 123456, "example.com", 0)
	require.Error(t, err)

	assert.Nil(t, zoneSerial)
}
