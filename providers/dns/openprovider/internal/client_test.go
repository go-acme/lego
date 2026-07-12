package internal

import (
	"context"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder(
		func(server *httptest.Server) (*Client, error) {
			client, err := NewClient("user", "secret")
			if err != nil {
				return nil, err
			}

			client.HTTPClient = server.Client()
			client.BaseURL, _ = url.Parse(server.URL)

			return client, nil
		},
		servermock.CheckHeader().
			WithJSONHeaders(),
	)
}

func TestClient_ListZone(t *testing.T) {
	client := mockBuilder().
		Route("GET /dns/zones",
			servermock.ResponseFromFixture("list_zones.json"),
			servermock.CheckQueryParameter().Strict().
				With("name_pattern", "example.com"),
		).
		Build(t)

	zones, err := client.ListZones(context.Background(), &ZonesRequest{NamePattern: "example.com"})
	require.NoError(t, err)

	expected := []Zone{{
		ID:   12345,
		Name: "example.com",
		Records: []Record{{
			Name:  "example.com",
			TTL:   900,
			Type:  "CNAME",
			Value: "abcd.example.com",
		}},
		ResellerID: 123456,
		Type:       "master",
	}}

	assert.Equal(t, expected, zones)
}

func TestClient_UpdateZone(t *testing.T) {
	client := mockBuilder().
		Route("PUT /dns/zones/example.com",
			servermock.ResponseFromFixture("update_zone.json"),
			servermock.CheckRequestJSONBodyFromFixture("update_zone-request.json"),
		).
		Build(t)

	payload := ZoneAction{
		ID:   12345,
		Name: "example.com",
		Records: RecordAction{
			Add: []Record{
				{
					Name:  "_acme-challenge",
					TTL:   600,
					Type:  "TXT",
					Value: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
				},
			},
		},
	}

	err := client.UpdateZone(context.Background(), "example.com", payload)
	require.NoError(t, err)
}
