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
			client, err := NewClient("secret")
			if err != nil {
				return nil, err
			}

			client.BaseURL, _ = url.Parse(server.URL)
			client.HTTPClient = server.Client()

			return client, nil
		},
		servermock.CheckHeader().WithJSONHeaders(),
		servermock.CheckHeader().With(AuthorizationHeader, "Bearer secret"))
}

func TestClient_FindZone(t *testing.T) {
	client := mockBuilder().
		Route("GET /zones",
			servermock.ResponseFromFixture("list_zones.json")).
		Build(t)

	zone, err := client.FindZone(t.Context(), "test.com")
	require.NoError(t, err)

	require.NotNil(t, zone)
	assert.Equal(t, "11af3414-ebba-11e9-8df5-66fbe8a334b4", zone.ID)
	assert.Equal(t, "test.com", zone.Properties.ZoneName)
}

func TestClient_FindZone_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /zones",
			servermock.ResponseFromFixture("list_zones_error.json").
				WithStatusCode(http.StatusUnauthorized)).
		Build(t)

	zone, err := client.FindZone(t.Context(), "test.com")
	require.Error(t, err)
	assert.Nil(t, zone)
}

func TestClient_GetRecords(t *testing.T) {
	client := mockBuilder().
		Route("GET /zones/azone01/records",
			servermock.ResponseFromFixture("get_records.json")).
		Build(t)

	records, err := client.GetRecords(t.Context(), "azone01")
	require.NoError(t, err)

	require.Len(t, records, 1)
	rec := records[0]
	assert.Equal(t, "22af3414-abbe-9e11-5df5-66fbe8e334b4", rec.ID)
	assert.Equal(t, "string", rec.Properties.Name)
	assert.Equal(t, "string", rec.Properties.Content)
	assert.Equal(t, "A", rec.Properties.Type)
}

func TestClient_GetRecords_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /zones/azone01/records",
			servermock.ResponseFromFixture("get_records_error.json").
				WithStatusCode(http.StatusUnauthorized)).
		Build(t)

	records, err := client.GetRecords(t.Context(), "azone01")
	require.Error(t, err)
	assert.Nil(t, records)
}

func TestClient_RemoveRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /zones/azone01/records/arecord01", nil).
		Build(t)

	err := client.RemoveRecord(t.Context(), "azone01", "arecord01")
	require.NoError(t, err)
}

func TestClient_RemoveRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /zones/azone01/records/arecord01",
			servermock.ResponseFromFixture("remove_record_error.json").
				WithStatusCode(http.StatusInternalServerError)).
		Build(t)

	err := client.RemoveRecord(t.Context(), "azone01", "arecord01")
	require.Error(t, err)
}

func TestClient_CreateRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /zones/azone01/records", nil).
		Build(t)

	rec := Record{Properties: RecordProperties{Name: "_acme-challenge", Type: "TXT", Content: "val", TTL: 300, Enabled: true}}
	err := client.CreateRecord(t.Context(), "azone01", rec)
	require.NoError(t, err)
}

func TestClient_CreateRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /zones/azone01/records",
			servermock.ResponseFromFixture("remove_record_error.json").
				WithStatusCode(http.StatusBadRequest)).
		Build(t)

	rec := Record{Properties: RecordProperties{Name: "_acme-challenge", Type: "TXT", Content: "val", TTL: 300, Enabled: true}}
	err := client.CreateRecord(t.Context(), "azone01", rec)
	require.Error(t, err)
}
