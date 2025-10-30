package sender

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDo_UserAgentOnAllHTTPMethod(t *testing.T) {
	var ua, method string

	server := httptest.NewTLSServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		ua = r.Header.Get("User-Agent")
		method = r.Method
	}))
	t.Cleanup(server.Close)

	doer := NewDoer(server.Client(), "")

	testCases := []struct {
		method string
		call   func(u string) (*http.Response, error)
	}{
		{
			method: http.MethodGet,
			call: func(u string) (*http.Response, error) {
				return doer.Get(u, nil)
			},
		},
		{
			method: http.MethodHead,
			call:   doer.Head,
		},
		{
			method: http.MethodPost,
			call: func(u string) (*http.Response, error) {
				return doer.Post(u, strings.NewReader("falalalala"), "text/plain", nil)
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.method, func(t *testing.T) {
			_, err := test.call(server.URL)
			require.NoError(t, err)

			assert.Equal(t, test.method, method)
			assert.Contains(t, ua, ourUserAgent, "User-Agent")
		})
	}
}

func TestDo_CustomUserAgent(t *testing.T) {
	customUA := "MyApp/1.2.3"
	doer := NewDoer(http.DefaultClient, customUA)

	ua := doer.formatUserAgent()
	assert.Contains(t, ua, ourUserAgent)
	assert.Contains(t, ua, customUA)

	if strings.HasSuffix(ua, " ") {
		t.Errorf("UA should not have trailing spaces; got '%s'", ua)
	}

	assert.Len(t, strings.Split(ua, " "), 5)
}

func TestDo_failWithHTTP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))
	t.Cleanup(server.Close)

	sender := NewDoer(server.Client(), "test")

	_, err := sender.Post(server.URL, strings.NewReader("data"), "text/plain", nil)
	require.ErrorContains(t, err, "HTTPS is required: http://")
}
