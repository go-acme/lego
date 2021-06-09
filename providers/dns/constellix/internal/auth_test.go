package internal

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTokenTransport_success(t *testing.T) {
	apiKey := "api"
	secretKey := "secret"

	transport, err := NewTokenTransport(apiKey, secretKey)
	require.NoError(t, err)
	assert.NotNil(t, transport)
}

func TestNewTokenTransport_missing_credentials(t *testing.T) {
	apiKey := ""
	secretKey := ""

	transport, err := NewTokenTransport(apiKey, secretKey)
	require.Error(t, err)
	assert.Nil(t, transport)
}

func TestTokenTransport_RoundTrip(t *testing.T) {
	apiKey := "api"
	secretKey := "secret"

	transport, err := NewTokenTransport(apiKey, secretKey)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)

	resp, err := transport.RoundTrip(req)
	require.NoError(t, err)

	assert.Regexp(t, `api:[^:]{28}:\d{13}`, resp.Request.Header.Get(securityTokenHeader))
}
