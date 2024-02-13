package internal

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T, method, pattern string, status int, file string) *Client {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	mux.HandleFunc(pattern, func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != method {
			http.Error(rw, fmt.Sprintf("unsupported method: %s", req.Method), http.StatusBadRequest)
			return
		}

		auth := req.Header.Get(apiKeyHeader)
		if auth != "secret" {
			http.Error(rw, fmt.Sprintf("invalid API key: %s", auth), http.StatusUnauthorized)
			return
		}

		if file == "" {
			rw.WriteHeader(status)
			return
		}

		open, err := os.Open(filepath.Join("fixtures", file))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		defer func() { _ = open.Close() }()

		rw.WriteHeader(status)
		_, err = io.Copy(rw, open)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	client := NewClient("secret")
	client.HTTPClient = server.Client()
	client.baseURL, _ = url.Parse(server.URL)

	return client
}

func TestUpdateZone(t *testing.T) {
	domain := "example.com"

	client := setupTest(t, http.MethodPatch, "/v1/zones/"+domain, http.StatusOK, "update.json")

	patch := ZoneRequest{
		Data: Zone{
			Type: "zone",
			ID:   domain,
			Attributes: Attributes{
				Records: map[string]map[string][]Record{
					"_acme-challenge.test": {
						"TXT": []Record{
							{Data: "test"},
							{Data: "test1"},
							{Data: "test2"},
						},
					},
				},
			},
		},
	}

	zone, err := client.UpdateZone(context.Background(), domain, patch)
	require.NoError(t, err)

	expected := &APIResponse[*Zone]{
		Meta: &Metadata{
			InvocationID: "95cdcc21-b9c3-4b21-8bd1-b05c34c56147",
		},
		Data: &Zone{
			Type: "zone",
			ID:   "dipcon.com",
			Attributes: Attributes{
				OrganisationID:          "10154",
				OrganisationDescription: "My Company AB",
				DNSTypeDescription:      "Anycast",
				Slave:                   false,
				Pending:                 false,
				Deleted:                 false,
				Settings: &Settings{
					MName:   "dns01.dipcon.com.",
					Refresh: 3600,
					Expire:  604800,
					TTL:     600,
				},
				Records: map[string]map[string][]Record{
					"@": {
						"NS": {
							{
								TTL:      3600,
								Data:     "193.14.90.194",
								Comments: "this is a comment",
							},
						},
					},
				},
				Redirects: map[string][]Redirect{
					"<name>": {
						{
							Path:        "/x/y",
							Destination: "https://abion.com/?ref=dipcon",
							Status:      301,
							Slugs:       true,
							Certificate: true,
						},
					},
				},
			},
		},
	}

	assert.Equal(t, expected, zone)
}

func TestGetZones(t *testing.T) {
	client := setupTest(t, http.MethodGet, "/v1/zones/", http.StatusOK, "zones.json")

	zones, err := client.GetZones(context.Background(), nil)
	require.NoError(t, err)

	expected := &APIResponse[[]Zone]{
		Meta: &Metadata{
			InvocationID: "95cdcc21-b9c3-4b21-8bd1-b05c34c56147",
			Pagination: &Pagination{
				Offset: 0,
				Limit:  1,
				Total:  1,
			},
		},
		Data: []Zone{
			{
				Type: "zone",
				ID:   "dipcon.com",
				Attributes: Attributes{
					OrganisationID:          "10154",
					OrganisationDescription: "My Company AB",
					DNSTypeDescription:      "Anycast",
					Slave:                   true,
					Pending:                 true,
					Deleted:                 true,
				},
			},
		},
	}

	assert.Equal(t, expected, zones)
}

func TestGetZone(t *testing.T) {
	domain := "example.com"

	client := setupTest(t, http.MethodGet, "/v1/zones/"+domain, http.StatusOK, "zone.json")

	zones, err := client.GetZone(context.Background(), domain)
	require.NoError(t, err)

	expected := &APIResponse[*Zone]{
		Meta: &Metadata{
			InvocationID: "95cdcc21-b9c3-4b21-8bd1-b05c34c56147",
		},
		Data: &Zone{
			Type: "zone",
			ID:   "dipcon.com",
			Attributes: Attributes{
				OrganisationID:          "10154",
				OrganisationDescription: "My Company AB",
				DNSTypeDescription:      "Anycast",
				Slave:                   false,
				Pending:                 false,
				Deleted:                 false,
				Settings: &Settings{
					MName:   "dns01.dipcon.com.",
					Refresh: 3600,
					Expire:  604800,
					TTL:     600,
				},
				Records: map[string]map[string][]Record{
					"@": {
						"NS": {
							{
								TTL:      3600,
								Data:     "193.14.90.194",
								Comments: "this is a comment",
							},
						},
					},
				},
				Redirects: map[string][]Redirect{
					"<name>": {
						{
							Path:        "/x/y",
							Destination: "https://abion.com/?ref=dipcon",
							Status:      301,
							Slugs:       true,
							Certificate: true,
						},
					},
				},
			},
		},
	}

	assert.Equal(t, expected, zones)
}
