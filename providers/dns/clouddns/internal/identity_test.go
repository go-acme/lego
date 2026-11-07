package internal

import (
	"testing"

	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_CreateAuthenticatedContext(t *testing.T) {
	client := mockBuilder().
		Route("POST /login",
			servermock.ResponseFromFixture("login.json"),
			servermock.CheckRequestJSONBodyFromFixture("login-request.json"),
		).
		Route("DELETE /api/record/xxx", nil).
		Build(t)

	ctx, err := client.CreateAuthenticatedContext(t.Context())
	require.NoError(t, err)

	at := getAccessToken(ctx)
	assert.Equal(t, "at", at)

	err = client.deleteRecord(ctx, Record{ID: "xxx"})
	require.NoError(t, err)
}

func TestClient_login(t *testing.T) {
	client := mockBuilder().
		Route("POST /login",
			servermock.ResponseFromFixture("login.json"),
			servermock.CheckRequestJSONBodyFromFixture("login-request.json"),
		).
		Build(t)

	result, err := client.login(t.Context())
	require.NoError(t, err)

	expected := &AuthResponse{
		Auth: Auth{
			AccessToken:  "at",
			ExpiresAt:    "string",
			ExpiresIn:    0,
			RefreshToken: "string",
			TokenType:    "Bearer",
		},
	}

	assert.Equal(t, expected, result)
}
