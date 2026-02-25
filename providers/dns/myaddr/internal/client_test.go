package internal

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	servermock2 "github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock2.Builder[*Client] {
	return servermock2.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			credentials := map[string]string{
				"example": "secret",
			}

			client, err := NewClient(credentials)
			if err != nil {
				return nil, err
			}

			client.HTTPClient = server.Client()
			client.baseURL, _ = url.Parse(server.URL)

			return client, nil
		},
		servermock2.CheckHeader().WithJSONHeaders(),
	)
}

func TestClient_AddTXTRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /update", nil,
			servermock2.CheckRequestJSONBody(`{"key":"secret","acme_challenge":"txt"}`)).
		Build(t)

	err := client.AddTXTRecord(t.Context(), "example", "txt")
	require.NoError(t, err)
}

func TestClient_AddTXTRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /update",
			servermock2.ResponseFromFixture("error.txt").
				WithStatusCode(http.StatusBadRequest)).
		Build(t)

	err := client.AddTXTRecord(t.Context(), "example", "txt")
	require.EqualError(t, err, `unexpected status code: [status code: 400] body: invalid value for "key"`)
}

func TestClient_AddTXTRecord_error_credentials(t *testing.T) {
	client := mockBuilder().
		Route("POST /update", nil).
		Build(t)

	err := client.AddTXTRecord(t.Context(), "nx", "txt")
	require.EqualError(t, err, "subdomain nx not found in credentials, check your credentials map")
}
