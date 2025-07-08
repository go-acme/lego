package internal

import (
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
			client := NewClient("user", "secret")
			client.HTTPClient = server.Client()
			client.baseURL, _ = url.Parse(server.URL)

			return client, nil
		},
		clientmock.CheckHeader().WithJSONHeaders().
			WithBasicAuth("user", "secret"),
	)
}

func TestClient_AddTXTRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /domain/addrecord",
			clientmock.ResponseFromFixture("add-record.json"),
			clientmock.CheckRequestJSONBody(`{"domainname":"example.com","host":"foo","type":"TXT","data":"txt","ttl":120}`)).
		Build(t)

	recordID, err := client.AddTXTRecord(t.Context(), "example.com", "foo", "txt", 120)
	require.NoError(t, err)

	assert.Equal(t, 123, recordID)
}

func TestClient_DeleteTXTRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /domain/deleterecord",
			clientmock.ResponseFromFixture("delete-record.json"),
			clientmock.CheckRequestJSONBody(`{"recordid":123}`)).
		Build(t)

	err := client.DeleteTXTRecord(t.Context(), 123)
	require.NoError(t, err)
}
