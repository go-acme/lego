package internal

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder(func(server *httptest.Server) (*Client, error) {
		client, err := NewClient("user", "secret")
		if err != nil {
			return nil, err
		}

		client.baseURL, _ = url.Parse(server.URL)
		client.HTTPClient = server.Client()

		return client, nil
	})
}

func TestClient_CreateTXTRecord(t *testing.T) {
	client := mockBuilder().
		Route("GET /txt-create.php",
			servermock.ResponseFromFixture("success.xml")).
		Build(t)

	err := client.CreateTXTRecord("_acme-challenge.example.com", "value")
	require.NoError(t, err)
}

func TestClient_CreateTXTRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /txt-create.php",
			servermock.ResponseFromFixture("error.xml")).
		Build(t)

	err := client.CreateTXTRecord("_acme-challenge.example.com", "value")
	require.EqualError(t, err, "[status code: 200] 708: Failed Login: user (_acme-challenge.example.com)")
}

func TestClient_DeleteTXTRecord(t *testing.T) {
	client := mockBuilder().
		Route("GET /txt-delete.php",
			servermock.ResponseFromFixture("success.xml")).
		Build(t)

	err := client.DeleteTXTRecord("_acme-challenge.example.com", "value")
	require.NoError(t, err)
}

func TestClient_DeleteTXTRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /txt-delete.php",
			servermock.ResponseFromFixture("error.xml")).
		Build(t)

	err := client.DeleteTXTRecord("_acme-challenge.example.com", "value")
	require.EqualError(t, err, "[status code: 200] 708: Failed Login: user (_acme-challenge.example.com)")
}
