package internal

import (
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
			client := NewClient("user", "secret", 123)
			client.HTTPClient = server.Client()
			client.BaseURL, _ = url.Parse(server.URL)

			return client, nil
		},
		servermock.CheckHeader().
			WithBasicAuth("user", "secret").
			WithJSONHeaders())
}

func TestClient_AddTxtRecords(t *testing.T) {
	client := mockBuilder().
		Route("POST /zone/example.com/_stream",
			servermock.ResponseFromFixture("add_record.json"),
			servermock.CheckRequestJSONBodyFromFile("add_record-request.json"),
			servermock.CheckHeader().
				With("X-Domainrobot-Context", "123")).
		Build(t)

	records := []*ResourceRecord{{}}

	zone, err := client.AddTxtRecords(t.Context(), "example.com", records)
	require.NoError(t, err)

	expected := &Zone{
		Name: "example.com",
		ResourceRecords: []*ResourceRecord{{
			Name:  "example.com",
			TTL:   120,
			Type:  "TXT",
			Value: "txt",
			Pref:  1,
		}},
		Action:            "xxx",
		VirtualNameServer: "yyy",
	}

	assert.Equal(t, expected, zone)
}

func TestClient_RemoveTXTRecords(t *testing.T) {
	client := mockBuilder().
		Route("POST /zone/example.com/_stream",
			servermock.ResponseFromFixture("remove_record.json"),
			servermock.CheckRequestJSONBodyFromFile("remove_record-request.json"),
			servermock.CheckHeader().
				With("X-Domainrobot-Context", "123")).
		Build(t)

	records := []*ResourceRecord{{}}

	err := client.RemoveTXTRecords(t.Context(), "example.com", records)
	require.NoError(t, err)
}
