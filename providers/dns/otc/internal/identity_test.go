package internal

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_Login(t *testing.T) {
	mock := NewDNSServerMock(t)
	mock.HandleAuthSuccessfully()

	client := NewClient("user", "secret", "example.com", "test")
	client.IdentityEndpoint, _ = url.JoinPath(mock.GetServerURL(), "/v3/auth/token")

	err := client.Login(t.Context())
	require.NoError(t, err)

	serverURL, _ := url.Parse(mock.GetServerURL())
	assert.Equal(t, serverURL.JoinPath("v2").String(), client.baseURL.String())
	assert.Equal(t, fakeOTCToken, client.token)
}
