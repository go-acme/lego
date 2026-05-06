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

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client, err := NewClient("abc123", "secret")
			if err != nil {
				return nil, err
			}

			client.BaseURL, _ = url.Parse(server.URL)
			client.HTTPClient = server.Client()

			return client, nil
		},
		servermock.CheckHeader().
			WithJSONHeaders().
			WithBasicAuth("abc123", "secret"),
	)
}

func TestClient_CreateRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /zones/ZONE-ID-3/versions/VERSION-ID-3/records",
			servermock.ResponseFromFixture("create_record.json"),
			servermock.CheckRequestJSONBodyFromFixture("create_record-request.json"),
		).
		Build(t)

	record := Record{
		Name: "_acme-challenge",
		Type: "TXT",
		TTL:  120,
		Records: []RecordTXT{
			{Data: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY"},
		},
	}

	result, err := client.CreateRecord(t.Context(), "ZONE-ID-3", "VERSION-ID-3", record)
	require.NoError(t, err)

	expected := &Record{
		ID:   "RECORD-ID-1",
		Name: "_acme-challenge",
		Type: "TXT",
		TTL:  120,
		Records: []RecordTXT{
			{Data: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY"},
		},
	}

	assert.Equal(t, expected, result)
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /zones/ZONE-ID-3/versions/VERSION-ID-3/records/RECORD-ID-1",
			servermock.ResponseFromFixture("delete_record.json"),
		).
		Build(t)

	result, err := client.DeleteRecord(t.Context(), "ZONE-ID-3", "VERSION-ID-3", "RECORD-ID-1")
	require.NoError(t, err)

	expected := &Record{
		ID:   "RECORD-ID-1",
		Name: "_acme-challenge",
		Type: "TXT",
		TTL:  120,
		Records: []RecordTXT{
			{Data: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY"},
		},
	}

	assert.Equal(t, expected, result)
}

func TestClient_ListZones(t *testing.T) {
	client := mockBuilder().
		Route("GET /zones",
			servermock.ResponseFromFixture("list_zones.json"),
		).
		Build(t)

	result, err := client.ListZones(t.Context())
	require.NoError(t, err)

	expected := []Zone{
		{ID: "ZONE-ID-1", Name: "example.net", DeletionProtection: true, CurrentVersionID: "VERSION-ID-1"},
		{ID: "ZONE-ID-2", Name: "example.org", DeletionProtection: true, CurrentVersionID: "VERSION-ID-2"},
		{ID: "ZONE-ID-3", Name: "example.com", DeletionProtection: true, CurrentVersionID: "VERSION-ID-3"},
	}

	assert.Equal(t, expected, result)
}

func TestClient_ListZones_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /zones",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized),
		).
		Build(t)

	_, err := client.ListZones(t.Context())
	require.EqualError(t, err, "401: Unauthorized")
}
