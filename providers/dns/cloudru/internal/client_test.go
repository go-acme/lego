package internal

import (
	"fmt"
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

func setupTest(t *testing.T, pattern string, handler http.HandlerFunc) *Client {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	mux.HandleFunc(pattern, handler)

	client := NewClient("user", "secret")
	client.HTTPClient = server.Client()
	client.APIEndpoint, _ = url.Parse(server.URL)
	client.token = &Token{
		AccessToken: "secret",
		ExpiresIn:   60,
		TokenType:   "Bearer",
		Deadline:    time.Now().Add(1 * time.Minute),
	}

	return client
}

func writeFixtureHandler(method, filename string) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != method {
			http.Error(rw, fmt.Sprintf("unsupported method %s", req.Method), http.StatusBadRequest)
			return
		}

		file, err := os.Open(filepath.Join("fixtures", filename))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		defer func() { _ = file.Close() }()

		_, _ = io.Copy(rw, file)
	}
}

func TestClient_GetZones(t *testing.T) {
	client := setupTest(t, "/zones", writeFixtureHandler(http.MethodGet, "zones.json"))

	ctx := mockContext()

	zones, err := client.GetZones(ctx, "xxx")
	require.NoError(t, err)

	expected := []Zone{
		{
			ID:        "59556fcd-95ff-451f-b49b-9732f21f944a",
			ParentID:  "2d7b6194-2b83-4f71-86fd-a1e727e347b2",
			Name:      "example.com",
			Valid:     true,
			Delegated: true,
			CreatedAt: time.Date(2023, 7, 23, 8, 12, 41, 0, time.UTC),
			UpdatedAt: time.Date(2023, 7, 24, 5, 50, 28, 0, time.UTC),
		},
	}
	assert.Equal(t, expected, zones)
}

func TestClient_GetRecords(t *testing.T) {
	client := setupTest(t, "/zones/zzz/records", writeFixtureHandler(http.MethodGet, "records.json"))

	ctx := mockContext()

	records, err := client.GetRecords(ctx, "zzz")
	require.NoError(t, err)

	expected := []Record{
		{
			ZoneID: "59556fcd-95ff-451f-b49b-9732f21f944a",
			Name:   "example.com.",
			Type:   "SOA",
			Values: []string{
				"cdns-ns01.sbercloud.ru. mail.sbercloud.ru 1 120 3600 604800 3600",
			},
			TTL:     "3600",
			Enables: true,
		},
		{
			ZoneID: "59556fcd-95ff-451f-b49b-9732f21f944a",
			Name:   "example.com.",
			Type:   "NS",
			Values: []string{
				"cdns-ns01.sbercloud.ru.",
				"cdns-ns02.sbercloud.ru.",
			},
			TTL:     "3600",
			Enables: true,
		},
		{
			ZoneID: "59556fcd-95ff-451f-b49b-9732f21f944a",
			Name:   "www.example.com.",
			Type:   "A",
			Values: []string{
				"8.8.8.8",
			},
			TTL:     "3600",
			Enables: true,
		},
	}
	assert.Equal(t, expected, records)
}

func TestClient_CreateRecord(t *testing.T) {
	client := setupTest(t, "/zones/zzz/records", writeFixtureHandler(http.MethodPost, "record.json"))

	ctx := mockContext()

	recordReq := Record{
		Name:   "www.example.com.",
		Type:   "TXT",
		Values: []string{"text"},
		TTL:    "3600",
	}

	record, err := client.CreateRecord(ctx, "zzz", recordReq)
	require.NoError(t, err)

	expected := &Record{
		ZoneID: "59556fcd-95ff-451f-b49b-9732f21f944a",
		Name:   "www.example.com.",
		Type:   "TXT",
		Values: []string{
			"txt",
		},
		TTL:     "3600",
		Enables: true,
	}
	assert.Equal(t, expected, record)
}

func TestClient_DeleteRecord(t *testing.T) {
	client := setupTest(t, "/zones/zzz/records/example.com/TXT", writeFixtureHandler(http.MethodDelete, "record.json"))

	ctx := mockContext()

	err := client.DeleteRecord(ctx, "zzz", "example.com", "TXT")
	require.NoError(t, err)
}
