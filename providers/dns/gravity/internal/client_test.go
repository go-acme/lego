package internal

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client, err := NewClient(server.URL, "user", "secret")
			if err != nil {
				return nil, err
			}

			client.HTTPClient = server.Client()

			return client, nil
		},
		servermock.CheckHeader().
			WithJSONHeaders(),
	)
}

func TestClient_Login(t *testing.T) {
	client := mockBuilder().
		Route("POST /api/v1/auth/login",
			http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				http.SetCookie(rw, &http.Cookie{
					Name:  "gravity_session",
					Value: "session_id",
					Path:  "/",
				})

				servermock.ResponseFromFixture("login.json").ServeHTTP(rw, req)
			}),
			servermock.CheckRequestJSONBodyFromFixture("login-request.json")).
		Build(t)

	auth, err := client.Login(t.Context())
	require.NoError(t, err)

	cookies := client.HTTPClient.Jar.Cookies(client.baseURL)

	require.Len(t, cookies, 1)

	assert.Equal(t, "gravity_session", cookies[0].Name)
	assert.Equal(t, "session_id", cookies[0].Value)

	expected := &Auth{Successful: true}

	assert.Equal(t, expected, auth)
}

func TestClient_Login_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /api/v1/auth/login",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized)).
		Build(t)

	_, err := client.Login(t.Context())
	require.EqualError(t, err, "status: UNAUTHENTICATED, error: unauthenticated, additionalProp1: string")
}

func TestClient_Me(t *testing.T) {
	client := mockBuilder().
		Route("GET /api/v1/auth/me",
			servermock.ResponseFromFixture("me.json")).
		Build(t)

	info, err := client.Me(t.Context())
	require.NoError(t, err)

	expected := &UserInfo{
		Username:      "admin",
		Authenticated: true,
		Permissions: []Permission{{
			Methods: []string{"GET", "POST", "PUT", "HEAD", "DELETE"},
			Path:    "/*",
		}},
	}

	assert.Equal(t, expected, info)
}

func TestClient_GetDNSZones(t *testing.T) {
	client := mockBuilder().
		Route("GET /api/v1/dns/",
			servermock.ResponseFromFixture("zones.json"),
			servermock.CheckQueryParameter().Strict().
				With("name", "example.com.")).
		Build(t)

	zones, err := client.GetDNSZones(t.Context(), "example.com.")
	require.NoError(t, err)

	expected := []Zone{{
		Name: "example.com.",
		HandlerConfigs: []HandlerConfig{
			{Type: "memory"},
			{Type: "etcd"},
		},
		DefaultTTL:  86400,
		RecordCount: 1,
	}}

	assert.Equal(t, expected, zones)
}

func TestClient_CreateDNSRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /api/v1/dns/zones/records",
			servermock.Noop().
				WithStatusCode(http.StatusNoContent),
			servermock.CheckRequestJSONBodyFromFixture("create_record-request.json"),
			servermock.CheckQueryParameter().Strict().
				With("zone", "example.com.").
				With("uid", "123").
				With("hostname", "_acme-challenge")).
		Build(t)

	record := Record{
		Data:     "txtTXTtxt",
		Hostname: "_acme-challenge",
		Type:     "TXT",
		UID:      "123",
	}

	err := client.CreateDNSRecord(t.Context(), "example.com.", record)
	require.NoError(t, err)
}

func TestClient_DeleteDNSRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /api/v1/dns/zones/records",
			servermock.Noop().
				WithStatusCode(http.StatusNoContent),
			servermock.CheckQueryParameter().Strict().
				With("zone", "example.com.").
				With("uid", "123").
				With("type", "TXT").
				With("hostname", "_acme-challenge")).
		Build(t)

	record := Record{
		Data:     "txtTXTtxt",
		Hostname: "_acme-challenge",
		Type:     "TXT",
		UID:      "123",
	}

	err := client.DeleteDNSRecord(t.Context(), "example.com.", record)
	require.NoError(t, err)
}
