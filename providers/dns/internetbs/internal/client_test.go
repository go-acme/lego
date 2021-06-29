package internal

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testBaseURL = "https://testapi.internet.bs"

const (
	testAPIKey   = "testapi"
	testPassword = "testpass"
)

func TestClient_AddRecord(t *testing.T) {
	client := setupTest(t, "/Domain/DnsRecord/Add", "./fixtures/Domain_DnsRecord_Add_SUCCESS.json")

	query := RecordQuery{
		FullRecordName: "www.example.com",
		Type:           "TXT",
		Value:          "xxx",
		TTL:            36000,
	}

	err := client.AddRecord(query)
	require.NoError(t, err)
}

func TestClient_AddRecord_error(t *testing.T) {
	client := setupTest(t, "/Domain/DnsRecord/Add", "./fixtures/Domain_DnsRecord_Add_FAILURE.json")

	query := RecordQuery{
		FullRecordName: "www.example.com.",
		Type:           "TXT",
		Value:          "xxx",
		TTL:            36000,
	}

	err := client.AddRecord(query)
	require.Error(t, err)
}

func TestClient_AddRecord_integration(t *testing.T) {
	env, ok := os.LookupEnv("INTERNET_BS_DEBUG")
	if !ok {
		t.Skip("skip integration test")
	}

	client := NewClient(testAPIKey, testPassword)
	client.baseURL, _ = url.Parse(testBaseURL)
	client.debug, _ = strconv.ParseBool(env)

	query := RecordQuery{
		FullRecordName: "www.example.com",
		Type:           "TXT",
		Value:          "xxx",
		TTL:            36000,
	}

	err := client.AddRecord(query)
	require.NoError(t, err)

	query = RecordQuery{
		FullRecordName: "www.example.com",
		Type:           "TXT",
		Value:          "yyy",
		TTL:            36000,
	}

	err = client.AddRecord(query)
	require.NoError(t, err)
}

func TestClient_RemoveRecord(t *testing.T) {
	client := setupTest(t, "/Domain/DnsRecord/Remove", "./fixtures/Domain_DnsRecord_Remove_SUCCESS.json")

	query := RecordQuery{
		FullRecordName: "www.example.com",
		Type:           "TXT",
		Value:          "",
	}
	err := client.RemoveRecord(query)
	require.NoError(t, err)
}

func TestClient_RemoveRecord_error(t *testing.T) {
	client := setupTest(t, "/Domain/DnsRecord/Remove", "./fixtures/Domain_DnsRecord_Remove_FAILURE.json")

	query := RecordQuery{
		FullRecordName: "www.example.com.",
		Type:           "TXT",
		Value:          "",
	}
	err := client.RemoveRecord(query)
	require.Error(t, err)
}

func TestClient_RemoveRecord_integration(t *testing.T) {
	env, ok := os.LookupEnv("INTERNET_BS_DEBUG")
	if !ok {
		t.Skip("skip integration test")
	}

	client := NewClient(testAPIKey, testPassword)
	client.baseURL, _ = url.Parse(testBaseURL)
	client.debug, _ = strconv.ParseBool(env)

	query := RecordQuery{
		FullRecordName: "www.example.com",
		Type:           "TXT",
		Value:          "",
	}

	err := client.RemoveRecord(query)
	require.NoError(t, err)
}

func TestClient_ListRecords(t *testing.T) {
	client := setupTest(t, "/Domain/DnsRecord/List", "./fixtures/Domain_DnsRecord_List_SUCCESS.json")

	query := ListRecordQuery{
		Domain: "example.com",
	}

	records, err := client.ListRecords(query)
	require.NoError(t, err)

	expected := []Record{
		{
			Name:  "example.com",
			Value: "ns-hongkong.internet.bs",
			TTL:   3600,
			Type:  "NS",
		},
		{
			Name:  "example.com",
			Value: "ns-toronto.internet.bs",
			TTL:   3600,
			Type:  "NS",
		},
		{
			Name:  "example.com",
			Value: "ns-london.internet.bs",
			TTL:   3600,
			Type:  "NS",
		},
		{
			Name:  "test.example.com",
			Value: "example1.com",
			TTL:   3600,
			Type:  "CNAME",
		},
		{
			Name:  "www.example.com",
			Value: "xxx",
			TTL:   36000,
			Type:  "TXT",
		},
		{
			Name:  "www.example.com",
			Value: "yyy",
			TTL:   36000,
			Type:  "TXT",
		},
	}

	assert.Equal(t, expected, records)
}

func TestClient_ListRecords_error(t *testing.T) {
	client := setupTest(t, "/Domain/DnsRecord/List", "./fixtures/Domain_DnsRecord_List_FAILURE.json")

	query := ListRecordQuery{
		Domain: "www.example.com",
	}

	_, err := client.ListRecords(query)
	require.Error(t, err)
}

func TestClient_ListRecords_integration(t *testing.T) {
	env, ok := os.LookupEnv("INTERNET_BS_DEBUG")
	if !ok {
		t.Skip("skip integration test")
	}

	client := NewClient(testAPIKey, testPassword)
	client.baseURL, _ = url.Parse(testBaseURL)
	client.debug, _ = strconv.ParseBool(env)

	query := ListRecordQuery{
		Domain: "example.com",
	}

	records, err := client.ListRecords(query)
	require.NoError(t, err)

	for _, record := range records {
		fmt.Println(record)
	}
}

func setupTest(t *testing.T, path, filename string) *Client {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	mux.HandleFunc(path, testHandler(filename))

	client := NewClient(testAPIKey, testPassword)
	client.baseURL, _ = url.Parse(server.URL)

	return client
}

func testHandler(filename string) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(rw, fmt.Sprintf("unsupported method: %s", req.Method), http.StatusMethodNotAllowed)
			return
		}

		if req.FormValue("apiKey") != testAPIKey {
			http.Error(rw, `{"transactid":"d46d812569acdb8b39c3933ec4351e79","status":"FAILURE","message":"Invalid API key and\/or Password","code":107002}`, http.StatusOK)
			return
		}

		if req.FormValue("password") != testPassword {
			http.Error(rw, `{"transactid":"d46d812569acdb8b39c3933ec4351e79","status":"FAILURE","message":"Invalid API key and\/or Password","code":107002}`, http.StatusOK)
			return
		}

		file, err := os.Open(filename)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		defer func() { _ = file.Close() }()

		_, err = io.Copy(rw, file)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
