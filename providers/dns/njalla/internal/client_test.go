package internal

import (
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupClient(server *httptest.Server) (*Client, error) {
	client := NewClient("secret")
	client.apiEndpoint = server.URL
	client.HTTPClient = server.Client()

	return client, nil
}

func TestClient_AddRecord(t *testing.T) {
	client := servermock.NewBuilder[*Client](setupClient,
		servermock.CheckHeader().WithJSONHeaders().
			WithAuthorization("Njalla secret"),
	).
		Route("POST /",
			servermock.ResponseFromFixture("add_record.json"),
			servermock.CheckRequestJSONBodyFromFixture("add_record-request.json")).
		Build(t)

	record := Record{
		Content: "foobar",
		Domain:  "test",
		Name:    "example.com",
		TTL:     300,
		Type:    "TXT",
	}

	result, err := client.AddRecord(t.Context(), record)
	require.NoError(t, err)

	expected := &Record{
		ID:      "123",
		Content: "foobar",
		Domain:  "test",
		Name:    "example.com",
		TTL:     300,
		Type:    "TXT",
	}
	assert.Equal(t, expected, result)
}

func TestClient_AddRecord_error(t *testing.T) {
	client := servermock.NewBuilder[*Client](setupClient,
		servermock.CheckHeader().WithJSONHeaders().
			WithAuthorization("Njalla invalid"),
	).
		Route("POST /", servermock.ResponseFromFixture("auth_error.json")).
		Build(t)

	client.token = "invalid"

	record := Record{
		Content: "test",
		Domain:  "test01",
		Name:    "example.com",
		TTL:     300,
		Type:    "TXT",
	}

	result, err := client.AddRecord(t.Context(), record)
	require.EqualError(t, err, "code: 403, message: Invalid token.")

	assert.Nil(t, result)
}

func TestClient_ListRecords(t *testing.T) {
	client := servermock.NewBuilder[*Client](setupClient,
		servermock.CheckHeader().WithJSONHeaders().
			WithAuthorization("Njalla secret"),
	).
		Route("POST /",
			servermock.ResponseFromFixture("list_records.json"),
			servermock.CheckRequestJSONBodyFromFixture("list_records-request.json")).
		Build(t)

	records, err := client.ListRecords(t.Context(), "example.com")
	require.NoError(t, err)

	expected := []Record{
		{
			ID:      "1",
			Domain:  "example.com",
			Content: "test",
			Name:    "test01",
			TTL:     300,
			Type:    "TXT",
		},
		{
			ID:      "2",
			Domain:  "example.com",
			Content: "txtTxt",
			Name:    "test02",
			TTL:     120,
			Type:    "TXT",
		},
	}

	assert.Equal(t, expected, records)
}

func TestClient_ListRecords_error(t *testing.T) {
	client := servermock.NewBuilder[*Client](setupClient,
		servermock.CheckHeader().WithJSONHeaders().
			WithAuthorization("Njalla invalid"),
	).
		Route("POST /", servermock.ResponseFromFixture("auth_error.json")).
		Build(t)

	client.token = "invalid"

	records, err := client.ListRecords(t.Context(), "example.com")
	require.EqualError(t, err, "code: 403, message: Invalid token.")

	assert.Empty(t, records)
}

func TestClient_RemoveRecord(t *testing.T) {
	client := servermock.NewBuilder[*Client](setupClient,
		servermock.CheckHeader().WithJSONHeaders().
			WithAuthorization("Njalla secret"),
	).
		Route("POST /",
			servermock.RawStringResponse(`{"jsonrpc":"2.0"}`),
			servermock.CheckRequestJSONBodyFromFixture("remove_record-request.json")).
		Build(t)

	err := client.RemoveRecord(t.Context(), "123", "example.com")
	require.NoError(t, err)
}

func TestClient_RemoveRecord_error(t *testing.T) {
	client := servermock.NewBuilder[*Client](setupClient,
		servermock.CheckHeader().WithJSONHeaders().
			WithAuthorization("Njalla secret"),
	).
		Route("POST /", servermock.ResponseFromFixture("remove_record_error_missing_domain.json")).
		Build(t)

	err := client.RemoveRecord(t.Context(), "123", "example.com")
	require.EqualError(t, err, "code: 400, message: missing domain")
}
