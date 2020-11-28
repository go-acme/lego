package pdns

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDNSProvider_makeRequest(t *testing.T) {
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
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			host, err := url.Parse(test.baseURL)
			require.NoError(t, err)

			config := &Config{Host: host, APIKey: "secret"}

			p := &DNSProvider{
				config:     config,
				apiVersion: test.apiVersion,
			}

			req, err := p.makeRequest(http.MethodGet, test.uri, nil)
			require.NoError(t, err)

			assert.Equal(t, test.expected, req.URL.String())
			assert.Equal(t, "secret", req.Header.Get("X-API-Key"))
		})
	}
}
