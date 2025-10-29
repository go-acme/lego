package clientdebug

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWrap_redact_env_vars(t *testing.T) {
	t.Setenv("LEGO_DEBUG_DNS_API_HTTP_CLIENT", "true")

	t.Setenv("MY_VAR_01", "env-aaaa-aaaa")
	t.Setenv("MY_VAR_02", "query-aaaa-aaaa")
	t.Setenv("MY_VAR_03", "path-aaaa-aaaa")
	t.Setenv("MY_VAR_04", "request-body-aaaa-aaaa")
	t.Setenv("MY_VAR_05", "request-header-aaaa-aaaa")
	t.Setenv("MY_VAR_06", "response-body-aaaa-aaaa")

	server := httptest.NewServer(fakeResponse())

	client := Wrap(server.Client(),
		WithEnvKeys("MY_VAR_01", "MY_VAR_02", "MY_VAR_03", "MY_VAR_04", "MY_VAR_05", "MY_VAR_06"),
	)

	req := fakeRequest(t, server.URL)

	resp, err := client.Transport.RoundTrip(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestWrap_redact_headers(t *testing.T) {
	t.Setenv("LEGO_DEBUG_DNS_API_HTTP_CLIENT", "true")

	server := httptest.NewServer(fakeResponse())

	client := Wrap(server.Client(),
		WithHeaders("Secret-Request-Header", "Super-Secret-Request-Header", "Secret-Response-Header"),
	)

	req := fakeRequest(t, server.URL)

	resp, err := client.Transport.RoundTrip(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestWrap_redact_values(t *testing.T) {
	t.Setenv("LEGO_DEBUG_DNS_API_HTTP_CLIENT", "true")

	server := httptest.NewServer(fakeResponse())

	client := Wrap(server.Client(),
		WithValues("query-aaaa-aaaa", "path-aaaa-aaaa", "request-body-aaaa-aaaa"),
	)

	req := fakeRequest(t, server.URL)

	resp, err := client.Transport.RoundTrip(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func fakeRequest(t *testing.T, baseURL string) *http.Request {
	t.Helper()

	endpoint, err := url.Parse(baseURL)
	require.NoError(t, err)

	query := endpoint.Query()
	query.Set("foo", "query-aaaa-aaaa")
	endpoint.RawQuery = query.Encode()

	endpoint = endpoint.JoinPath("path-aaaa-aaaa")

	body := `{
	"foo": "request-body-aaaa-aaaa"
}
`

	req := httptest.NewRequest(http.MethodGet, endpoint.String(), bytes.NewBufferString(body))

	req.Header.Set("X-Authorization", "not-redacted")

	req.Header.Set("Secret-Request-Header", "request-header-aaaa-aaaa")
	req.Header.Set("Super-Secret-Request-Header", "env-aaaa-aaaa")

	req.Header.Set("Authorization", "header-aaaa-0000")
	req.Header.Set("Token", "header-aaaa-0001")
	req.Header.Set("X-Token", "header-aaaa-0002")
	req.Header.Set("Auth-Token", "header-aaaa-0003")
	req.Header.Set("X-Auth-Token", "header-aaaa-0004")
	req.Header.Set("Api-Key", "header-aaaa-0006")
	req.Header.Set("X-Api-Key", "header-aaaa-0007")
	req.Header.Set("X-Api-Secret", "header-aaaa-0008")

	req.SetBasicAuth("user", "secret")

	return req
}

func fakeResponse() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Secret-Response-Header", "response-header-aaaa-aaaa")
		_, _ = w.Write([]byte(`{
			"foo": "response-body-aaaa-aaaa"
		}`,
		))
	}
}
