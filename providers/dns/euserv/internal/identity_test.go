package internal

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupIdentifierClient(server *httptest.Server) (*Identifier, error) {
	client, err := NewIdentifier("email@example.com", "secret", "orderA")
	if err != nil {
		return nil, err
	}

	client.BaseURL = server.URL
	client.HTTPClient = server.Client()

	return client, nil
}

func mockContext(t *testing.T) context.Context {
	t.Helper()

	return context.WithValue(t.Context(), sessionIDKey, "aae22b4573276b6c14e72da4c66f9546781201291131766515229")
}

func TestIdentifier_getSessionID(t *testing.T) {
	identifier := servermock.NewBuilder[*Identifier](setupIdentifierClient).
		Route("GET /",
			servermock.ResponseFromFixture("session.json"),
			servermock.CheckQueryParameter().Strict().
				With("lang_id", "2").
				With("method", "json")).
		Build(t)

	id, err := identifier.getSessionID(t.Context())
	require.NoError(t, err)

	assert.Equal(t, "cd52d8bc00c7ad0ec14abcd8d6f9fd1e1010052291546606942", id)
}

func TestIdentifier_login(t *testing.T) {
	identifier := servermock.NewBuilder[*Identifier](setupIdentifierClient).
		Route("GET /",
			servermock.ResponseFromFixture("login.json"),
			servermock.CheckQueryParameter().Strict().
				With("subaction", "login").
				With("email", "email@example.com").
				With("password", "secret").
				With("ord_id", "orderA").
				With("api_version", "2.14.2-0").
				With("sess_id", "aae22b4573276b6c14e72da4c66f9546781201291131766515229").
				With("lang_id", "2").
				With("method", "json")).
		Build(t)

	request := LoginRequest{
		Email:      "email@example.com",
		Password:   "secret",
		OrderID:    "orderA",
		APIVersion: defaultAPIVersion,
	}

	id, err := identifier.login(mockContext(t), request)
	require.NoError(t, err)

	assert.Equal(t, "cd52d8bc00c7ad0ec14abcd8d6f9fd1e1010052291546606942", id)
}

func TestIdentifier_login_error(t *testing.T) {
	identifier := servermock.NewBuilder[*Identifier](setupIdentifierClient).
		Route("GET /",
			servermock.ResponseFromFixture("error.json")).
		Build(t)

	request := LoginRequest{
		Email:      "email@example.com",
		Password:   "secret",
		OrderID:    "orderA",
		APIVersion: defaultAPIVersion,
	}

	_, err := identifier.login(mockContext(t), request)
	require.EqualError(t, err, "10006: Login failed.<br>Please check email address/customer ID and password.")
}

func TestIdentifier_Logout(t *testing.T) {
	identifier := servermock.NewBuilder[*Identifier](setupIdentifierClient).
		Route("GET /",
			servermock.ResponseFromFixture("logout.json"),
			servermock.CheckQueryParameter().Strict().
				With("action", "logout").
				With("lang_id", "2").
				With("sess_id", "aae22b4573276b6c14e72da4c66f9546781201291131766515229").
				With("method", "json")).
		Build(t)

	err := identifier.Logout(mockContext(t))
	require.NoError(t, err)
}
