package internal

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockContext() context.Context {
	return context.WithValue(context.Background(), tokenKey, "593959ca04f0de9689b586c6a647d15d")
}

func TestIdentifier_Authentication(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	mux.HandleFunc("/", testHandler("auth.xml"))

	client := NewIdentifier("user", "secret")
	client.authEndpoint = server.URL

	credentialToken, err := client.Authentication(context.Background(), 60, false)
	require.NoError(t, err)

	assert.Equal(t, "593959ca04f0de9689b586c6a647d15d", credentialToken)
}

func TestIdentifier_Authentication_error(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	mux.HandleFunc("/", testHandler("auth_fault.xml"))

	client := NewIdentifier("user", "secret")
	client.authEndpoint = server.URL

	_, err := client.Authentication(context.Background(), 60, false)
	require.Error(t, err)
}
