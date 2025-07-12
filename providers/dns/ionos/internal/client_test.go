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
		servermock.CheckHeader().With(APIKeyHeader, "secret"))
}

func TestClient_ListZones(t *testing.T) {
	client := mockBuilder().
		Route("GET /v1/zones",
			servermock.ResponseFromFixture("list_zones.json")).
		Build(t)

	zones, err := client.ListZones(t.Context())
	require.NoError(t, err)

	expected := []Zone{{
		ID:   "11af3414-ebba-11e9-8df5-66fbe8a334b4",
		Name: "test.com",
		Type: "NATIVE",
	}}

	assert.Equal(t, expected, zones)
}

func TestClient_ListZones_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /v1/zones",
			servermock.ResponseFromFixture("list_zones_error.json").
				WithStatusCode(http.StatusUnauthorized)).
		Build(t)

	zones, err := client.ListZones(t.Context())
	require.Error(t, err)

	assert.Nil(t, zones)

	var cErr *ClientError
	assert.ErrorAs(t, err, &cErr)
	assert.Equal(t, http.StatusUnauthorized, cErr.StatusCode)
}

func TestClient_GetRecords(t *testing.T) {
	client := mockBuilder().
		Route("GET /v1/zones/azone01",
			servermock.ResponseFromFixture("get_records.json")).
		Build(t)

	records, err := client.GetRecords(t.Context(), "azone01", nil)
	require.NoError(t, err)

	expected := []Record{{
		ID:      "22af3414-abbe-9e11-5df5-66fbe8e334b4",
		Name:    "string",
		Content: "string",
		Type:    "A",
	}}

	assert.Equal(t, expected, records)
}

func TestClient_GetRecords_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /v1/zones/azone01",
			servermock.ResponseFromFixture("get_records_error.json").
				WithStatusCode(http.StatusUnauthorized)).
		Build(t)

	records, err := client.GetRecords(t.Context(), "azone01", nil)
	require.Error(t, err)

	assert.Nil(t, records)

	var cErr *ClientError
	assert.ErrorAs(t, err, &cErr)
	assert.Equal(t, http.StatusUnauthorized, cErr.StatusCode)
}

func TestClient_RemoveRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /v1/zones/azone01/records/arecord01", nil).
		Build(t)

	err := client.RemoveRecord(t.Context(), "azone01", "arecord01")
	require.NoError(t, err)
}

func TestClient_RemoveRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /v1/zones/azone01/records/arecord01",
			servermock.ResponseFromFixture("remove_record_error.json").
				WithStatusCode(http.StatusInternalServerError)).
		Build(t)

	err := client.RemoveRecord(t.Context(), "azone01", "arecord01")
	require.Error(t, err)

	var cErr *ClientError
	assert.ErrorAs(t, err, &cErr)
	assert.Equal(t, http.StatusInternalServerError, cErr.StatusCode)
}

func TestClient_ReplaceRecords(t *testing.T) {
	client := mockBuilder().
		Route("PATCH /v1/zones/azone01", nil).
		Build(t)

	records := []Record{{
		ID:      "22af3414-abbe-9e11-5df5-66fbe8e334b4",
		Name:    "string",
		Content: "string",
		Type:    "A",
	}}

	err := client.ReplaceRecords(t.Context(), "azone01", records)
	require.NoError(t, err)
}

func TestClient_ReplaceRecords_error(t *testing.T) {
	client := mockBuilder().
		Route("PATCH /v1/zones/azone01",
			servermock.ResponseFromFixture("replace_records_error.json").
				WithStatusCode(http.StatusBadRequest)).
		Build(t)

	records := []Record{{
		ID:      "22af3414-abbe-9e11-5df5-66fbe8e334b4",
		Name:    "string",
		Content: "string",
		Type:    "A",
	}}

	err := client.ReplaceRecords(t.Context(), "azone01", records)
	require.Error(t, err)

	var cErr *ClientError
	assert.ErrorAs(t, err, &cErr)
	assert.Equal(t, http.StatusBadRequest, cErr.StatusCode)
}
