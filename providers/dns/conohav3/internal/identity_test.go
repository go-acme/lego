package internal

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	servermock2 "github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupIdentifier(server *httptest.Server) (*Identifier, error) {
	identifier, err := NewIdentifier("c3j1")
	if err != nil {
		return nil, err
	}

	identifier.HTTPClient = server.Client()
	identifier.baseURL, _ = url.Parse(server.URL)

	return identifier, nil
}

func TestGetToken_HeaderToken(t *testing.T) {
	identifier := servermock2.NewBuilder[*Identifier](setupIdentifier,
		servermock2.CheckHeader().WithJSONHeaders(),
	).
		Route("POST /v3/auth/tokens",
			servermock2.ResponseFromFixture("empty.json").
				WithStatusCode(http.StatusCreated).
				WithHeader("x-subject-token", "sample-header-token-123")).
		Build(t)

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

	token, err := identifier.GetToken(t.Context(), auth)
	require.NoError(t, err)

	assert.Equal(t, "sample-header-token-123", token)
}
