package internal

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_sign(t *testing.T) {
	testCases := []struct {
		desc       string
		apiKey     string
		apiSecret  string
		path       string
		body       string
		xRequestID string
		expected   string
	}{
		{
			desc:       "simple",
			apiKey:     "key",
			apiSecret:  "secret",
			path:       "restful/v2/domains/example.com/records",
			body:       `{"sub_list":[]}`,
			xRequestID: "8e5bc897-75c8-4a9c-9362-5671405c7761",
			expected:   "1v6pCnviEO/dqOIRV2wSc+YRJHt92L+P2xqkyuTVQyk=",
		},
		{
			desc:       "path with slash prefix",
			apiKey:     "key",
			apiSecret:  "secret",
			path:       "/restful/v2/domains/example.com/records",
			body:       `{"sub_list":[]}`,
			xRequestID: "8e5bc897-75c8-4a9c-9362-5671405c7761",
			expected:   "1v6pCnviEO/dqOIRV2wSc+YRJHt92L+P2xqkyuTVQyk=",
		},
		{
			desc:       "another secret",
			apiKey:     "key",
			apiSecret:  "other-secret",
			path:       "restful/v2/domains/example.com/records",
			body:       `{"sub_list":[]}`,
			xRequestID: "8e5bc897-75c8-4a9c-9362-5671405c7761",
			expected:   "ro+dZ59kmfLkWb1CWDsp/rVUeXSScDDr01J7Yg8Bj/E=",
		},
		{
			desc:      "no xRequestID and body",
			apiKey:    "key",
			apiSecret: "secret",
			path:      "restful/v2/domains/example.com/records",
			expected:  "bmpyYo9wpBlZBrcjmZOGrHdp1nQcsesmr8fxibvpefA=",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			endpoint, err := url.Parse("https://example.com")
			require.NoError(t, err)

			endpoint.JoinPath(test.path)

			req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, endpoint.JoinPath(test.path).String(), http.NoBody)
			require.NoError(t, err)

			client := &Client{apiKey: test.apiKey, apiSecret: test.apiSecret}

			actual := client.sign(req, test.xRequestID, test.body)

			assert.Equal(t, test.expected, actual)
		})
	}
}
