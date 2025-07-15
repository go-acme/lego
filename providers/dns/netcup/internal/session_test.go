package internal

import (
	"context"
	"net/http"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockContext(t *testing.T) context.Context {
	t.Helper()

	return context.WithValue(t.Context(), sessionIDKey, "session-id")
}

func TestClient_Login(t *testing.T) {
	client := mockBuilder().
		Route("POST /", servermock.ResponseFromFixture("login.json"),
			servermock.CheckRequestJSONBodyFromFixture("login-request.json")).
		Build(t)

	sessionID, err := client.login(t.Context())
	require.NoError(t, err)

	assert.Equal(t, "api-session-id", sessionID)
}

func TestClient_Login_errors(t *testing.T) {
	testCases := []struct {
		desc     string
		handler  http.Handler
		expected string
	}{
		{
			desc:     "HTTP error",
			handler:  servermock.Noop().WithStatusCode(http.StatusInternalServerError),
			expected: `loging error: unexpected status code: [status code: 500] body: `,
		},
		{
			desc:     "API error",
			handler:  servermock.ResponseFromFixture("login_error.json"),
			expected: `loging error: an error occurred during the action login: [Status=error, StatusCode=4013, ShortMessage=Validation Error., LongMessage=Message is empty.]`,
		},
		{
			desc:     "responsedata marshaling error",
			handler:  servermock.ResponseFromFixture("login_error_unmarshal.json"),
			expected: `loging error: unable to unmarshal response: [status code: 200] body: "" error: json: cannot unmarshal string into Go value of type internal.LoginResponse`,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			client := mockBuilder().
				Route("POST /", test.handler).
				Build(t)

			sessionID, err := client.login(t.Context())
			assert.EqualError(t, err, test.expected)
			assert.Empty(t, sessionID)
		})
	}
}

func TestClient_Logout(t *testing.T) {
	client := mockBuilder().
		Route("POST /", servermock.ResponseFromFixture("logout.json"),
			servermock.CheckRequestJSONBodyFromFixture("logout-request.json")).
		Build(t)

	err := client.Logout(mockContext(t))
	require.NoError(t, err)
}

func TestClient_Logout_errors(t *testing.T) {
	testCases := []struct {
		desc     string
		handler  http.Handler
		expected string
	}{
		{
			desc:    "HTTP error",
			handler: servermock.Noop().WithStatusCode(http.StatusInternalServerError),
		},
		{
			desc:    "API error",
			handler: servermock.ResponseFromFixture("login_error.json"),
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			client := mockBuilder().
				Route("POST /", test.handler).
				Build(t)

			err := client.Logout(t.Context())
			require.Error(t, err)
		})
	}
}
