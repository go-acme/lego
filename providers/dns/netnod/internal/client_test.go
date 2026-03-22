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
		servermock.CheckHeader().
			WithJSONHeaders().
			WithAuthorization("Token secret"),
	)
}

func TestClient_GetZones(t *testing.T) {
	client := mockBuilder().
		Route("GET /api/v1/zones",
			servermock.ResponseFromFixture("zones.json"),
			servermock.CheckQueryParameter().Strict().
				With("offset", "1").
				With("limit", "100").
				With("endcustomer", "customer123"),
		).
		Build(t)

	pager := Pager{
		Offset: 1,
		Limit:  100,
	}

	result, err := client.GetZones(t.Context(), "customer123", pager)
	require.NoError(t, err)

	expected := &PagedResponse[[]Zone]{
		Data: []Zone{
			{
				ID:             "example.com.",
				Name:           "example.com.",
				NotifiedSerial: 2025110401,
			},
			{
				ID:             "test.com.",
				Name:           "test.com.",
				NotifiedSerial: 2025110402,
			},
		},
		Offset: 0,
		Limit:  50,
		Total:  2,
	}

	assert.Equal(t, expected, result)
}

func TestClient_PartialZoneUpdate(t *testing.T) {
	client := mockBuilder().
		Route("PATCH /api/v1/zones/example.com.",
			servermock.Noop().
				WithStatusCode(http.StatusNoContent),
			servermock.CheckRequestJSONBodyFromFixture("partial_zone_update-request.json"),
		).
		Build(t)

	rrSets := []RRSet{
		{
			Name:       "subdomain.example.com.",
			Type:       "A",
			ChangeType: ChangeTypeDelete,
		},
		{
			Name:       "newhost.example.com.",
			Type:       "A",
			ChangeType: ChangeTypeReplace,
			Records: []Record{
				{
					Content: "1.2.3.5",
				},
			},
		},
	}

	err := client.PartialZoneUpdate(t.Context(), "example.com.", rrSets)
	require.NoError(t, err)
}
