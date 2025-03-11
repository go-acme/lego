package active24

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T, pattern string, status int, filename string) *Client {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	mux.HandleFunc(pattern, func(rw http.ResponseWriter, req *http.Request) {
		if filename == "" {
			rw.WriteHeader(status)
			return
		}

		file, err := os.Open(filepath.Join("fixtures", filename))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		defer func() { _ = file.Close() }()

		rw.WriteHeader(status)
		_, err = io.Copy(rw, file)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	client, err := NewClient("example.com", "user", "secret")
	require.NoError(t, err)

	client.HTTPClient = server.Client()
	client.baseURL, _ = url.Parse(server.URL)

	return client
}

func TestClient_GetServices(t *testing.T) {
	client := setupTest(t, "GET /v1/user/self/service", http.StatusOK, "services.json")

	services, err := client.GetServices(context.Background())
	require.NoError(t, err)

	expected := []Service{
		{
			ID:          1111,
			ServiceName: ".sk doména",
			Status:      "active",
			Name:        "mydomain.sk",
			CreateTime:  1374357600,
			ExpireTime:  1405914526,
			Price:       12.3,
		},
		{
			ID:          2222,
			ServiceName: "The Hosting",
			Status:      "active",
			Name:        "myname_1",
			CreateTime:  1400145443,
			ExpireTime:  1431702371,
			Price:       55.2,
		},
	}

	assert.Equal(t, expected, services)
}

func TestClient_GetServices_errors(t *testing.T) {
	client := setupTest(t, "GET /v1/user/self/service", http.StatusUnauthorized, "error_v1.json")

	_, err := client.GetServices(context.Background())
	require.EqualError(t, err, "401: No username or password.")
}

func TestClient_GetRecords(t *testing.T) {
	client := setupTest(t, "GET /v2/service/aaa/dns/record", http.StatusOK, "records.json")

	filter := RecordFilter{
		Name:    "example.com",
		Type:    []string{"TXT"},
		Content: "txt",
	}

	records, err := client.GetRecords(context.Background(), "aaa", filter)
	require.NoError(t, err)

	expected := []Record{{
		ID:       13,
		Name:     "string",
		Content:  "string",
		TTL:      120,
		Priority: 1,
		Port:     443,
		Weight:   50,
	}}

	assert.Equal(t, expected, records)
}

func TestClient_GetRecords_errors(t *testing.T) {
	client := setupTest(t, "GET /v2/service/aaa/dns/record", http.StatusForbidden, "error_403.json")

	filter := RecordFilter{
		Name:    "example.com",
		Type:    []string{"TXT"},
		Content: "txt",
	}

	_, err := client.GetRecords(context.Background(), "aaa", filter)
	require.EqualError(t, err, "403: /errors/httpException: This action is unauthorized.")
}

func TestClient_CreateRecord(t *testing.T) {
	client := setupTest(t, "POST /v2/service/aaa/dns/record", http.StatusNoContent, "")

	err := client.CreateRecord(context.Background(), "aaa", Record{})
	require.NoError(t, err)
}

func TestClient_CreateRecord_errors(t *testing.T) {
	client := setupTest(t, "POST /v2/service/aaa/dns/record", http.StatusForbidden, "error_403.json")

	err := client.CreateRecord(context.Background(), "aaa", Record{})
	require.EqualError(t, err, "403: /errors/httpException: This action is unauthorized.")
}

func TestClient_DeleteRecord(t *testing.T) {
	client := setupTest(t, "DELETE /v2/service/aaa/dns/record/123", http.StatusNoContent, "")

	err := client.DeleteRecord(context.Background(), "aaa", "123")
	require.NoError(t, err)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client := setupTest(t, "DELETE /v2/service/aaa/dns/record/123", http.StatusForbidden, "error_403.json")

	err := client.DeleteRecord(context.Background(), "aaa", "123")
	require.EqualError(t, err, "403: /errors/httpException: This action is unauthorized.")
}

func TestClient_sign(t *testing.T) {
	client, err := NewClient("example.com", "user", "secret")
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodGet, "/v1/user/self/service", nil)
	require.NoError(t, err)

	err = client.sign(req, time.Date(2025, 6, 28, 1, 2, 3, 4, time.UTC))
	require.NoError(t, err)

	username, password, ok := req.BasicAuth()
	require.True(t, ok)

	assert.Equal(t, "user", username)
	assert.Equal(t, "743e2257421b260ed561f3e7af4b035414636393", password)
}
