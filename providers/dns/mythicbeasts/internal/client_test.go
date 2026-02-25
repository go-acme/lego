package internal

import (
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	servermock2 "github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock2.Builder[*Client] {
	return servermock2.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client := NewClient("user", "secret")
			client.HTTPClient = server.Client()
			client.APIEndpoint, _ = url.Parse(server.URL)
			client.token = &Token{
				Token:     "secret",
				Lifetime:  60,
				TokenType: "bearer",
				Deadline:  time.Now().Add(1 * time.Minute),
			}

			return client, nil
		},
		servermock2.CheckHeader().WithJSONHeaders().
			WithAuthorization("Bearer "+fakeToken),
	)
}

func TestClient_CreateTXTRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /zones/example.com/records/foo/TXT",
			servermock2.ResponseFromFixture("post-zoneszonerecords.json"),
			servermock2.CheckRequestJSONBody(`{"records":[{"host":"foo","ttl":120,"type":"TXT","data":"txt"}]}`)).
		Build(t)

	err := client.CreateTXTRecord(mockContext(t), "example.com", "foo", "txt", 120)
	require.NoError(t, err)
}

func TestClient_RemoveTXTRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /zones/example.com/records/foo/TXT",
			servermock2.ResponseFromFixture("delete-zoneszonerecords.json"),
			servermock2.CheckQueryParameter().Strict().
				With("data", "txt")).
		Build(t)

	err := client.RemoveTXTRecord(mockContext(t), "example.com", "foo", "txt")
	require.NoError(t, err)
}
