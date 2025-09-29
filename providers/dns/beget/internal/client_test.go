package internal

import (
	"context"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client := NewClient("user", "secret")

			client.HTTPClient = server.Client()
			client.BaseURL, _ = url.Parse(server.URL)

			return client, nil
		},
		servermock.CheckQueryParameter().
			With("login", "user").
			With("passwd", "secret").
			With("input_format", "json").
			With("output_format", "json"),
	)
}

func TestClient_GetTXTRecords(t *testing.T) {
	client := mockBuilder().
		Route("GET /dns/getData",
			servermock.ResponseFromFixture("getData-real.json"),
			servermock.CheckQueryParameter().
				With("input_data", `{"fqdn":"example.com"}`),
		).
		Build(t)

	data, err := client.GetTXTRecords(context.Background(), "example.com")
	require.NoError(t, err)

	expected := []Record{{Data: "v=spf1 redirect=beget.com", TTL: 300}}

	assert.Equal(t, expected, data)
}

func TestClient_ChangeTXTRecord(t *testing.T) {
	client := mockBuilder().
		Route("GET /dns/changeRecords",
			servermock.ResponseFromFixture("changeRecords-doc.json"),
			servermock.CheckQueryParameter().
				With("input_data", `{"fqdn":"sub.example.com","records":{"TXT":[{"value":"txtTXTtxt","priority":10,"ttl":300}]}}`),
		).
		Build(t)

	records := []Record{{Value: "txtTXTtxt", TTL: 300, Priority: 10}}

	err := client.ChangeTXTRecord(context.Background(), "sub.example.com", records)
	require.NoError(t, err)
}

func TestClient_ChangeTXTRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /dns/changeRecords",
			servermock.ResponseFromFixture("error.json")).
		Build(t)

	records := []Record{{Data: "txtTXTtxt", TTL: 300}}

	err := client.ChangeTXTRecord(context.Background(), "sub.example.com", records)
	require.Error(t, err)

	require.EqualError(t, err, "API error: NO_SUCH_METHOD: No such method")
}

func TestClient_ChangeTXTRecord_answer_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /dns/changeRecords",
			servermock.ResponseFromFixture("answer_error.json")).
		Build(t)

	records := []Record{{Data: "txtTXTtxt", TTL: 300}}

	err := client.ChangeTXTRecord(context.Background(), "sub.example.com", records)
	require.Error(t, err)

	require.EqualError(t, err, "API answer error: INVALID_DATA: Login length cannot be greater than 12 characters")
}

func TestClient_ChangeTXTRecord_remove(t *testing.T) {
	client := mockBuilder().
		Route("GET /dns/changeRecords",
			servermock.ResponseFromFixture("changeRecords-doc.json"),
			servermock.CheckQueryParameter().
				With("input_data", `{"fqdn":"sub.example.com","records":{}}`),
		).
		Build(t)

	err := client.ChangeTXTRecord(context.Background(), "sub.example.com", nil)
	require.NoError(t, err)
}
