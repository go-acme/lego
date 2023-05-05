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

func mockContext() context.Context {
	return context.WithValue(context.Background(), tokenKey, &Token{Token: "xxx"})
}

func tokenHandler(rw http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(rw, fmt.Sprintf("invalid method, got %s want %s", req.Method, http.MethodPost), http.StatusMethodNotAllowed)
		return
	}

	username, password, ok := req.BasicAuth()
	if !ok || username != "user" || password != "secret" {
		http.Error(rw, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	_ = json.NewEncoder(rw).Encode(Token{
		Token:     "xxx",
		Lifetime:  666,
		TokenType: "bearer",
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

	tok, err := client.obtainToken(context.Background())
	require.NoError(t, err)

	assert.NotNil(t, tok)
	assert.NotZero(t, tok.Deadline)
	assert.Equal(t, "xxx", tok.Token)
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

	ctx, err := client.CreateAuthenticatedContext(context.Background())
	require.NoError(t, err)

	tok := getToken(ctx)

	assert.NotNil(t, tok)
	assert.NotZero(t, tok.Deadline)
	assert.Equal(t, "xxx", tok.Token)
}
