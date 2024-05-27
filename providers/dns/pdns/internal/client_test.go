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

		apiKey := req.Header.Get("X-API-Key")
		if apiKey != "secret" {
			http.Error(rw, fmt.Sprintf("invalid credentials: %s", apiKey), http.StatusBadRequest)
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

	serverURL, _ := url.Parse(server.URL)

	client := NewClient(serverURL, "server", 0, "secret")
	client.HTTPClient = server.Client()

	return client
}

func TestClient_joinPath(t *testing.T) {
	testCases := []struct {
		desc       string
		apiVersion int
		baseURL    string
		uri        string
		expected   string
	}{
		{
			desc:       "host with path",
			apiVersion: 1,
			baseURL:    "https://example.com/test",
			uri:        "/foo",
			expected:   "https://example.com/test/api/v1/foo",
		},
		{
			desc:       "host with path + trailing slash",
			apiVersion: 1,
			baseURL:    "https://example.com/test/",
			uri:        "/foo",
			expected:   "https://example.com/test/api/v1/foo",
		},
		{
			desc:       "no URI",
			apiVersion: 1,
			baseURL:    "https://example.com/test",
			uri:        "",
			expected:   "https://example.com/test/api/v1",
		},
		{
			desc:       "host without path",
			apiVersion: 1,
			baseURL:    "https://example.com",
			uri:        "/foo",
			expected:   "https://example.com/api/v1/foo",
		},
		{
			desc:       "api",
			apiVersion: 1,
			baseURL:    "https://example.com",
			uri:        "/api",
			expected:   "https://example.com/api",
		},
		{
			desc:       "API version 0, host with path",
			apiVersion: 0,
			baseURL:    "https://example.com/test",
			uri:        "/foo",
			expected:   "https://example.com/test/foo",
		},
		{
			desc:       "API version 0, host with path + trailing slash",
			apiVersion: 0,
			baseURL:    "https://example.com/test/",
			uri:        "/foo",
			expected:   "https://example.com/test/foo",
		},
		{
			desc:       "API version 0, no URI",
			apiVersion: 0,
			baseURL:    "https://example.com/test",
			uri:        "",
			expected:   "https://example.com/test",
		},
		{
			desc:       "API version 0, host without path",
			apiVersion: 0,
			baseURL:    "https://example.com",
			uri:        "/foo",
			expected:   "https://example.com/foo",
		},
		{
			desc:       "API version 0, api",
			apiVersion: 0,
			baseURL:    "https://example.com",
			uri:        "/api",
			expected:   "https://example.com/api",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			host, err := url.Parse(test.baseURL)
			require.NoError(t, err)

			client := NewClient(host, "test", test.apiVersion, "secret")

			endpoint := client.joinPath(test.uri)

			assert.Equal(t, test.expected, endpoint.String())
		})
	}
}

func TestClient_GetHostedZone(t *testing.T) {
	client := setupTest(t, http.MethodGet, "/api/v1/servers/server/zones/example.org.", http.StatusOK, "zone.json")
	client.apiVersion = 1

	zone, err := client.GetHostedZone(context.Background(), "example.org.")
	require.NoError(t, err)

	expected := &HostedZone{
		ID:   "example.org.",
		Name: "example.org.",
		URL:  "api/v1/servers/localhost/zones/example.org.",
		Kind: "Master",
		RRSets: []RRSet{
			{
				Name:    "example.org.",
				Type:    "NS",
				Records: []Record{{Content: "ns2.example.org."}, {Content: "ns1.example.org."}},
				TTL:     86400,
			},
			{
				Name:    "example.org.",
				Type:    "SOA",
				Records: []Record{{Content: "ns1.example.org. hostmaster.example.org. 2015120401 10800 15 604800 10800"}},
				TTL:     86400,
			},
			{
				Name:    "ns1.example.org.",
				Type:    "A",
				Records: []Record{{Content: "192.168.0.1"}},
				TTL:     86400,
			},
			{
				Name:    "www.example.org.",
				Type:    "A",
				Records: []Record{{Content: "192.168.0.2"}},
				TTL:     86400,
			},
		},
	}

	assert.Equal(t, expected, zone)
}

func TestClient_GetHostedZone_error(t *testing.T) {
	client := setupTest(t, http.MethodGet, "/api/v1/servers/server/zones/example.org.", http.StatusUnprocessableEntity, "error.json")
	client.apiVersion = 1

	_, err := client.GetHostedZone(context.Background(), "example.org.")
	require.ErrorAs(t, err, &apiError{})
}

func TestClient_GetHostedZone_v0(t *testing.T) {
	client := setupTest(t, http.MethodGet, "/servers/server/zones/example.org.", http.StatusOK, "zone.json")
	client.apiVersion = 0

	zone, err := client.GetHostedZone(context.Background(), "example.org.")
	require.NoError(t, err)

	expected := &HostedZone{
		ID:   "example.org.",
		Name: "example.org.",
		URL:  "api/v1/servers/localhost/zones/example.org.",
		Kind: "Master",
		RRSets: []RRSet{
			{
				Name:    "example.org.",
				Type:    "NS",
				Records: []Record{{Content: "ns2.example.org."}, {Content: "ns1.example.org."}},
				TTL:     86400,
			},
			{
				Name:    "example.org.",
				Type:    "SOA",
				Records: []Record{{Content: "ns1.example.org. hostmaster.example.org. 2015120401 10800 15 604800 10800"}},
				TTL:     86400,
			},
			{
				Name:    "ns1.example.org.",
				Type:    "A",
				Records: []Record{{Content: "192.168.0.1"}},
				TTL:     86400,
			},
			{
				Name:    "www.example.org.",
				Type:    "A",
				Records: []Record{{Content: "192.168.0.2"}},
				TTL:     86400,
			},
		},
	}

	assert.Equal(t, expected, zone)
}

func TestClient_UpdateRecords(t *testing.T) {
	client := setupTest(t, http.MethodPatch, "/api/v1/servers/localhost/zones/example.org.", http.StatusOK, "zone.json")
	client.apiVersion = 1
	client.serverName = "localhost"

	zone := &HostedZone{
		ID:   "example.org.",
		Name: "example.org.",
		URL:  "api/v1/servers/localhost/zones/example.org.",
		Kind: "Master",
	}

	rrSets := RRSets{
		RRSets: []RRSet{{
			Name:       "example.org.",
			Type:       "NS",
			ChangeType: "REPLACE",
			Records: []Record{{
				Content: "192.0.2.5",
				Name:    "ns1.example.org.",
				TTL:     86400,
				Type:    "A",
			}},
		}},
	}

	err := client.UpdateRecords(context.Background(), zone, rrSets)
	require.NoError(t, err)
}

func TestClient_UpdateRecords_NonRootApi(t *testing.T) {
	client := setupTest(t, http.MethodPatch, "/some/path/api/v1/servers/localhost/zones/example.org.", http.StatusOK, "zone.json")
	client.Host = client.Host.JoinPath("some", "path")
	client.apiVersion = 1
	client.serverName = "localhost"

	zone := &HostedZone{
		ID:   "example.org.",
		Name: "example.org.",
		URL:  "some/path/api/v1/servers/server/zones/example.org.",
		Kind: "Master",
	}

	rrSets := RRSets{
		RRSets: []RRSet{{
			Name:       "example.org.",
			Type:       "NS",
			ChangeType: "REPLACE",
			Records: []Record{{
				Content: "192.0.2.5",
				Name:    "ns1.example.org.",
				TTL:     86400,
				Type:    "A",
			}},
		}},
	}

	err := client.UpdateRecords(context.Background(), zone, rrSets)
	require.NoError(t, err)
}

func TestClient_UpdateRecords_v0(t *testing.T) {
	client := setupTest(t, http.MethodPatch, "/servers/localhost/zones/example.org.", http.StatusOK, "zone.json")
	client.apiVersion = 0
	client.serverName = "localhost"

	zone := &HostedZone{
		ID:   "example.org.",
		Name: "example.org.",
		URL:  "servers/localhost/zones/example.org.",
		Kind: "Master",
	}

	rrSets := RRSets{
		RRSets: []RRSet{{
			Name:       "example.org.",
			Type:       "NS",
			ChangeType: "REPLACE",
			Records: []Record{{
				Content: "192.0.2.5",
				Name:    "ns1.example.org.",
				TTL:     86400,
				Type:    "A",
			}},
		}},
	}

	err := client.UpdateRecords(context.Background(), zone, rrSets)
	require.NoError(t, err)
}

func TestClient_Notify(t *testing.T) {
	client := setupTest(t, http.MethodPut, "/api/v1/servers/localhost/zones/example.org./notify", http.StatusOK, "")
	client.apiVersion = 1
	client.serverName = "localhost"

	zone := &HostedZone{
		ID:   "example.org.",
		Name: "example.org.",
		URL:  "api/v1/servers/localhost/zones/example.org.",
		Kind: "Master",
	}

	err := client.Notify(context.Background(), zone)
	require.NoError(t, err)
}

func TestClient_Notify_NonRootApi(t *testing.T) {
	client := setupTest(t, http.MethodPut, "/some/path/api/v1/servers/localhost/zones/example.org./notify", http.StatusOK, "")
	client.Host = client.Host.JoinPath("some", "path")
	client.apiVersion = 1
	client.serverName = "localhost"

	zone := &HostedZone{
		ID:   "example.org.",
		Name: "example.org.",
		URL:  "/some/path/api/v1/servers/server/zones/example.org.",
		Kind: "Master",
	}

	err := client.Notify(context.Background(), zone)
	require.NoError(t, err)
}

func TestClient_Notify_v0(t *testing.T) {
	client := setupTest(t, http.MethodPut, "/api/v1/servers/localhost/zones/example.org./notify", http.StatusOK, "")
	client.apiVersion = 0

	zone := &HostedZone{
		ID:   "example.org.",
		Name: "example.org.",
		URL:  "servers/localhost/zones/example.org.",
		Kind: "Master",
	}

	err := client.Notify(context.Background(), zone)
	require.NoError(t, err)
}

func TestClient_getAPIVersion(t *testing.T) {
	client := setupTest(t, http.MethodGet, "/api", http.StatusOK, "versions.json")

	version, err := client.getAPIVersion(context.Background())
	require.NoError(t, err)

	assert.Equal(t, 4, version)
}
