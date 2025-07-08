package internal

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_CreateAuthenticatedContext(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/login", func(rw http.ResponseWriter, req *http.Request) {
		response := AuthResponse{
			Auth: Auth{
				AccessToken:  "at",
				RefreshToken: "",
			},
		}

		err := json.NewEncoder(rw).Encode(response)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})
	mux.HandleFunc("/api/record/xxx", func(rw http.ResponseWriter, req *http.Request) {
		authorization := req.Header.Get(authorizationHeader)
		if authorization != "Bearer at" {
			http.Error(rw, "invalid credential: "+authorization, http.StatusUnauthorized)
			return
		}
	})

	ctx, err := client.CreateAuthenticatedContext(t.Context())
	require.NoError(t, err)

	at := getAccessToken(ctx)
	assert.Equal(t, "at", at)

	err = client.deleteRecord(ctx, Record{ID: "xxx"})
	require.NoError(t, err)
}
