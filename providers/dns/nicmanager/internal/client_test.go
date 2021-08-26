package internal

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_ZoneInfo(t *testing.T) {
	client := setupTest(t, "/anycast/nicmanager-anycastdns4.net", testHandler(http.MethodGet, http.StatusOK, "zone.json"))

	zone, err := client.ZoneInfo("nicmanager-anycastdns4.net")
	require.NoError(t, err)

	expected := &Zone{
		Name:   "nicmanager-anycastdns4.net",
		Active: true,
		Records: []Record{
			{
				ID:      186,
				Name:    "nicmanager-anycastdns4.net",
				Type:    "A",
				Content: "123.123.123.123",
				TTL:     3600,
			},
		},
	}

	assert.Equal(t, expected, zone)
}

func TestClient_ZoneInfo_error(t *testing.T) {
	client := setupTest(t, "/anycast/foo", testHandler(http.MethodGet, http.StatusNotFound, "error.json"))

	_, err := client.ZoneInfo("foo")
	require.Error(t, err)
}

func TestClient_AddRecord(t *testing.T) {
	client := setupTest(t, "/anycast/zonedomain.tld/records", testHandler(http.MethodPost, http.StatusAccepted, "error.json"))

	record := RecordCreateUpdate{
		Type:  "TXT",
		Name:  "lego",
		Value: "content",
		TTL:   3600,
	}

	err := client.ResourceRecordCreate("zonedomain.tld", record)
	require.NoError(t, err)
}

func TestClient_AddRecord_error(t *testing.T) {
	client := setupTest(t, "/anycast/zonedomain.tld", testHandler(http.MethodPost, http.StatusUnauthorized, "error.json"))

	record := RecordCreateUpdate{
		Type:  "TXT",
		Name:  "zonedomain.tld",
		Value: "content",
		TTL:   3600,
	}

	err := client.ResourceRecordCreate("zonedomain.tld", record)
	require.Error(t, err)
}

func TestClient_DeleteRecord(t *testing.T) {
	client := setupTest(t, "/anycast/zonedomain.tld/records/6", testHandler(http.MethodDelete, http.StatusAccepted, "error.json"))

	err := client.ResourceRecordDelete("zonedomain.tld", 6)
	require.NoError(t, err)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client := setupTest(t, "/anycast/zonedomain.tld/records/6", testHandler(http.MethodDelete, http.StatusNoContent, ""))

	err := client.ResourceRecordDelete("zonedomain.tld", 7)
	require.Error(t, err)
}

func setupTest(t *testing.T, path string, handler http.Handler) *NicManagerClient {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	mux.Handle(path, handler)

	client := NewNicManagerClient(&http.Client{})
	client.SetAccount("foo", "bar")
	client.SetOTP("2hsn")
	client.Password = "foo"
	client.baseURL = server.URL

	return client
}

func testHandler(method string, statusCode int, filename string) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != method {
			http.Error(rw, fmt.Sprintf(`{"message":"unsupported method: %s"}`, req.Method), http.StatusMethodNotAllowed)
			return
		}

		username, password, ok := req.BasicAuth()
		if !ok || username != "foo.bar" || password != "foo" {
			http.Error(rw, `{"message":"Unauthenticated"}`, http.StatusUnauthorized)
			return
		}

		rw.WriteHeader(statusCode)

		if statusCode == http.StatusNoContent {
			return
		}

		file, err := os.Open(filepath.Join("fixtures", filename))
		if err != nil {
			http.Error(rw, fmt.Sprintf(`{"message":"%v"}`, err), http.StatusInternalServerError)
			return
		}

		defer func() { _ = file.Close() }()

		_, err = io.Copy(rw, file)
		if err != nil {
			http.Error(rw, fmt.Sprintf(`{"message":"%v"}`, err), http.StatusInternalServerError)
			return
		}
	}
}
