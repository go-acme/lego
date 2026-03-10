package internal

import (
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_Login(t *testing.T) {
	client := mockBuilder().
		Route("GET /authenticate/login/",
			servermock.ResponseFromFixture("login.json"),
			servermock.CheckHeader().
				WithAuthorization("Bearer secret"),
		).
		Build(t)

	token, err := client.Login(t.Context())
	require.NoError(t, err)

	assert.Equal(t, "session-token", token)
}

func TestClient_Login_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /authenticate/login/",
			servermock.ResponseFromFixture("error.json"),
		).
		Build(t)

	_, err := client.Login(t.Context())
	require.EqualError(t, err, "2003: Required parameter missing")
}
