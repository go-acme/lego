package internal

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T) (*Client, *http.ServeMux) {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := NewClient(context.Background(), "STACK_ID", "CLIENT_ID", "CLIENT_SECRET")
	client.httpClient = server.Client()
	client.baseURL, _ = url.Parse(server.URL + "/")

	return client, mux
}

func TestClient_GetZoneRecords(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/STACK_ID/zones/A/records", func(w http.ResponseWriter, _ *http.Request) {
		content := `
			{
				"records": [
					{"id":"1","name":"foo1","type":"TXT","ttl":120,"data":"txtTXTtxt"},
					{"id":"2","name":"foo2","type":"TXT","ttl":121,"data":"TXTtxtTXT"}
				]
			}`

		_, err := w.Write([]byte(content))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	records, err := client.GetZoneRecords(context.Background(), "foo1", &Zone{ID: "A", Domain: "test"})
	require.NoError(t, err)

	expected := []Record{
		{ID: "1", Name: "foo1", Type: "TXT", TTL: 120, Data: "txtTXTtxt"},
		{ID: "2", Name: "foo2", Type: "TXT", TTL: 121, Data: "TXTtxtTXT"},
	}

	assert.Equal(t, expected, records)
}

func TestClient_GetZoneRecords_apiError(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/STACK_ID/zones/A/records", func(w http.ResponseWriter, _ *http.Request) {
		content := `
{
	"code": 401,
	"error": "an unauthorized request is attempted."
}`

		w.WriteHeader(http.StatusUnauthorized)
		_, err := w.Write([]byte(content))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	_, err := client.GetZoneRecords(context.Background(), "foo1", &Zone{ID: "A", Domain: "test"})

	expected := &ErrorResponse{Code: 401, Message: "an unauthorized request is attempted."}
	assert.Equal(t, expected, err)
}

func TestClient_GetZones(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/STACK_ID/zones", func(w http.ResponseWriter, _ *http.Request) {
		content := `
{
  "pageInfo": {
    "totalCount": "5",
    "hasPreviousPage": false,
    "hasNextPage": false,
    "startCursor": "1",
    "endCursor": "1"
  },
  "zones": [
    {
      "stackId": "my_stack",
      "accountId": "my_account",
      "id": "A",
      "domain": "foo.com",
      "version": "1",
      "labels": {
        "property1": "val1",
        "property2": "val2"
      },
      "created": "2018-10-07T02:31:49Z",
      "updated": "2018-10-07T02:31:49Z",
      "nameservers": [
        "1.1.1.1"
      ],
      "verified": "2018-10-07T02:31:49Z",
      "status": "ACTIVE",
      "disabled": false
    }
  ]
}`

		_, err := w.Write([]byte(content))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	zone, err := client.GetZones(context.Background(), "sub.foo.com")
	require.NoError(t, err)

	expected := &Zone{ID: "A", Domain: "foo.com"}

	assert.Equal(t, expected, zone)
}
