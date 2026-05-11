package internal

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client, err := NewClient()
			if err != nil {
				return nil, err
			}

			client.BaseURL, _ = url.Parse(server.URL)
			client.HTTPClient = server.Client()

			return client, nil
		},
	)
}

func TestClient_GetDomains(t *testing.T) {
	client := mockBuilder().
		Route("GET /groups/grp1/domains",
			servermock.ResponseFromFixture("domains_get.json"),
			servermock.CheckQueryParameter().Strict().
				With("filter[name]", "*.example.com"),
		).
		Build(t)

	result, err := client.GetDomains(t.Context(), "grp1", "*.example.com", nil)
	require.NoError(t, err)

	expected := &DomainsResponse{
		BasePaginatedResponse: BasePaginatedResponse[string, string, []Domain]{
			BaseResponse: BaseResponse[string, string, []Domain]{
				Errors:   []string{"string"},
				Messages: []string{"string"},
				Data: []Domain{{
					ID:          "e32f37b7-a251-4981-9613-4d3dac6c4532",
					Name:        "example.com",
					IDNName:     "example.com",
					TLD:         "string",
					CountryCode: "string",
					RegistryLock: &RegistryLock{
						Enabled:   true,
						ExpiresAt: time.Date(2026, 3, 3, 14, 44, 40, 12000000, time.UTC),
					},
					Flagged: true,
					ActiveZone: &ActiveZone{
						ID:                  "62b873d3-a31c-4921-a309-548810913c4f",
						DefaultRecordTTL:    0,
						Signed:              true,
						ResourceRecordCount: 0,
						Secondary:           &SecondaryZone{PrimaryIP: "string", OtherIPs: []string{"string"}},
						Networks:            []string{"string"},
					},
					ExternalComments: "string",
					Tags:             []Tag{{ID: 0, Name: "string", Type: "string"}},
				}},
				StatusCode: 200,
			},
			Pagination: Pagination{
				First: "string",
				Last:  "string",
				Prev:  "string",
				Next:  "string",
			},
		},
		NameResults: []NameResult{{Key: "string", Count: 0}},
		TermResults: []TermResult{{Key: "string", Count: 0}},
	}

	assert.Equal(t, expected, result)
}

func TestClient_GetDomains_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /groups/grp1/domains",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized),
			servermock.CheckQueryParameter().Strict().
				With("filter[name]", "*.example.com"),
		).
		Build(t)

	_, err := client.GetDomains(t.Context(), "grp1", "*.example.com", nil)
	require.EqualError(t, err, "401: incorrect_login_details: Incorrect login details: (string)")
}

func TestClient_CreateRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /groups/grp1/zones/z1/records",
			servermock.ResponseFromFixture("record_create.json"),
			servermock.CheckHeader().
				WithContentTypeFromURLEncoded(),
			servermock.CheckForm().
				With("name", "_acme-challenge").
				With("ttl", "120").
				With("value", "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY").
				With("type", "TXT"),
		).
		Build(t)

	record := RecordRequest{
		Name:    "_acme-challenge",
		Type:    "TXT",
		TTL:     120,
		Content: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
	}

	recordID, err := client.CreateRecord(t.Context(), "grp1", "z1", record)
	require.NoError(t, err)

	assert.Equal(t, "8a746001-d319-4583-bfb6-ae8aacc628aa", recordID)
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /groups/grp1/zones/z1/records/r1",
			servermock.ResponseFromFixture("record_delete.json"),
		).
		Build(t)

	err := client.DeleteRecord(t.Context(), "grp1", "z1", "r1")
	require.NoError(t, err)
}
