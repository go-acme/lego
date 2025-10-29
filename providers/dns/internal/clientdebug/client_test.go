package clientdebug

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"text/template"
	"time"

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

	buf := bytes.NewBufferString("")

	server, client, req := setupTest(t, buf,
		WithEnvKeys("MY_VAR_01", "MY_VAR_02", "MY_VAR_03", "MY_VAR_04", "MY_VAR_05", "MY_VAR_06"),
	)

	resp, err := client.Transport.RoundTrip(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	assertDump(t, server, buf, "env_vars.txt")
}

func TestWrap_redact_headers(t *testing.T) {
	t.Setenv("LEGO_DEBUG_DNS_API_HTTP_CLIENT", "true")

	buf := bytes.NewBufferString("")

	server, client, req := setupTest(t, buf,
		WithHeaders("Secret-Request-Header", "Super-Secret-Request-Header", "Secret-Response-Header"),
	)

	resp, err := client.Transport.RoundTrip(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	assertDump(t, server, buf, "headers.txt")
}

func TestWrap_redact_values(t *testing.T) {
	t.Setenv("LEGO_DEBUG_DNS_API_HTTP_CLIENT", "true")

	buf := bytes.NewBufferString("")

	server, client, req := setupTest(t, buf,
		WithValues("query-aaaa-aaaa", "path-aaaa-aaaa", "request-body-aaaa-aaaa"),
	)

	resp, err := client.Transport.RoundTrip(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	assertDump(t, server, buf, "values.txt")
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
	"bar": "response-body-aaaa-aaaa"
}`,
		))
	}
}

func withWriter(w io.Writer) Option {
	return func(d *DumpTransport) {
		if w != nil {
			d.writer = w
		}
	}
}

func setupTest(t *testing.T, buf io.Writer, opts ...Option) (*httptest.Server, *http.Client, *http.Request) {
	t.Helper()

	server := httptest.NewServer(fakeResponse())

	opts = append(opts, withWriter(buf))

	client := Wrap(server.Client(), opts...)

	req := fakeRequest(t, server.URL)

	return server, client, req
}

func assertDump(t *testing.T, server *httptest.Server, actual *bytes.Buffer, filename string) {
	t.Helper()

	tmpl, err := template.New(filename).ParseFiles(filepath.Join("testdata", filename))
	require.NoError(t, err)

	expected := bytes.NewBufferString("")

	location, err := time.LoadLocation("GMT")
	require.NoError(t, err)

	baseURL, err := url.Parse(server.URL)
	require.NoError(t, err)

	err = tmpl.Execute(expected, map[string]string{
		"Host": baseURL.Host,
		"Date": time.Now().In(location).Format(time.RFC1123),
	})
	require.NoError(t, err)

	assert.Equal(t, expected.String(), strings.ReplaceAll(actual.String(), "\r", ""))
}
