package internal

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupNew(t *testing.T, expectedForm, filename string) *Client {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	mux.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		err := req.ParseForm()
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		exp := regexp.MustCompile(`"auth":"\w+",`)

		form := req.PostForm.Get("request")
		form = exp.ReplaceAllString(form, `"auth":"xxx",`)

		if form != expectedForm {
			t.Logf("invalid form data: %s", req.PostForm.Get("request"))
			http.Error(rw, fmt.Sprintf("invalid form data: %s", req.PostForm.Get("request")), http.StatusBadRequest)
			return
		}

		data, err := os.ReadFile(fmt.Sprintf("./fixtures/%s.json", filename))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		rw.Header().Set("Content-Type", "application/json")
		_, _ = rw.Write(data)
	})

	client := NewClient("user", "secret")
	client.baseURL = server.URL

	return client
}

func TestClient_GetRecords(t *testing.T) {
	expectedForm := `{"request":{"user":"user","auth":"xxx","command":"dns-rows-list","data":{"domain":"example.com"}}}`
	client := setupNew(t, expectedForm, commandDNSRowsList)

	records, err := client.GetRecords(t.Context(), "example.com.")
	require.NoError(t, err)

	assert.Len(t, records, 4)

	expected := []DNSRow{
		{
			ID:   "911",
			TTL:  "1800",
			Type: "A",
			Data: "1.2.3.4",
		},
		{
			ID:   "913",
			TTL:  "1800",
			Type: "MX",
			Data: "1 mail1.wedos.net",
		},
		{
			ID:   "914",
			TTL:  "1800",
			Type: "MX",
			Data: "10 mailbackup.wedos.net",
		},
		{
			ID:   "912",
			Name: "*",
			TTL:  "1800",
			Type: "A",
			Data: "1.2.3.4",
		},
	}

	assert.Equal(t, expected, records)
}

func TestClient_AddRecord(t *testing.T) {
	expectedForm := `{"request":{"user":"user","auth":"xxx","command":"dns-row-add","data":{"domain":"example.com","name":"foo","ttl":1800,"type":"TXT","rdata":"foobar"}}}`

	client := setupNew(t, expectedForm, commandDNSRowAdd)

	record := DNSRow{
		ID:   "",
		Name: "foo",
		TTL:  "1800",
		Type: "TXT",
		Data: "foobar",
	}

	err := client.AddRecord(t.Context(), "example.com.", record)
	require.NoError(t, err)
}

func TestClient_AddRecord_update(t *testing.T) {
	expectedForm := `{"request":{"user":"user","auth":"xxx","command":"dns-row-update","data":{"row_id":"1","domain":"example.com","ttl":1800,"type":"TXT","rdata":"foobar"}}}`

	client := setupNew(t, expectedForm, commandDNSRowUpdate)

	record := DNSRow{
		ID:   "1",
		Name: "foo",
		TTL:  "1800",
		Type: "TXT",
		Data: "foobar",
	}

	err := client.AddRecord(t.Context(), "example.com.", record)
	require.NoError(t, err)
}

func TestClient_DeleteRecord(t *testing.T) {
	expectedForm := `{"request":{"user":"user","auth":"xxx","command":"dns-row-delete","data":{"row_id":"1","domain":"example.com","rdata":""}}}`

	client := setupNew(t, expectedForm, commandDNSRowDelete)

	err := client.DeleteRecord(t.Context(), "example.com.", "1")
	require.NoError(t, err)
}

func TestClient_Commit(t *testing.T) {
	expectedForm := `{"request":{"user":"user","auth":"xxx","command":"dns-domain-commit","data":{"name":"example.com"}}}`

	client := setupNew(t, expectedForm, commandDNSDomainCommit)

	err := client.Commit(t.Context(), "example.com.")
	require.NoError(t, err)
}
