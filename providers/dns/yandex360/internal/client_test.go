package internal

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/clientmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *clientmock.Builder[*Client] {
	return clientmock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client, err := NewClient("secret", 123456)
			if err != nil {
				return nil, err
			}

			client.HTTPClient = server.Client()
			client.baseURL, _ = url.Parse(server.URL)

			return client, nil
		},
		clientmock.CheckHeader().WithJSONHeaders().
			WithAuthorization("OAuth secret"))
}

func TestClient_AddRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /directory/v1/org/123456/domains/example.com/dns",
			clientmock.ResponseFromFixture("add-record.json"),
			clientmock.CheckRequestJSONBody(`{"name":"_acme-challenge","text":"txtxtxt","ttl":60,"type":"TXT"}`)).
		Build(t)

	record := Record{
		Name: "_acme-challenge",
		Text: "txtxtxt",
		TTL:  60,
		Type: "TXT",
	}

	newRecord, err := client.AddRecord(t.Context(), "example.com", record)
	require.NoError(t, err)

	expected := &Record{
		ID:   789465,
		Name: "foo",
		Text: "_acme-challenge",
		TTL:  60,
		Type: "txtxtxt",
	}

	assert.Equal(t, expected, newRecord)
}

func TestClient_AddRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /directory/v1/org/123456/domains/example.com/dns",
			clientmock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized)).
		Build(t)

	record := Record{
		Name: "_acme-challenge",
		Text: "txtxtxt",
		TTL:  60,
		Type: "TXT",
	}

	newRecord, err := client.AddRecord(t.Context(), "example.com", record)
	require.Error(t, err)

	assert.Nil(t, newRecord)
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /directory/v1/org/123456/domains/example.com/dns/789456",
			clientmock.ResponseFromFixture("delete-record.json")).
		Build(t)

	err := client.DeleteRecord(t.Context(), "example.com", 789456)
	require.NoError(t, err)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /directory/v1/org/123456/domains/example.com/dns/789456",
			clientmock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized)).
		Build(t)

	err := client.DeleteRecord(t.Context(), "example.com", 789456)
	require.Error(t, err)
}
