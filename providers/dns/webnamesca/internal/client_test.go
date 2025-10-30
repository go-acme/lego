package internal

import (
	"net/http"
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
			client, err := NewClient("user", "secret")
			if err != nil {
				return nil, err
			}

			client.BaseURL, _ = url.Parse(server.URL)
			client.HTTPClient = server.Client()

			return client, nil
		},
		servermock.CheckHeader().
			With("API-User", "user").
			With("API-Key", "secret").
			WithJSONHeaders(),
	)
}

func TestClient_AddTXTRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /domains/example.com/add-txt-record",
			servermock.ResponseFromFixture("add_txt_record.json"),
			servermock.CheckQueryParameter().Strict().
				With("hostName", "foo.example.com").
				With("txt", "value")).
		Build(t)

	result, err := client.AddTXTRecord(t.Context(), "example.com", "foo.example.com", "value")
	require.NoError(t, err)

	expected := []DNSRecordSet{{
		Hostname: "_acme-challenge.example.com",
		Type:     "TXT",
		Records:  []string{"value"},
	}}

	assert.Equal(t, expected, result)
}

func TestClient_AddTXTRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /domains/example.com/add-txt-record",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusBadRequest)).
		Build(t)

	_, err := client.AddTXTRecord(t.Context(), "example.com", "foo.example.com", "value")
	require.EqualError(t, err, "message: User does not exist., details: string, logiD: 35579, result: {}")
}

func TestClient_DeleteTXTRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /domains/example.com/delete-txt-record",
			servermock.ResponseFromFixture("delete_txt_record.json"),
			servermock.CheckQueryParameter().Strict().
				With("hostName", "foo.example.com").
				With("txt", "value")).
		Build(t)

	result, err := client.DeleteTXTRecord(t.Context(), "example.com", "foo.example.com", "value")
	require.NoError(t, err)

	expected := []DNSRecordSet{{
		Hostname: "_acme-challenge.example.com",
		Type:     "TXT",
		Records:  []string{"value"},
	}}

	assert.Equal(t, expected, result)
}

func TestClient_DeleteTXTRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /domains/example.com/delete-txt-record",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusBadRequest)).
		Build(t)

	_, err := client.DeleteTXTRecord(t.Context(), "example.com", "foo.example.com", "value")
	require.EqualError(t, err, "message: User does not exist., details: string, logiD: 35579, result: {}")
}
