package internal

import (
	"net/http/httptest"
	"net/url"
	"testing"

	servermock2 "github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock2.Builder[*Client] {
	return servermock2.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			serverURL, _ := url.Parse(server.URL)

			client := NewClient(serverURL, "user", "secret")
			client.HTTPClient = server.Client()

			return client, nil
		},
		servermock2.CheckHeader().WithContentType("text/xml").
			With("Http_auth_login", "user").
			With("Http_auth_passwd", "secret"),
	)
}

func TestClient_GetSite(t *testing.T) {
	client := mockBuilder().
		Route("POST /enterprise/control/agent.php", servermock2.ResponseFromFixture("get-site.xml")).
		Build(t)

	siteID, err := client.GetSite(t.Context(), "example.com")
	require.NoError(t, err)

	assert.Equal(t, 82, siteID)
}

func TestClient_GetSite_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /enterprise/control/agent.php", servermock2.ResponseFromFixture("get-site-error.xml")).
		Build(t)

	siteID, err := client.GetSite(t.Context(), "example.com")
	require.Error(t, err)

	assert.Equal(t, 0, siteID)
}

func TestClient_GetSite_system_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /enterprise/control/agent.php", servermock2.ResponseFromFixture("global-error.xml")).
		Build(t)

	siteID, err := client.GetSite(t.Context(), "example.com")
	require.Error(t, err)

	assert.Equal(t, 0, siteID)
}

func TestClient_AddRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /enterprise/control/agent.php", servermock2.ResponseFromFixture("add-record.xml")).
		Build(t)

	recordID, err := client.AddRecord(t.Context(), 123, "_acme-challenge.example.com", "txtTXTtxt")
	require.NoError(t, err)

	assert.Equal(t, 4537, recordID)
}

func TestClient_AddRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /enterprise/control/agent.php", servermock2.ResponseFromFixture("add-record-error.xml")).
		Build(t)

	recordID, err := client.AddRecord(t.Context(), 123, "_acme-challenge.example.com", "txtTXTtxt")
	require.ErrorAs(t, err, new(RecResult))

	assert.Equal(t, 0, recordID)
}

func TestClient_AddRecord_system_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /enterprise/control/agent.php", servermock2.ResponseFromFixture("global-error.xml")).
		Build(t)

	recordID, err := client.AddRecord(t.Context(), 123, "_acme-challenge.example.com", "txtTXTtxt")
	require.ErrorAs(t, err, new(*System))

	assert.Equal(t, 0, recordID)
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /enterprise/control/agent.php", servermock2.ResponseFromFixture("delete-record.xml")).
		Build(t)

	recordID, err := client.DeleteRecord(t.Context(), 4537)
	require.NoError(t, err)

	assert.Equal(t, 4537, recordID)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /enterprise/control/agent.php", servermock2.ResponseFromFixture("delete-record-error.xml")).
		Build(t)

	recordID, err := client.DeleteRecord(t.Context(), 4537)
	require.ErrorAs(t, err, new(RecResult))

	assert.Equal(t, 0, recordID)
}

func TestClient_DeleteRecord_system_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /enterprise/control/agent.php", servermock2.ResponseFromFixture("global-error.xml")).
		Build(t)

	recordID, err := client.DeleteRecord(t.Context(), 4537)
	require.ErrorAs(t, err, new(*System))

	assert.Equal(t, 0, recordID)
}
