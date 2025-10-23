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
			WithAuthorization("Token secret"),
	)
}

func TestClient_CreateRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /api/clouddns/v1/zone.json/example.com/records",
			servermock.ResponseFromFixture("create_record.json"),
			servermock.CheckHeader().
				WithContentType("application/json; charset=utf-8"),
			servermock.CheckRequestJSONBodyFromFixture("create_record-request.json")).
		Build(t)

	record := Record{
		Name:  "_acme-challenge",
		RData: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		TTL:   300,
		Type:  "TXT",
	}

	zone, err := client.CreateRecord(t.Context(), "example.com", record)
	require.NoError(t, err)

	expected := &Zone{
		Name:     "example.com",
		TTL:      86400,
		ZoneName: "example.com",
		Revisions: []Revision{{
			Identifier: "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
			Records: []Record{{
				Identifier: "12345678-1234-1234-1234-123456789abc",
				Name:       "_acme-challenge",
				RData:      "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
				TTL:        300,
				Type:       "TXT",
			}},
			State: "deployed",
		}},
	}

	assert.Equal(t, expected, zone)
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /api/clouddns/v1/zone.json/example.com/records/12345678-1234-1234-1234-123456789abc",
			servermock.Noop()).
		Build(t)

	err := client.DeleteRecord(t.Context(), "example.com", "12345678-1234-1234-1234-123456789abc")
	require.NoError(t, err)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /api/clouddns/v1/zone.json/example.com/records/12345678-1234-1234-1234-123456789abc",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized)).
		Build(t)

	err := client.DeleteRecord(t.Context(), "example.com", "12345678-1234-1234-1234-123456789abc")
	require.EqualError(t, err, "401: Unauthorized")
}

func TestClient_GetZone(t *testing.T) {
	client := mockBuilder().
		Route("GET /api/clouddns/v1/zone.json/example.com",
			servermock.ResponseFromFixture("get_zone.json")).
		Build(t)

	zone, err := client.GetZone(t.Context(), "example.com")
	require.NoError(t, err)

	expected := &Zone{
		Identifier: "fdb355ffd07c48aba3d4f6bf6a116296",
		Name:       "example.com",
		TTL:        3600,
		ZoneName:   "",
		Revisions: []Revision{{
			Identifier: "eeed7e08-f1ad-442b-9e75-369a0958c7d8",
			Records: []Record{
				{
					Identifier: "5ced498b-c89d-4487-824d-c03ded84f849",
					Immutable:  true,
					Name:       "@",
					RData:      "acns02.xaas.systems.",
					Region:     "9a1609af9dae4ce1a4ef63f51d305321",
					TTL:        3600,
					Type:       "NS",
				},
				{
					Identifier: "12345678-1234-1234-1234-123456789abc",
					Immutable:  false,
					Name:       "_acme-challenge",
					RData:      "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
					Region:     "",
					TTL:        300,
					Type:       "TXT",
				},
			},
			State: "active",
		}},
	}

	assert.Equal(t, expected, zone)
}
