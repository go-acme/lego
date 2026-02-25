package internal

import (
	"context"
	"testing"

	servermock2 "github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockContext(t *testing.T) context.Context {
	t.Helper()

	return context.WithValue(t.Context(), tokenKey, "tok")
}

func TestClient_login(t *testing.T) {
	client := mockBuilder().
		Route("POST /Session", servermock2.ResponseFromFixture("login.json"),
			servermock2.CheckRequestJSONBody(`{"customer_name":"bob","user_name":"user","password":"secret"}`)).
		Build(t)

	sess, err := client.login(t.Context())
	require.NoError(t, err)

	expected := session{Token: "tok", Version: "456"}

	assert.Equal(t, expected, sess)
}

func TestClient_Logout(t *testing.T) {
	client := servermock2.NewBuilder[*Client](setupClient,
		servermock2.CheckHeader().WithJSONHeaders().
			With(authTokenHeader, "tok"),
	).
		Route("DELETE /Session", nil).
		Build(t)

	err := client.Logout(mockContext(t))
	require.NoError(t, err)
}

func TestClient_CreateAuthenticatedContext(t *testing.T) {
	client := mockBuilder().
		Route("POST /Session", servermock2.ResponseFromFixture("login.json"),
			servermock2.CheckRequestJSONBody(`{"customer_name":"bob","user_name":"user","password":"secret"}`)).
		Build(t)

	ctx, err := client.CreateAuthenticatedContext(t.Context())
	require.NoError(t, err)

	at := getToken(ctx)
	assert.Equal(t, "tok", at)
}
