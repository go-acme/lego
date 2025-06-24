package internal

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockContext(t *testing.T) context.Context {
	t.Helper()

	return context.WithValue(t.Context(), tokenKey, "tok")
}

func TestClient_login(t *testing.T) {
	client := setupTest(t, "/Session", unauthenticatedHandler(http.MethodPost, http.StatusOK, "login.json"))

	sess, err := client.login(t.Context())
	require.NoError(t, err)

	expected := session{Token: "tok", Version: "456"}

	assert.Equal(t, expected, sess)
}

func TestClient_Logout(t *testing.T) {
	client := setupTest(t, "/Session", authenticatedHandler(http.MethodDelete, http.StatusOK, ""))

	err := client.Logout(mockContext(t))
	require.NoError(t, err)
}

func TestClient_CreateAuthenticatedContext(t *testing.T) {
	client := setupTest(t, "/Session", unauthenticatedHandler(http.MethodPost, http.StatusOK, "login.json"))

	ctx, err := client.CreateAuthenticatedContext(t.Context())
	require.NoError(t, err)

	at := getToken(ctx)
	assert.Equal(t, "tok", at)
}
