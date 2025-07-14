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

func mockBuilder(apiKey string) *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client := NewClient(apiKey)
			client.baseURL, _ = url.Parse(server.URL)
			client.HTTPClient = server.Client()

			return client, nil
		},
		servermock.CheckHeader().WithJSONHeaders().
			WithAuthorization(apiKey))
}

func TestClient_GetTxtRecord(t *testing.T) {
	const apiKey = "myKeyA"

	const domain = "example.com"

	client := mockBuilder(apiKey).
		Route("GET /cdn/4.0/domains/"+domain+"/dns-records",
			servermock.ResponseFromFixture("get_txt_record.json"),
			servermock.CheckQueryParameter().With("search", "acme-challenge")).
		Build(t)

	_, err := client.GetTxtRecord(t.Context(), domain, "_acme-challenge", "txtxtxt")
	require.NoError(t, err)
}

func TestClient_CreateRecord(t *testing.T) {
	const apiKey = "myKeyB"

	const domain = "example.com"

	client := mockBuilder(apiKey).
		Route("POST /cdn/4.0/domains/"+domain+"/dns-records",
			servermock.ResponseFromFixture("create_txt_record.json").
				WithStatusCode(http.StatusCreated),
			servermock.CheckRequestJSONBodyFromFixture("create_record-request.json")).
		Build(t)

	record := DNSRecord{
		Name:  "_acme-challenge",
		Type:  "txt",
		Value: &TXTRecordValue{Text: "txtxtxt"},
		TTL:   600,
	}

	newRecord, err := client.CreateRecord(t.Context(), domain, record)
	require.NoError(t, err)

	expected := &DNSRecord{
		ID:            "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
		Type:          "txt",
		Value:         map[string]any{"text": "txtxtxt"},
		Name:          "_acme-challenge",
		TTL:           120,
		UpstreamHTTPS: "default",
		IPFilterMode: &IPFilterMode{
			Count:     "single",
			Order:     "none",
			GeoFilter: "none",
		},
	}

	assert.Equal(t, expected, newRecord)
}

func TestClient_DeleteRecord(t *testing.T) {
	const apiKey = "myKeyC"

	const domain = "example.com"
	const recordID = "recordId"

	client := mockBuilder(apiKey).
		Route("DELETE /cdn/4.0/domains/"+domain+"/dns-records/"+recordID, nil).
		Build(t)

	err := client.DeleteRecord(t.Context(), domain, recordID)
	require.NoError(t, err)
}
