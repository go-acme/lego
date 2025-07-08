package internal

import (
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/clientmock"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *clientmock.Builder[*Client] {
	return clientmock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client, err := NewClient("user", "secret")
			if err != nil {
				return nil, err
			}

			client.HTTPClient = server.Client()
			client.baseURL = server.URL

			return client, nil
		},
		clientmock.CheckHeader().
			WithBasicAuth("user", "secret"))
}

func TestClient_AddRecord(t *testing.T) {
	client := mockBuilder().
		Route("GET /",
			clientmock.ResponseFromFixture("add_success.txt"),
			clientmock.CheckQueryParameter().Strict().
				With("do", "add").
				With("hostname", "_acme-challenge.sub.example.com.").
				With("type", "txt").
				With("value", "test").
				With("ttl", "300"),
		).
		Build(t)

	record := Record{
		Hostname: "_acme-challenge.sub.example.com.",
		Type:     "txt",
		TTL:      300,
		Value:    "test",
	}

	err := client.AddRecord(t.Context(), record)
	require.NoError(t, err)
}

func TestClient_AddRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /",
			clientmock.ResponseFromFixture("error.txt"),
			clientmock.CheckQueryParameter().
				With("do", "add")).
		Build(t)

	record := Record{
		Hostname: "_acme-challenge.sub.example.com.",
		Type:     "txt",
		TTL:      300,
		Value:    "test",
	}

	err := client.AddRecord(t.Context(), record)
	require.Error(t, err)

	require.EqualError(t, err, "unexpected response: notfqdn: Host _acme-challenge.sub.example.com. malformed / vhn")
}

func TestClient_RemoveRecord(t *testing.T) {
	client := mockBuilder().
		Route("GET /",
			clientmock.ResponseFromFixture("remove_success.txt"),
			clientmock.CheckQueryParameter().Strict().
				With("do", "remove").
				With("hostname", "_acme-challenge.sub.example.com.").
				With("type", "txt").
				With("value", "test").
				With("ttl", "300"),
		).
		Build(t)

	record := Record{
		Hostname: "_acme-challenge.sub.example.com.",
		Type:     "txt",
		TTL:      300,
		Value:    "test",
	}

	err := client.RemoveRecord(t.Context(), record)
	require.NoError(t, err)
}

func TestClient_RemoveRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /",
			clientmock.ResponseFromFixture("error.txt"),
			clientmock.CheckQueryParameter().
				With("do", "remove")).
		Build(t)

	record := Record{
		Hostname: "_acme-challenge.sub.example.com.",
		Type:     "txt",
		TTL:      300,
		Value:    "test",
	}

	err := client.RemoveRecord(t.Context(), record)
	require.Error(t, err)

	require.EqualError(t, err, "unexpected response: notfqdn: Host _acme-challenge.sub.example.com. malformed / vhn")
}
