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
			serverURL, _ := url.Parse(server.URL)

			client := NewClient(serverURL, "server", 0, "secret")
			client.HTTPClient = server.Client()

			return client, nil
		},
		servermock.CheckHeader().WithJSONHeaders().With(APIKeyHeader, "secret"))
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
	client := mockBuilder().
		Route("GET /api/v1/servers/server/zones/example.org.",
			servermock.ResponseFromFixture("zone.json")).
		Build(t)

	client.apiVersion = 1

	zone, err := client.GetHostedZone(t.Context(), "example.org.")
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
	client := mockBuilder().
		Route("GET /api/v1/servers/server/zones/example.org.",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnprocessableEntity)).
		Build(t)

	client.apiVersion = 1

	_, err := client.GetHostedZone(t.Context(), "example.org.")
	require.ErrorAs(t, err, &apiError{})
}

func TestClient_GetHostedZone_v0(t *testing.T) {
	client := mockBuilder().
		Route("GET /servers/server/zones/example.org.",
			servermock.ResponseFromFixture("zone.json")).
		Build(t)

	client.apiVersion = 0

	zone, err := client.GetHostedZone(t.Context(), "example.org.")
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
	client := mockBuilder().
		Route("PATCH /api/v1/servers/localhost/zones/example.org.",
			servermock.ResponseFromFixture("zone.json"),
			servermock.CheckRequestJSONBodyFromFixture("zone-request.json")).
		Build(t)

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

	err := client.UpdateRecords(t.Context(), zone, rrSets)
	require.NoError(t, err)
}

func TestClient_UpdateRecords_NonRootApi(t *testing.T) {
	client := mockBuilder().
		Route("PATCH /some/path/api/v1/servers/localhost/zones/example.org.",
			servermock.ResponseFromFixture("zone.json"),
			servermock.CheckRequestJSONBodyFromFixture("zone-request.json")).
		Build(t)

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

	err := client.UpdateRecords(t.Context(), zone, rrSets)
	require.NoError(t, err)
}

func TestClient_UpdateRecords_v0(t *testing.T) {
	client := mockBuilder().
		Route("PATCH /servers/localhost/zones/example.org.",
			servermock.ResponseFromFixture("zone.json"),
			servermock.CheckRequestJSONBodyFromFixture("zone-request.json")).
		Build(t)

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

	err := client.UpdateRecords(t.Context(), zone, rrSets)
	require.NoError(t, err)
}

func TestClient_Notify(t *testing.T) {
	client := mockBuilder().
		Route("PUT /api/v1/servers/localhost/zones/example.org./notify", nil).
		Build(t)

	client.apiVersion = 1
	client.serverName = "localhost"

	zone := &HostedZone{
		ID:   "example.org.",
		Name: "example.org.",
		URL:  "api/v1/servers/localhost/zones/example.org.",
		Kind: "Master",
	}

	err := client.Notify(t.Context(), zone)
	require.NoError(t, err)
}

func TestClient_Notify_NonRootApi(t *testing.T) {
	client := mockBuilder().
		Route("PUT /some/path/api/v1/servers/localhost/zones/example.org./notify", nil).
		Build(t)

	client.Host = client.Host.JoinPath("some", "path")
	client.apiVersion = 1
	client.serverName = "localhost"

	zone := &HostedZone{
		ID:   "example.org.",
		Name: "example.org.",
		URL:  "/some/path/api/v1/servers/server/zones/example.org.",
		Kind: "Master",
	}

	err := client.Notify(t.Context(), zone)
	require.NoError(t, err)
}

func TestClient_Notify_v0(t *testing.T) {
	client := mockBuilder().
		Route("PUT /some/path/api/v1/servers/localhost/zones/example.org./notify", nil).
		Build(t)

	client.apiVersion = 0

	zone := &HostedZone{
		ID:   "example.org.",
		Name: "example.org.",
		URL:  "servers/localhost/zones/example.org.",
		Kind: "Master",
	}

	err := client.Notify(t.Context(), zone)
	require.NoError(t, err)
}

func TestClient_getAPIVersion(t *testing.T) {
	client := mockBuilder().
		Route("GET /api",
			servermock.ResponseFromFixture("versions.json")).
		Build(t)

	version, err := client.getAPIVersion(t.Context())
	require.NoError(t, err)

	assert.Equal(t, 4, version)
}
