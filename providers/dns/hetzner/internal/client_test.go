package internal

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockBuilder(apiKey string) *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client := NewClient(apiKey)
			client.baseURL, _ = url.Parse(server.URL)
			client.HTTPClient = server.Client()

			return client, nil
		},
		servermock.CheckHeader().WithJSONHeaders().
			With(authHeader, apiKey))
}

func TestClient_GetTxtRecord(t *testing.T) {
	const zoneID = "zoneA"

	client := mockBuilder("myKeyA").
		Route("GET /api/v1/records", servermock.ResponseFromFixture("get_txt_record.json"),
			servermock.CheckQueryParameter().Strict().
				With("zone_id", zoneID)).
		Build(t)

	record, err := client.GetTxtRecord(t.Context(), "test1", "txttxttxt", zoneID)
	require.NoError(t, err)

	expected := &DNSRecord{
		ID:       "1b",
		Name:     "test1",
		Type:     "TXT",
		Value:    "txttxttxt",
		Priority: 0,
		TTL:      600,
		ZoneID:   "zoneA",
	}

	assert.Equal(t, expected, record)
}

func TestClient_CreateRecord(t *testing.T) {
	const zoneID = "zoneA"

	client := mockBuilder("myKeyB").
		Route("POST /api/v1/records", servermock.ResponseFromFixture("create_txt_record.json"),
			servermock.CheckRequestJSONBodyFromFile("create_txt_record-request.json")).
		Build(t)

	record := DNSRecord{
		Name:   "test",
		Type:   "TXT",
		Value:  "txttxttxt",
		TTL:    600,
		ZoneID: zoneID,
	}

	err := client.CreateRecord(t.Context(), record)
	require.NoError(t, err)
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder("myKeyC").
		Route("DELETE /api/v1/records/recordID", nil).
		Build(t)

	err := client.DeleteRecord(t.Context(), "recordID")
	require.NoError(t, err)
}

func TestClient_GetZoneID(t *testing.T) {
	client := mockBuilder("myKeyD").
		Route("GET /api/v1/zones", servermock.ResponseFromFixture("get_zone_id.json")).
		Build(t)

	zoneID, err := client.GetZoneID(t.Context(), "example.com")
	require.NoError(t, err)

	assert.Equal(t, "zoneA", zoneID)
}
