package internal

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const fakeToken = "BAMAuthToken: dQfuRMTUxNjc3MjcyNDg1ODppcGFybXM="

func TestClient_CreateAuthenticatedContext(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := NewClient(server.URL, "user", "secret")
	client.HTTPClient = server.Client()

	mux.HandleFunc("/Services/REST/v1/login", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(rw, fmt.Sprintf("unsupported method %s", req.Method), http.StatusBadRequest)
			return
		}

		query := req.URL.Query()
		if query.Get("username") != "user" {
			http.Error(rw, fmt.Sprintf("invalid username %s", query.Get("username")), http.StatusUnauthorized)
			return
		}

		if query.Get("password") != "secret" {
			http.Error(rw, fmt.Sprintf("invalid password %s", query.Get("password")), http.StatusUnauthorized)
			return
		}

		_, _ = fmt.Fprint(rw, fakeToken)
	})
	mux.HandleFunc("/Services/REST/v1/delete", func(rw http.ResponseWriter, req *http.Request) {
		authorization := req.Header.Get(authorizationHeader)
		if authorization != fakeToken {
			http.Error(rw, fmt.Sprintf("invalid credential: %s", authorization), http.StatusUnauthorized)
			return
		}
	})

	ctx, err := client.CreateAuthenticatedContext(t.Context())
	require.NoError(t, err)

	at := getToken(ctx)
	assert.Equal(t, fakeToken, at)

	err = client.Delete(ctx, 123)
	require.NoError(t, err)
}
