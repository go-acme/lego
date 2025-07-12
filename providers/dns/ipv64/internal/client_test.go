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

const testAPIKey = "secret"

func setupClient(server *httptest.Server) (*Client, error) {
	client := NewClient(OAuthStaticAccessToken(server.Client(), testAPIKey))
	client.HTTPClient = server.Client()
	client.baseURL, _ = url.Parse(server.URL)

	return client, nil
}

func TestClient_GetDomains(t *testing.T) {
	client := servermock.NewBuilder[*Client](setupClient).
		Route("GET /api",
			servermock.ResponseFromFixture("get_domains.json"),
			servermock.CheckQueryParameter().Strict().
				With("get_domains", "")).
		Build(t)

	domains, err := client.GetDomains(t.Context())
	require.NoError(t, err)

	expected := &Domains{
		APIResponse: APIResponse{
			Status: "200 OK",
			Info:   "success",
		},
		APICall: "get_domains",
		Subdomains: map[string]Subdomain{
			"lego.home64.net": {
				Updates:          0,
				Wildcard:         1,
				DomainUpdateHash: "Dr4l6jFVgkXITqZPEyMHLNsGAfwoSu9v",
				Records: []Record{
					{
						RecordID:   50665,
						Content:    "2606:2800:220:1:248:1893:25c8:1946",
						TTL:        60,
						Type:       "AAAA",
						Prefix:     "",
						LastUpdate: "2023-07-19 13:18:59",
						RecordKey:  "MTA0YzdmMWVjYTFiNDBmZjYwMTU0OGUy",
					},
				},
			},
			"lego.ipv64.net": {
				Updates:          0,
				Wildcard:         1,
				DomainUpdateHash: "Dr4l6jFVgkXITqZPEyMHLNsGAfwoSu9v",
				Records: []Record{
					{
						RecordID:   50664,
						Content:    "2606:2800:220:1:248:1893:25c8:1946",
						TTL:        60,
						Type:       "AAAA",
						Prefix:     "",
						LastUpdate: "2023-07-19 13:18:59",
						RecordKey:  "ZDMxOWUxMjZjOTk5MmQ3N2M3ODc4NjJj",
					},
				},
			},
		},
	}

	assert.Equal(t, expected, domains)
}

func TestClient_GetDomains_error(t *testing.T) {
	client := servermock.NewBuilder[*Client](setupClient).
		Route("GET /api",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized)).
		Build(t)

	domains, err := client.GetDomains(t.Context())
	require.Error(t, err)

	require.Nil(t, domains)
}

func TestClient_AddRecord(t *testing.T) {
	client := servermock.NewBuilder[*Client](setupClient,
		servermock.CheckHeader().WithContentTypeFromURLEncoded()).
		Route("POST /api",
			servermock.ResponseFromFixture("add_record.json").
				WithStatusCode(http.StatusCreated),
			servermock.CheckForm().Strict().
				With("add_record", "lego.ipv64.net").
				With("content", "value").
				With("praefix", "_acme-challenge").
				With("type", "TXT"),
		).
		Build(t)

	err := client.AddRecord(t.Context(), "lego.ipv64.net", "_acme-challenge", "TXT", "value")
	require.NoError(t, err)
}

func TestClient_AddRecord_error(t *testing.T) {
	client := servermock.NewBuilder[*Client](setupClient).
		Route("POST /api",
			servermock.ResponseFromFixture("add_record-error.json").
				WithStatusCode(http.StatusBadRequest)).
		Build(t)

	err := client.AddRecord(t.Context(), "lego.ipv64.net", "_acme-challenge", "TXT", "value")
	require.Error(t, err)
}

func TestClient_DeleteRecord(t *testing.T) {
	client := servermock.NewBuilder[*Client](setupClient,
		servermock.CheckHeader().WithContentTypeFromURLEncoded()).
		Route("DELETE /api",
			// the query parameters can be checked because the Go server ignores the body of a DELETE request.
			servermock.ResponseFromFixture("del_record.json").
				WithStatusCode(http.StatusAccepted)).
		Build(t)

	err := client.DeleteRecord(t.Context(), "lego.ipv64.net", "_acme-challenge", "TXT", "value")
	require.NoError(t, err)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client := servermock.NewBuilder[*Client](setupClient).
		Route("DELETE /api",
			servermock.ResponseFromFixture("del_record-error.json").
				WithStatusCode(http.StatusBadRequest)).
		Build(t)

	err := client.DeleteRecord(t.Context(), "lego.ipv64.net", "_acme-challenge", "TXT", "value")
	require.Error(t, err)
}
