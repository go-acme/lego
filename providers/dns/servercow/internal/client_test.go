package internal

import (
	"encoding/json"
	"net/http/httptest"
	"net/url"
	"os"
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
			client.baseURL, _ = url.Parse(server.URL)

			return client, nil
		},
		servermock.CheckHeader().WithJSONHeaders().
			With("X-Auth-Username", "user").
			With("X-Auth-Password", "secret"),
	)
}

func TestClient_GetRecords(t *testing.T) {
	client := mockBuilder().
		Route("GET /lego.wtf", servermock.ResponseFromFixture("records-01.json")).
		Build(t)

	records, err := client.GetRecords(t.Context(), "lego.wtf")
	require.NoError(t, err)

	recordsJSON, err := json.Marshal(records)
	require.NoError(t, err)

	expectedContent, err := os.ReadFile("./fixtures/records-01.json")
	require.NoError(t, err)

	assert.JSONEq(t, string(expectedContent), string(recordsJSON))
}

func TestClient_GetRecords_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /lego.wtf", servermock.JSONEncode(Message{ErrorMsg: "authentication failed"})).
		Build(t)

	records, err := client.GetRecords(t.Context(), "lego.wtf")
	require.Error(t, err)

	assert.Nil(t, records)
}

func TestClient_CreateUpdateRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /lego.wtf",
			servermock.JSONEncode(Message{Message: "ok"}),
			servermock.CheckRequestJSONBody(`{"name":"_acme-challenge.www","type":"TXT","ttl":30,"content":["aaa","bbb"]}`)).
		Build(t)

	record := Record{
		Name:    "_acme-challenge.www",
		Type:    "TXT",
		TTL:     30,
		Content: Value{"aaa", "bbb"},
	}

	msg, err := client.CreateUpdateRecord(t.Context(), "lego.wtf", record)
	require.NoError(t, err)

	expected := &Message{Message: "ok"}
	assert.Equal(t, expected, msg)
}

func TestClient_CreateUpdateRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /lego.wtf",
			servermock.JSONEncode(Message{ErrorMsg: "parameter type must be cname, txt, tlsa, caa, a or aaaa"})).
		Build(t)

	record := Record{
		Name: "_acme-challenge.www",
	}

	msg, err := client.CreateUpdateRecord(t.Context(), "lego.wtf", record)
	require.Error(t, err)

	assert.Nil(t, msg)
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /lego.wtf",
			servermock.JSONEncode(Message{Message: "ok"}),
			servermock.CheckRequestJSONBody(`{"name":"_acme-challenge.www","type":"TXT"}`)).
		Build(t)

	record := Record{
		Name: "_acme-challenge.www",
		Type: "TXT",
	}

	msg, err := client.DeleteRecord(t.Context(), "lego.wtf", record)
	require.NoError(t, err)

	expected := &Message{Message: "ok"}
	assert.Equal(t, expected, msg)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /lego.wtf",
			servermock.JSONEncode(Message{ErrorMsg: "parameter type must be cname, txt, tlsa, caa, a or aaaa"})).
		Build(t)

	record := Record{
		Name: "_acme-challenge.www",
	}

	msg, err := client.DeleteRecord(t.Context(), "lego.wtf", record)
	require.Error(t, err)

	assert.Nil(t, msg)
}
