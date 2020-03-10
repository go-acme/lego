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

	transport, err := NewTokenTransport(apiKey)
	require.NoError(t, err)
	assert.NotNil(t, transport)
}

func TestNewTokenTransport_missing_credentials(t *testing.T) {
	apiKey := ""

	transport, err := NewTokenTransport(apiKey)
	require.Error(t, err)
	assert.Nil(t, transport)
}

func TestTokenTransport_RoundTrip(t *testing.T) {
	apiKey := "api"

	transport, err := NewTokenTransport(apiKey)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)

	resp, err := transport.RoundTrip(req)
	require.NoError(t, err)

	assert.Equal(t, "api", resp.Request.Header.Get(apiKeyHeader))
}
