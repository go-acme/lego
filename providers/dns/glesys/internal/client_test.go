package internal

import (
	"net/http/httptest"
	"net/url"
	"testing"

	servermock2 "github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock2.Builder[*Client] {
	return servermock2.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client := NewClient("user", "secret")
			client.HTTPClient = server.Client()
			client.baseURL, _ = url.Parse(server.URL)

			return client, nil
		},
		servermock2.CheckHeader().WithJSONHeaders().
			WithBasicAuth("user", "secret"),
	)
}

func TestClient_AddTXTRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /domain/addrecord",
			servermock2.ResponseFromFixture("add-record.json"),
			servermock2.CheckRequestJSONBody(`{"domainname":"example.com","host":"foo","type":"TXT","data":"txt","ttl":120}`)).
		Build(t)

	recordID, err := client.AddTXTRecord(t.Context(), "example.com", "foo", "txt", 120)
	require.NoError(t, err)

	assert.Equal(t, 123, recordID)
}

func TestClient_DeleteTXTRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /domain/deleterecord",
			servermock2.ResponseFromFixture("delete-record.json"),
			servermock2.CheckRequestJSONBody(`{"recordid":123}`)).
		Build(t)

	err := client.DeleteTXTRecord(t.Context(), 123)
	require.NoError(t, err)
}
