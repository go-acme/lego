package internal

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client, err := NewClient(server.URL)
			if err != nil {
				return nil, err
			}

			client.HTTPClient = server.Client()

			return client, nil
		})
}

func TestClient_Login(t *testing.T) {
	client := mockBuilder().
		Route("POST /",
			servermock.ResponseFromFixture("login.json"),
			servermock.CheckRequestJSONBodyFromFixture("login-request.json"),
			servermock.CheckQueryParameter().Strict().
				With("login", ""),
		).
		Build(t)

	sessionID, err := client.Login(t.Context(), "user", "secret")
	require.NoError(t, err)

	assert.Equal(t, "abc", sessionID)
}

func TestClient_Login_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /",
			servermock.ResponseFromFixture("error.json"),
		).
		Build(t)

	_, err := client.Login(t.Context(), "user", "secret")
	require.EqualError(t, err, `code: remote_fault, message: The login failed. Username or password wrong., response: false`)
}

func TestClient_GetClientID(t *testing.T) {
	client := mockBuilder().
		Route("POST /",
			servermock.ResponseFromFixture("client_get_id.json"),
			servermock.CheckRequestJSONBodyFromFixture("client_get_id-request.json"),
			servermock.CheckQueryParameter().Strict().
				With("client_get_id", ""),
		).
		Build(t)

	id, err := client.GetClientID(t.Context(), "sessionA", "sysA")
	require.NoError(t, err)

	assert.Equal(t, 123, id)
}

func TestClient_GetZoneID(t *testing.T) {
	client := mockBuilder().
		Route("POST /",
			servermock.ResponseFromFixture("dns_zone_get_id.json"),
			servermock.CheckRequestJSONBodyFromFixture("dns_zone_get_id-request.json"),
			servermock.CheckQueryParameter().Strict().
				With("dns_zone_get_id", ""),
		).
		Build(t)

	zoneID, err := client.GetZoneID(t.Context(), "sessionA", "example.com")
	require.NoError(t, err)

	assert.Equal(t, 123, zoneID)
}

func TestClient_GetZone(t *testing.T) {
	client := mockBuilder().
		Route("POST /",
			servermock.ResponseFromFixture("dns_zone_get.json"),
			servermock.CheckRequestJSONBodyFromFixture("dns_zone_get-request.json"),
			servermock.CheckQueryParameter().Strict().
				With("dns_zone_get", ""),
		).
		Build(t)

	zone, err := client.GetZone(t.Context(), "sessionA", "example.com.")
	require.NoError(t, err)

	expected := &Zone{
		ID:         "456",
		ServerID:   "123",
		SysUserID:  "789",
		SysGroupID: "2",
		Origin:     "example.com.",
		Serial:     "2025102902",
		Active:     "Y",
	}

	assert.Equal(t, expected, zone)
}

func TestClient_GetTXT(t *testing.T) {
	client := mockBuilder().
		Route("POST /",
			servermock.ResponseFromFixture("dns_txt_get.json"),
			servermock.CheckRequestJSONBodyFromFixture("dns_txt_get-request.json"),
			servermock.CheckQueryParameter().Strict().
				With("dns_txt_get", ""),
		).
		Build(t)

	record, err := client.GetTXT(t.Context(), "sessionA", "example.com.")
	require.NoError(t, err)

	expected := &Record{ID: 123}

	assert.Equal(t, expected, record)
}

func TestClient_AddTXT(t *testing.T) {
	client := mockBuilder().
		Route("POST /",
			servermock.ResponseFromFixture("dns_txt_add.json"),
			servermock.CheckRequestJSONBodyFromFixture("dns_txt_add-request.json"),
			servermock.CheckQueryParameter().Strict().
				With("dns_txt_add", ""),
		).
		Build(t)

	now := time.Date(2025, 12, 25, 1, 1, 1, 0, time.UTC)

	params := RecordParams{
		ServerID:     "serverA",
		Zone:         "example.com.",
		Name:         "foo.example.com.",
		Type:         "txt",
		Data:         "txtTXTtxt",
		Aux:          "0",
		TTL:          "3600",
		Active:       "y",
		Stamp:        now.Format("2006-01-02 15:04:05"),
		UpdateSerial: true,
	}

	recordID, err := client.AddTXT(t.Context(), "sessionA", "clientA", params)
	require.NoError(t, err)

	assert.Equal(t, "123", recordID)
}

func TestClient_DeleteTXT(t *testing.T) {
	client := mockBuilder().
		Route("POST /",
			servermock.ResponseFromFixture("dns_txt_delete.json"),
			servermock.CheckRequestJSONBodyFromFixture("dns_txt_delete-request.json"),
			servermock.CheckQueryParameter().Strict().
				With("dns_txt_delete", ""),
		).
		Build(t)

	count, err := client.DeleteTXT(t.Context(), "sessionA", "123")
	require.NoError(t, err)

	assert.Equal(t, 1, count)
}
