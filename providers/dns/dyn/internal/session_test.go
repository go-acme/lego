package internal

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockContext() context.Context {
	return context.WithValue(context.Background(), tokenKey, "tok")
}

func TestClient_login(t *testing.T) {
	client := setupTest(t, "/Session", unauthenticatedHandler(http.MethodPost, http.StatusOK, "login.json"))

	sess, err := client.login(context.Background())
	require.NoError(t, err)

	expected := session{Token: "tok", Version: "456"}

	assert.Equal(t, expected, sess)
}

func TestClient_Logout(t *testing.T) {
	client := setupTest(t, "/Session", authenticatedHandler(http.MethodDelete, http.StatusOK, ""))

	err := client.Logout(mockContext())
	require.NoError(t, err)
}

func TestClient_CreateAuthenticatedContext(t *testing.T) {
	client := setupTest(t, "/Session", unauthenticatedHandler(http.MethodPost, http.StatusOK, "login.json"))

	ctx, err := client.CreateAuthenticatedContext(context.Background())
	require.NoError(t, err)

	at := getToken(ctx)
	assert.Equal(t, "tok", at)
}
