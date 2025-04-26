package internal

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetToken_HeaderToken(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	identifier, err := NewIdentifier("c3j1")
	require.NoError(t, err)

	identifier.HTTPClient = server.Client()
	identifier.baseURL, _ = url.Parse(server.URL)

	mux.HandleFunc("/v3/auth/tokens", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("x-subject-token", "sample-header-token-123")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{}`))
	})

	auth := Auth{
		Identity: Identity{
			Methods: []string{"password"},
			Password: Password{
				User: User{
					ID:       "dummy-id",
					Password: "dummy-password",
				},
			},
		},
		Scope: Scope{
			Project: Project{
				ID: "dummy-project-id",
			},
		},
	}

	token, err := identifier.GetToken(context.Background(), auth)
	require.NoError(t, err)

	assert.Equal(t, "sample-header-token-123", token)
}
