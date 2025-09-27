package internal

import (
	"context"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client := NewClient("user", "secret")

			client.HTTPClient = server.Client()
			client.baseURL, _ = url.Parse(server.URL)

			return client, nil
		},
		servermock.CheckQueryParameter().
			With("login", "user").
			With("passwd", "secret").
			With("input_format", "json").
			With("output_format", "json"),
	)
}

func TestClient_AddTXTRecord(t *testing.T) {
	client := mockBuilder().
		Route("GET /dns/changeRecords",
			servermock.ResponseFromFixture("changeRecords.json"),
			servermock.CheckQueryParameter().
				With("input_data", `{"fqdn":"sub.example.com","records":{"TXT":[{"priority":10,"value":"txtTXTtxt"}]}}`),
		).
		Build(t)

	err := client.AddTXTRecord(context.Background(), "example.com", "sub", "txtTXTtxt")
	require.NoError(t, err)
}

func TestClient_AddTXTRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /dns/changeRecords",
			servermock.ResponseFromFixture("error.json")).
		Build(t)

	err := client.AddTXTRecord(context.Background(), "example.com", "sub", "txtTXTtxt")
	require.Error(t, err)

	require.EqualError(t, err, "API error: NO_SUCH_METHOD: No such method")
}

func TestClient_AddTXTRecord_answer_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /dns/changeRecords",
			servermock.ResponseFromFixture("answer_error.json")).
		Build(t)

	err := client.AddTXTRecord(context.Background(), "example.com", "sub", "txtTXTtxt")
	require.Error(t, err)

	require.EqualError(t, err, "API answer error: INVALID_DATA: Login length cannot be greater than 12 characters")
}

func TestClient_RemoveTxtRecord(t *testing.T) {
	client := mockBuilder().
		Route("GET /dns/changeRecords",
			servermock.ResponseFromFixture("changeRecords.json"),
			servermock.CheckQueryParameter().
				With("input_data", `{"fqdn":"sub.example.com","records":{}}`),
		).
		Build(t)

	err := client.RemoveTxtRecord(context.Background(), "example.com", "sub")
	require.NoError(t, err)
}

func TestClient_RemoveTxtRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /dns/changeRecords",
			servermock.ResponseFromFixture("error.json")).
		Build(t)

	err := client.RemoveTxtRecord(context.Background(), "example.com", "sub")
	require.Error(t, err)

	require.EqualError(t, err, "API error: NO_SUCH_METHOD: No such method")
}
