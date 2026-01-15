package internal

import (
	"context"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client, err := NewClient(server.URL, "userA", "secret")
			if err != nil {
				return nil, err
			}

			client.baseURL, _ = url.Parse(server.URL)
			client.HTTPClient = server.Client()

			return client, nil
		},
		servermock.CheckHeader().
			WithJSONHeaders(),
	)
}

func mockToken(ctx context.Context) context.Context {
	return context.WithValue(ctx, tokenKey, "secretToken")
}

func TestClient_CreateSession(t *testing.T) {
	client := mockBuilder().
		Route("POST /api/v2/sessions",
			servermock.ResponseFromFixture("postSession.json"),
			servermock.CheckRequestJSONBodyFromFixture("postSession-request.json"),
		).
		Build(t)

	info := LoginInfo{
		Username: "userA",
		Password: "secret",
	}

	result, err := client.CreateSession(mockToken(t.Context()), info)
	require.NoError(t, err)

	expected := &Session{
		ID:                             12345,
		Type:                           "UserSession",
		APIToken:                       "VZoO2Z0BjBaJyvuhE4vNJRWqI9upwDHk70UNi0Ez",
		APITokenExpirationDateTime:     time.Date(2022, time.September, 15, 17, 52, 7, 0, time.UTC),
		BasicAuthenticationCredentials: "YXBpOlQ0OExOT3VIRGhDcnVBNEo1bGFES3JuS3hTZC9QK3pjczZXTzBJMDY=",
		RemoteAddress:                  "192.168.1.1",
		ReadOnly:                       true,
		LoginDateTime:                  time.Date(2022, time.September, 14, 17, 45, 3, 0, time.UTC),
		LogoutDateTime:                 time.Date(2022, time.September, 14, 19, 45, 3, 0, time.UTC),
		State:                          "LOGGED_IN",
		Response:                       "Authentication Error: Ensure that your username and password are correct.",
	}

	assert.Equal(t, expected, result)
}

func TestClient_CreateAuthenticatedContext(t *testing.T) {
	client := mockBuilder().
		Route("POST /api/v2/sessions",
			servermock.ResponseFromFixture("postSession.json"),
			servermock.CheckRequestJSONBodyFromFixture("postSession-request.json"),
		).
		Build(t)

	ctx, err := client.CreateAuthenticatedContext(t.Context())
	require.NoError(t, err)

	assert.Equal(t, "YXBpOlQ0OExOT3VIRGhDcnVBNEo1bGFES3JuS3hTZC9QK3pjczZXTzBJMDY=", getToken(ctx))
}
