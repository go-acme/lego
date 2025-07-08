package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockContext(t *testing.T) context.Context {
	t.Helper()

	return context.WithValue(t.Context(), tokenKey, &Token{AccessToken: "xxx"})
}

func tokenHandler(rw http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(rw, fmt.Sprintf("invalid method, got %s want %s", req.Method, http.MethodPost), http.StatusMethodNotAllowed)
		return
	}

	err := req.ParseForm()
	if err != nil {
		http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	grantType := req.Form.Get("grant_type")
	clientID := req.Form.Get("client_id")
	clientSecret := req.Form.Get("client_secret")

	if clientID != "user" || clientSecret != "secret" || grantType != "access_key" {
		http.Error(rw, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	_ = json.NewEncoder(rw).Encode(Token{
		AccessToken: "xxx",
		TokenID:     "yyy",
		ExpiresIn:   666,
		TokenType:   "Bearer",
		Scope:       "openid profile email roles",
	})
}

func TestClient_obtainToken(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	mux.HandleFunc("/", tokenHandler)

	client := NewClient("user", "secret")
	client.HTTPClient = server.Client()
	client.AuthEndpoint, _ = url.Parse(server.URL)

	assert.Nil(t, client.token)

	tok, err := client.obtainToken(t.Context())
	require.NoError(t, err)

	assert.NotNil(t, tok)
	assert.NotZero(t, tok.Deadline)
	assert.Equal(t, "xxx", tok.AccessToken)
}

func TestClient_CreateAuthenticatedContext(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	mux.HandleFunc("/", tokenHandler)

	client := NewClient("user", "secret")
	client.HTTPClient = server.Client()
	client.AuthEndpoint, _ = url.Parse(server.URL)

	assert.Nil(t, client.token)

	ctx, err := client.CreateAuthenticatedContext(t.Context())
	require.NoError(t, err)

	tok := getToken(ctx)

	assert.NotNil(t, tok)
	assert.NotZero(t, tok.Deadline)
	assert.Equal(t, "xxx", tok.AccessToken)
}
