package internal

import (
	"context"
	"net/http"
	"testing"

	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_Login(t *testing.T) {
	client := mockBuilder().
		Route("POST /auth/login",
			servermock.ResponseFromFixture("login.json"),
			servermock.CheckRequestJSONBodyFromFixture("login-request.json"),
		).
		Build(t)

	token, err := client.Login(context.Background())
	require.NoError(t, err)

	assert.Equal(t, "20d65561c9636d262e59aec8582c20c7", token)
}

func TestClient_Login_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /auth/login",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusInternalServerError),
		).
		Build(t)

	_, err := client.Login(context.Background())
	require.EqualError(t, err, "[status code 500] Authentication/Authorization Failed (code: 196)")
}

func TestClient_CreateAuthenticatedContext(t *testing.T) {
	client := mockBuilder().
		Route("POST /auth/login",
			servermock.ResponseFromFixture("login.json"),
			servermock.CheckRequestJSONBodyFromFixture("login-request.json"),
		).
		Build(t)

	ctx, err := client.CreateAuthenticatedContext(context.Background())
	require.NoError(t, err)

	token := getToken(ctx)

	assert.Equal(t, "20d65561c9636d262e59aec8582c20c7", token)
}
