//go:build lego.debug

package clientdebug

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWrap_default(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	client := Wrap(server.Client())

	assert.Equal(t, server.Client(), client)

	resp, err := client.Transport.RoundTrip(httptest.NewRequest(http.MethodGet, server.URL, nil))
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestWrap_enable(t *testing.T) {
	t.Setenv("LEGO_DEBUG_DNS_API_HTTP_CLIENT", "true")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	client := Wrap(server.Client())

	resp, err := client.Transport.RoundTrip(httptest.NewRequest(http.MethodGet, server.URL, nil))
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
