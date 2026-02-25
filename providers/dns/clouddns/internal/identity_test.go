package internal

import (
	"testing"

	servermock2 "github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_CreateAuthenticatedContext(t *testing.T) {
	client := mockBuilder().
		Route("POST /login",
			servermock2.ResponseFromFixture("login.json"),
			servermock2.CheckRequestJSONBodyFromFixture("login-request.json")).
		Route("DELETE /api/record/xxx", nil).
		Build(t)

	ctx, err := client.CreateAuthenticatedContext(t.Context())
	require.NoError(t, err)

	at := getAccessToken(ctx)
	assert.Equal(t, "at", at)

	err = client.deleteRecord(ctx, Record{ID: "xxx"})
	require.NoError(t, err)
}
