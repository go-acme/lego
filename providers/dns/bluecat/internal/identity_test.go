package internal

import (
	"testing"

	servermock2 "github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const fakeToken = "BAMAuthToken: dQfuRMTUxNjc3MjcyNDg1ODppcGFybXM="

func TestClient_CreateAuthenticatedContext(t *testing.T) {
	client := servermock2.NewBuilder[*Client](setupClient).
		Route("GET /Services/REST/v1/login",
			servermock2.RawStringResponse(fakeToken),
			servermock2.CheckQueryParameter().
				With("username", "user").
				With("password", "secret")).
		Route("DELETE /Services/REST/v1/delete", nil,
			servermock2.CheckHeader().
				WithAuthorization(fakeToken)).
		Build(t)

	ctx, err := client.CreateAuthenticatedContext(t.Context())
	require.NoError(t, err)

	at := getToken(ctx)
	assert.Equal(t, fakeToken, at)

	err = client.Delete(ctx, 123)
	require.NoError(t, err)
}
