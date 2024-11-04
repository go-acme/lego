package internal

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClient_Login(t *testing.T) {
	client := setupTest(t, "POST /login", http.StatusOK, "")

	err := client.Login(context.Background())
	require.NoError(t, err)
}

func TestClient_Login_error(t *testing.T) {
	client := setupTest(t, "POST /login", http.StatusUnauthorized, "error.json")

	err := client.Login(context.Background())
	require.Error(t, err)

	require.EqualError(t, err, "401: bad_login: Unknown username or password")
}
