package internal

import (
	"context"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder(
		func(server *httptest.Server) (*Client, error) {
			client := NewClient("user", "secret", "example.com", "test")
			client.HTTPClient = server.Client()
			client.baseURL, _ = url.Parse(server.URL)

			return client, nil
		},
		servermock.CheckHeader().WithJSONHeaders(),
	)
}

func TestClient_GetZoneID(t *testing.T) {
	client := mockBuilder().
		Route("GET /zones",
			servermock.ResponseFromFixture("zones_GET.json"),
			servermock.CheckQueryParameter().Strict().
				With("name", "example.com.")).
		Build(t)

	zoneID, err := client.GetZoneID(context.Background(), "example.com.")
	require.NoError(t, err)

	assert.Equal(t, "123123", zoneID)
}

func TestClient_GetZoneID_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /zones",
			servermock.ResponseFromFixture("zones_GET_empty.json"),
			servermock.CheckQueryParameter().Strict().
				With("name", "example.com.")).
		Build(t)

	_, err := client.GetZoneID(context.Background(), "example.com.")
	require.EqualError(t, err, "zone example.com. not found")
}

func TestClient_GetRecordSetID(t *testing.T) {
	client := mockBuilder().
		Route("GET /zones/123123/recordsets",
			servermock.ResponseFromFixture("zones-recordsets_GET.json"),
			servermock.CheckQueryParameter().Strict().
				With("name", "example.com.").
				With("type", "TXT"),
		).
		Build(t)

	recordSetID, err := client.GetRecordSetID(context.Background(), "123123", "example.com.")
	require.NoError(t, err)

	assert.Equal(t, "321321", recordSetID)
}

func TestClient_GetRecordSetID_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /zones/123123/recordsets",
			servermock.ResponseFromFixture("zones-recordsets_GET_empty.json"),
			servermock.CheckQueryParameter().Strict().
				With("name", "example.com.").
				With("type", "TXT"),
		).
		Build(t)

	_, err := client.GetRecordSetID(context.Background(), "123123", "example.com.")
	require.EqualError(t, err, "record not found")
}

func TestClient_CreateRecordSet(t *testing.T) {
	client := mockBuilder().
		Route("POST /zones/123123/recordsets",
			servermock.ResponseFromFixture("zones-recordsets_POST.json")).
		Build(t)

	rs := RecordSets{
		Name:        "_acme-challenge.example.com.",
		Description: "Added TXT record for ACME dns-01 challenge using lego client",
		Type:        "TXT",
		TTL:         300,
		Records:     []string{strconv.Quote("w6uP8Tcg6K2QR905Rms8iXTlksL6OD1KOWBxTK7wxPI")},
	}
	err := client.CreateRecordSet(context.Background(), "123123", rs)
	require.NoError(t, err)
}

func TestClient_DeleteRecordSet(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /zones/123123/recordsets/321321",
			servermock.ResponseFromFixture("zones-recordsets_DELETE.json")).
		Build(t)

	err := client.DeleteRecordSet(context.Background(), "123123", "321321")
	require.NoError(t, err)
}
