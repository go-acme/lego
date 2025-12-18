package sender

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-acme/lego/v4/acme"
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

func Test_checkError(t *testing.T) {
	testCases := []struct {
		desc   string
		resp   *http.Response
		assert func(t *testing.T, err error)
	}{
		{
			desc: "default",
			resp: &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(bytes.NewBufferString(`{"type":"urn:ietf:params:acme:error:example","detail":"message","status":404}`)),
			},
			assert: errorAs[*acme.ProblemDetails],
		},
		{
			desc: "badNonce",
			resp: &http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       io.NopCloser(bytes.NewBufferString(`{"type":"urn:ietf:params:acme:error:badNonce","detail":"message","status":400}`)),
			},
			assert: errorAs[*acme.NonceError],
		},
		{
			desc: "alreadyReplaced",
			resp: &http.Response{
				StatusCode: http.StatusConflict,
				Body:       io.NopCloser(bytes.NewBufferString(`{"type":"urn:ietf:params:acme:error:alreadyReplaced","detail":"message","status":409}`)),
			},
			assert: errorAs[*acme.AlreadyReplacedError],
		},
		{
			desc: "rateLimited",
			resp: &http.Response{
				StatusCode: http.StatusConflict,
				Header: http.Header{
					"Retry-After": []string{"1"},
				},
				Body: io.NopCloser(bytes.NewBufferString(`{"type":"urn:ietf:params:acme:error:rateLimited","detail":"message","status":429}`)),
			},
			assert: errorAs[*acme.RateLimitedError],
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "https://example.com", nil)

			err := checkError(req, test.resp)
			require.Error(t, err)

			pb := &acme.ProblemDetails{}
			assert.ErrorAs(t, err, &pb)

			test.assert(t, err)
		})
	}
}

func errorAs[T error](t *testing.T, err error) {
	t.Helper()

	var zero T
	assert.ErrorAs(t, err, &zero)
}
