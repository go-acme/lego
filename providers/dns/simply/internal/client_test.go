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
			client, err := NewClient("accountname", "apikey")
			if err != nil {
				return nil, err
			}

			client.baseURL, _ = url.Parse(server.URL)
			client.HTTPClient = server.Client()

			return client, nil
		},
		servermock.CheckHeader().WithJSONHeaders().
			WithBasicAuth("accountname", "apikey"))
}

func TestClient_GetRecords(t *testing.T) {
	client := mockBuilder().
		Route("GET /my/products/azone01/dns/records",
			servermock.ResponseFromFixture("get_records.json")).
		Build(t)

	records, err := client.GetRecords(t.Context(), "azone01")
	require.NoError(t, err)

	expected := []Record{
		{
			ID:       1,
			Name:     "@",
			TTL:      3600,
			Data:     "ns1.simply.com",
			Type:     "NS",
			Priority: 0,
		},
		{
			ID:       2,
			Name:     "@",
			TTL:      3600,
			Data:     "ns2.simply.com",
			Type:     "NS",
			Priority: 0,
		},
		{
			ID:       3,
			Name:     "@",
			TTL:      3600,
			Data:     "ns3.simply.com",
			Type:     "NS",
			Priority: 0,
		},
		{
			ID:       4,
			Name:     "@",
			TTL:      3600,
			Data:     "ns4.simply.com",
			Type:     "NS",
			Priority: 0,
		},
	}

	assert.Equal(t, expected, records)
}

func TestClient_GetRecords_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /my/products/azone01/dns/records",
			servermock.ResponseFromFixture("bad_auth_error.json").
				WithStatusCode(http.StatusBadRequest)).
		Build(t)

	records, err := client.GetRecords(t.Context(), "azone01")
	require.Error(t, err)

	assert.Nil(t, records)
}

func TestClient_AddRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /my/products/azone01/dns/records",
			servermock.ResponseFromFixture("add_record.json")).
		Build(t)

	record := Record{
		Name:     "arecord01",
		Data:     "content",
		Type:     "TXT",
		TTL:      120,
		Priority: 0,
	}

	recordID, err := client.AddRecord(t.Context(), "azone01", record)
	require.NoError(t, err)

	assert.EqualValues(t, 123456789, recordID)
}

func TestClient_AddRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /my/products/azone01/dns/records",
			servermock.ResponseFromFixture("bad_zone_error.json").
				WithStatusCode(http.StatusNotFound)).
		Build(t)

	record := Record{
		Name:     "arecord01",
		Data:     "content",
		Type:     "TXT",
		TTL:      120,
		Priority: 0,
	}

	recordID, err := client.AddRecord(t.Context(), "azone01", record)
	require.Error(t, err)

	assert.Zero(t, recordID)
}

func TestClient_EditRecord(t *testing.T) {
	client := mockBuilder().
		Route("PUT /my/products/azone01/dns/records/123456789",
			servermock.ResponseFromFixture("success.json")).
		Build(t)

	record := Record{
		Name:     "arecord01",
		Data:     "content",
		Type:     "TXT",
		TTL:      120,
		Priority: 0,
	}

	err := client.EditRecord(t.Context(), "azone01", 123456789, record)
	require.NoError(t, err)
}

func TestClient_EditRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("PUT /my/products/azone01/dns/records/123456789",
			servermock.ResponseFromFixture("invalid_record_id.json").
				WithStatusCode(http.StatusNotFound)).
		Build(t)

	record := Record{
		Name:     "arecord01",
		Data:     "content",
		Type:     "TXT",
		TTL:      120,
		Priority: 0,
	}

	err := client.EditRecord(t.Context(), "azone01", 123456789, record)
	require.Error(t, err)
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /my/products/azone01/dns/records/123456789",
			servermock.ResponseFromFixture("success.json")).
		Build(t)

	err := client.DeleteRecord(t.Context(), "azone01", 123456789)
	require.NoError(t, err)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /my/products/azone01/dns/records/123456789",
			servermock.ResponseFromFixture("invalid_record_id.json").
				WithStatusCode(http.StatusNotFound)).
		Build(t)

	err := client.DeleteRecord(t.Context(), "azone01", 123456789)
	require.Error(t, err)
}
