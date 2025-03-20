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

func TestClient_GetDNSSettings(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	mux.HandleFunc("/", testHandler("get_dns_settings.xml"))

	client := NewClient("user")
	client.baseURL = server.URL

	records, err := client.GetDNSSettings(mockContext(), "example.com", "")
	require.NoError(t, err)

	expected := []ReturnInfo{
		{
			ID:         "57297429",
			Zone:       "example.org",
			Name:       "",
			Type:       "A",
			Data:       "10.0.0.1",
			Changeable: "Y",
			Aux:        0,
		},
		{
			ID:         int64(0),
			Zone:       "example.org",
			Name:       "",
			Type:       "NS",
			Data:       "ns5.kasserver.com.",
			Changeable: "N",
			Aux:        0,
		},
		{
			ID:         int64(0),
			Zone:       "example.org",
			Name:       "",
			Type:       "NS",
			Data:       "ns6.kasserver.com.",
			Changeable: "N",
			Aux:        0,
		},
		{
			ID:         "57297479",
			Zone:       "example.org",
			Name:       "*",
			Type:       "A",
			Data:       "10.0.0.1",
			Changeable: "Y",
			Aux:        0,
		},
		{
			ID:         "57297481",
			Zone:       "example.org",
			Name:       "",
			Type:       "MX",
			Data:       "user.kasserver.com.",
			Changeable: "Y",
			Aux:        10,
		},
		{
			ID:         "57297483",
			Zone:       "example.org",
			Name:       "",
			Type:       "TXT",
			Data:       "v=spf1 mx a ?all",
			Changeable: "Y",
			Aux:        0,
		},
		{
			ID:         "57297485",
			Zone:       "example.org",
			Name:       "_dmarc",
			Type:       "TXT",
			Data:       "v=DMARC1; p=none;",
			Changeable: "Y",
			Aux:        0,
		},
	}

	assert.Equal(t, expected, records)
}

func TestClient_AddDNSSettings(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	mux.HandleFunc("/", testHandler("add_dns_settings.xml"))

	client := NewClient("user")
	client.baseURL = server.URL

	record := DNSRequest{
		ZoneHost:   "42cnc.de.",
		RecordType: "TXT",
		RecordName: "lego",
		RecordData: "abcdefgh",
	}

	recordID, err := client.AddDNSSettings(mockContext(), record)
	require.NoError(t, err)

	assert.Equal(t, "57347444", recordID)
}

func TestClient_DeleteDNSSettings(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	mux.HandleFunc("/", testHandler("delete_dns_settings.xml"))

	client := NewClient("user")
	client.baseURL = server.URL

	r, err := client.DeleteDNSSettings(mockContext(), "57347450")
	require.NoError(t, err)

	assert.Equal(t, "TRUE", r)
}

func testHandler(filename string) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(rw, fmt.Sprintf("unsupported method: %s", req.Method), http.StatusMethodNotAllowed)
			return
		}

		file, err := os.Open(filepath.Join("fixtures", filename))
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
