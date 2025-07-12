package internal

import (
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const fakeToken = "BAMAuthToken: dQfuRMTUxNjc3MjcyNDg1ODppcGFybXM="

func TestClient_CreateAuthenticatedContext(t *testing.T) {
	client := servermock.NewBuilder[*Client](setupClient).
		Route("GET /Services/REST/v1/login",
			servermock.RawStringResponse(fakeToken),
			servermock.CheckQueryParameter().
				With("username", "user").
				With("password", "secret")).
		Route("DELETE /Services/REST/v1/delete", nil,
			servermock.CheckHeader().
				WithAuthorization(fakeToken)).
		Build(t)

	ctx, err := client.CreateAuthenticatedContext(t.Context())
	require.NoError(t, err)

	at := getToken(ctx)
	assert.Equal(t, fakeToken, at)

	err = client.Delete(ctx, 123)
	require.NoError(t, err)
}
