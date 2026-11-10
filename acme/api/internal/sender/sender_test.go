package sender

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-acme/lego/v5/acme"
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
		call   func(ctx context.Context, u string) (*http.Response, error)
	}{
		{
			method: http.MethodGet,
			call: func(ctx context.Context, u string) (*http.Response, error) {
				return doer.Get(ctx, u, nil)
			},
		},
		{
			method: http.MethodHead,
			call:   doer.Head,
		},
		{
			method: http.MethodPost,
			call: func(ctx context.Context, u string) (*http.Response, error) {
				return doer.Post(ctx, u, strings.NewReader("falalalala"), "text/plain", nil)
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.method, func(t *testing.T) {
			_, err := test.call(t.Context(), server.URL)
			require.NoError(t, err)

			assert.Equal(t, test.method, method)
			assert.Contains(t, ua, ourUserAgent, "User-Agent")
		})
	}
}

func Test_formatUserAgent(t *testing.T) {
	customUA := "MyApp/1.2.3"

	ua := formatUserAgent(customUA)
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

	_, err := sender.Post(t.Context(), server.URL, strings.NewReader("data"), "text/plain", nil)
	require.ErrorContains(t, err, "HTTPS is required: http://")
}

func Test_checkError(t *testing.T) {
	testCases := []struct {
		desc        string
		resp        *http.Response
		assertTypes []func(t *testing.T, err error)
		expected    error
		assert      func(t *testing.T, err, expected error)
	}{
		{
			desc: "default",
			resp: &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(bytes.NewBufferString(`{"type":"urn:ietf:params:acme:error:example","detail":"message","status":404}`)),
			},
			assertTypes: []func(t *testing.T, err error){
				errorAs[*acme.ProblemDetails],
			},
			expected: &acme.ProblemDetails{
				Type:       "urn:ietf:params:acme:error:example",
				Detail:     "message",
				HTTPStatus: 404,
				Method:     "POST",
				URL:        "https://example.com",
			},
		},
		{
			desc: "badNonce",
			resp: &http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       io.NopCloser(bytes.NewBufferString(`{"type":"urn:ietf:params:acme:error:badNonce","detail":"message","status":400}`)),
			},
			assertTypes: []func(t *testing.T, err error){
				errorAs[*acme.NonceError],
				errorAs[*acme.ProblemDetails],
			},
			expected: &acme.NonceError{
				ProblemDetails: &acme.ProblemDetails{
					Type:       acme.BadNonceErrorType,
					Detail:     "message",
					HTTPStatus: 400,
					Method:     "POST",
					URL:        "https://example.com",
				},
			},
		},
		{
			desc: "alreadyReplaced",
			resp: &http.Response{
				StatusCode: http.StatusConflict,
				Body:       io.NopCloser(bytes.NewBufferString(`{"type":"urn:ietf:params:acme:error:alreadyReplaced","detail":"message","status":409}`)),
			},
			assertTypes: []func(t *testing.T, err error){
				errorAs[*acme.AlreadyReplacedError],
				errorAs[*acme.ProblemDetails],
			},
			expected: &acme.AlreadyReplacedError{
				ProblemDetails: &acme.ProblemDetails{
					Type:       acme.AlreadyReplacedErrorType,
					Detail:     "message",
					HTTPStatus: 409,
					Method:     "POST",
					URL:        "https://example.com",
				},
			},
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
			assertTypes: []func(t *testing.T, err error){
				errorAs[*acme.RateLimitedError],
				errorAs[*acme.ProblemDetails],
			},
			expected: &acme.RateLimitedError{
				ProblemDetails: &acme.ProblemDetails{
					Type:       acme.RateLimitedErrorType,
					Detail:     "message",
					HTTPStatus: 429,
					Method:     "POST",
					URL:        "https://example.com",
					RetryAfter: 1000000000,
				},
				RetryAfter: 1000000000,
			},
		},
		{
			desc: "service unavailable (unexpected)",
			resp: &http.Response{
				StatusCode: http.StatusServiceUnavailable,
				Header: http.Header{
					"Retry-After": []string{"2"},
				},
				Body: io.NopCloser(bytes.NewBufferString(http.StatusText(http.StatusServiceUnavailable))),
			},
			assertTypes: []func(t *testing.T, err error){
				errorAs[*acme.ProblemDetails],
			},
			expected: &acme.ProblemDetails{
				Type:       "urn:lego:params:acme:error:unmarshal",
				HTTPStatus: 503,
				Method:     "POST",
				URL:        "https://example.com",
				RetryAfter: 2000000000,
				Detail:     "Service Unavailable",
			},
			assert: errorEqual[*acme.ProblemDetails],
		},
		{
			desc: "service unavailable",
			resp: &http.Response{
				StatusCode: http.StatusServiceUnavailable,
				Header: http.Header{
					"Retry-After": []string{"1"},
				},
				Body: io.NopCloser(bytes.NewBufferString(`{"type":"urn:ietf:params:acme:error:incorrectResponse","detail":"message","status":503}`)),
			},
			assertTypes: []func(t *testing.T, err error){
				errorAs[*acme.ProblemDetails],
			},
			expected: &acme.ProblemDetails{
				Type:       acme.IncorrectResponseErrorType,
				Detail:     "message",
				HTTPStatus: 503,
				Method:     "POST",
				URL:        "https://example.com",
				RetryAfter: 1000000000,
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "https://example.com", nil)

			err := checkError(req, test.resp)
			require.Error(t, err)

			for _, fn := range test.assertTypes {
				fn(t, err)
			}

			if test.assert != nil {
				test.assert(t, err, test.expected)
			} else {
				errorEqual[error](t, err, test.expected)
			}
		})
	}
}

func errorAs[T error](t *testing.T, err error) {
	t.Helper()

	var zero T
	assert.ErrorAs(t, err, &zero)
}

func errorEqual[T error](t *testing.T, err, expected error) {
	t.Helper()

	var zero T
	assert.ErrorAs(t, err, &zero)
	assert.Equal(t, expected, zero)
}
