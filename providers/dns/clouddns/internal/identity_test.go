package internal

import (
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_CreateAuthenticatedContext(t *testing.T) {
	client := mockBuilder().
		Route("POST /login",
			servermock.ResponseFromFixture("login.json"),
			servermock.CheckRequestJSONBodyFromFile("login-request.json")).
		Route("DELETE /api/record/xxx", nil).
		Build(t)

	ctx, err := client.CreateAuthenticatedContext(t.Context())
	require.NoError(t, err)

	at := getAccessToken(ctx)
	assert.Equal(t, "at", at)

	err = client.deleteRecord(ctx, Record{ID: "xxx"})
	require.NoError(t, err)
}
