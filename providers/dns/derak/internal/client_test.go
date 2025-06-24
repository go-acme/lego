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

func setupTest(t *testing.T) (*Client, *http.ServeMux) {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := NewClient("secret")
	client.baseURL, _ = url.Parse(server.URL)
	client.zoneEndpoint = server.URL
	client.HTTPClient = server.Client()

	return client, mux
}

func testHandler(method string, statusCode int, filename string) func(rw http.ResponseWriter, req *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != method {
			http.Error(rw, fmt.Sprintf("unsupported method %s", req.Method), http.StatusBadRequest)
			return
		}

		username, password, ok := req.BasicAuth()
		if !ok {
			http.Error(rw, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		if username != "api" {
			http.Error(rw, fmt.Sprintf("username: want %s got %s", username, "user"), http.StatusUnauthorized)
			return
		}

		if password != "secret" {
			http.Error(rw, fmt.Sprintf("password: want %s got %s", password, "secret"), http.StatusUnauthorized)
			return
		}

		file, err := os.Open(filepath.Join("fixtures", filename))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		defer func() { _ = file.Close() }()

		rw.WriteHeader(statusCode)

		_, err = io.Copy(rw, file)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func TestGetRecords(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/zones/47c0ecf6c91243308c649ad1d2d618dd/dnsrecords",
		testHandler(http.MethodGet, http.StatusOK, "records-GET.json"))

	records, err := client.GetRecords(t.Context(), "47c0ecf6c91243308c649ad1d2d618dd", &GetRecordsParameters{DNSType: "TXT", Content: `"test"'`})
	require.NoError(t, err)

	excepted := &GetRecordsResponse{Data: []Record{
		{
			Type:    "A",
			Host:    "example.com",
			Content: "188.114.97.3",
			ID:      "812bee17a0b440b0bd5ee099a78b839c",
		},
		{
			Type:    "A",
			Host:    "example.com",
			Content: "188.114.96.3",
			ID:      "90e6029da45d4a36bf31056cf85d0cab",
		},
		{
			Type:    "AAAA",
			Host:    "example.com",
			Content: "2a06:98c1:3121::7",
			ID:      "0ac0320da0d24b5ca4f1648986a17340",
		},
		{
			Type:    "AAAA",
			Host:    "example.com",
			Content: "2a06:98c1:3120::7",
			ID:      "c91599694aea413498a0b3cd0a54a585",
		},
		{
			Type:    "A",
			Host:    "www",
			Content: "188.114.96.7",
			ID:      "c21f974992d549499f92e768bc468374",
		},
		{
			Type:    "A",
			Host:    "www",
			Content: "188.114.97.7",
			ID:      "90c3c1f05dca426893f10f122d18ad7a",
		},
		{
			Type:    "AAAA",
			Host:    "www",
			Content: "2a06:98c1:3121::",
			ID:      "379ab0ac0e434bc9aee5287e497f88a5",
		},
		{
			Type:    "AAAA",
			Host:    "www",
			Content: "2a06:98c1:3120::",
			ID:      "a1c4f9e50ba74791a4d70dc96999474c",
		},
	}, Count: 8}

	assert.Equal(t, excepted, records)
}

func TestGetRecords_error(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/zones/47c0ecf6c91243308c649ad1d2d618dd/dnsrecords",
		testHandler(http.MethodGet, http.StatusUnauthorized, "error.json"))

	_, err := client.GetRecords(t.Context(), "47c0ecf6c91243308c649ad1d2d618dd", &GetRecordsParameters{DNSType: "TXT", Content: `"test"'`})
	require.Error(t, err)
}

func TestGetRecord(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/zones/47c0ecf6c91243308c649ad1d2d618dd/dnsrecords/812bee17a0b440b0bd5ee099a78b839c",
		testHandler(http.MethodGet, http.StatusOK, "record-GET.json"))

	record, err := client.GetRecord(t.Context(), "47c0ecf6c91243308c649ad1d2d618dd", "812bee17a0b440b0bd5ee099a78b839c")
	require.NoError(t, err)

	excepted := &Record{
		Type:    "A",
		Host:    "example.com",
		Content: "188.114.97.3",
		ID:      "812bee17a0b440b0bd5ee099a78b839c",
	}

	assert.Equal(t, excepted, record)
}

func TestGetRecord_error(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/zones/47c0ecf6c91243308c649ad1d2d618dd/dnsrecords/812bee17a0b440b0bd5ee099a78b839c",
		testHandler(http.MethodGet, http.StatusUnauthorized, "error.json"))

	_, err := client.GetRecord(t.Context(), "47c0ecf6c91243308c649ad1d2d618dd", "812bee17a0b440b0bd5ee099a78b839c")
	require.Error(t, err)
}

func TestCreateRecord(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/zones/47c0ecf6c91243308c649ad1d2d618dd/dnsrecords",
		testHandler(http.MethodPut, http.StatusCreated, "record-PUT.json"))

	r := Record{
		Type:    "TXT",
		Host:    "test",
		Content: "test",
		TTL:     120,
	}

	record, err := client.CreateRecord(t.Context(), "47c0ecf6c91243308c649ad1d2d618dd", r)
	require.NoError(t, err)

	excepted := &Record{
		Type:    "A",
		Host:    "example.com",
		Content: "188.114.97.3",
		ID:      "812bee17a0b440b0bd5ee099a78b839c",
	}

	assert.Equal(t, excepted, record)
}

func TestCreateRecord_error(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/zones/47c0ecf6c91243308c649ad1d2d618dd/dnsrecords",
		testHandler(http.MethodPut, http.StatusUnauthorized, "error.json"))

	r := Record{
		Type:    "TXT",
		Host:    "test",
		Content: "test",
		TTL:     120,
	}

	_, err := client.CreateRecord(t.Context(), "47c0ecf6c91243308c649ad1d2d618dd", r)
	require.Error(t, err)
}

func TestEditRecord(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/zones/47c0ecf6c91243308c649ad1d2d618dd/dnsrecords/eebc813de2f94d67b09d91e10e2d65c2",
		testHandler(http.MethodPatch, http.StatusOK, "record-PATCH.json"))

	record, err := client.EditRecord(t.Context(), "47c0ecf6c91243308c649ad1d2d618dd", "eebc813de2f94d67b09d91e10e2d65c2", Record{
		Content: "foo",
	})
	require.NoError(t, err)

	excepted := &Record{
		Type:    "A",
		Host:    "example.com",
		Content: "188.114.97.3",
		ID:      "812bee17a0b440b0bd5ee099a78b839c",
	}

	assert.Equal(t, excepted, record)
}

func TestEditRecord_error(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/zones/47c0ecf6c91243308c649ad1d2d618dd/dnsrecords/eebc813de2f94d67b09d91e10e2d65c2",
		testHandler(http.MethodPatch, http.StatusUnauthorized, "error.json"))

	_, err := client.EditRecord(t.Context(), "47c0ecf6c91243308c649ad1d2d618dd", "eebc813de2f94d67b09d91e10e2d65c2", Record{
		Content: "foo",
	})
	require.Error(t, err)
}

func TestDeleteRecord(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/zones/47c0ecf6c91243308c649ad1d2d618dd/dnsrecords/653464211b7447a1bee6b8fcb9fb86df",
		testHandler(http.MethodDelete, http.StatusOK, "record-DELETE.json"))

	err := client.DeleteRecord(t.Context(), "47c0ecf6c91243308c649ad1d2d618dd", "653464211b7447a1bee6b8fcb9fb86df")
	require.NoError(t, err)
}

func TestDeleteRecord_error(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/zones/47c0ecf6c91243308c649ad1d2d618dd/dnsrecords/653464211b7447a1bee6b8fcb9fb86df",
		testHandler(http.MethodDelete, http.StatusUnauthorized, "error.json"))

	err := client.DeleteRecord(t.Context(), "47c0ecf6c91243308c649ad1d2d618dd", "653464211b7447a1bee6b8fcb9fb86df")
	require.Error(t, err)
}

func TestGetZones(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/", testHandler(http.MethodGet, http.StatusOK, "service-cdn-zones.json"))

	zones, err := client.GetZones(t.Context())
	require.NoError(t, err)

	excepted := []Zone{{
		ID:               "47c0ecf6c91243308c649ad1d2d618dd",
		Tags:             []string{},
		ContextID:        "47c0ecf6c91243308c649ad1d2d618dd",
		ContextType:      "CDN",
		HumanReadable:    "example.com",
		Serial:           "2301449956",
		CreationTime:     1679090659902,
		CreationTimeDate: time.Date(2023, time.March, 17, 22, 4, 19, 902000000, time.UTC),
		Status:           "active",
		IsMoved:          true,
		Paused:           false,
		ServiceType:      "CDN",
		Limbo:            false,
		TeamName:         "test",
		TeamID:           "640ef58496738d38fa7246a4",
		MyTeam:           true,
		RoleName:         "owner",
		IsBoard:          true,
		BoardRole:        []string{"owner"},
	}}

	assert.Equal(t, excepted, zones)
}

func TestGetZones_error(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/", testHandler(http.MethodGet, http.StatusUnauthorized, "error.json"))

	_, err := client.GetZones(t.Context())
	require.Error(t, err)
}
