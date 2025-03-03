package internal

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

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

	client, err := NewClient("secret", "shortname")
	require.NoError(t, err)

	client.HTTPClient = server.Client()
	client.baseURL, _ = url.Parse(server.URL)

	return client
}

func TestClient_Create(t *testing.T) {
	client := setupTest(t, "POST /api/config/dns/namespaces/system/dns_zones/example.com/rrsets/groupA", http.StatusOK, "create.json")

	rrSet := RRSet{
		Description: "lego",
		TTL:         60,
		TXTRecord: &TXTRecord{
			Name:   "wwww",
			Values: []string{"txt"},
		},
	}

	result, err := client.CreateRRSet(context.Background(), "example.com", "groupA", rrSet)
	require.NoError(t, err)

	expected := &APIRRSet{
		DNSZoneName: "string",
		GroupName:   "string",
		RRSet: RRSet{
			Description: "string",
			TXTRecord: &TXTRecord{
				Name:   "string",
				Values: []string{"string"},
			},
		},
	}

	assert.Equal(t, expected, result)
}

func TestClient_Create_error(t *testing.T) {
	client := setupTest(t, "POST /api/config/dns/namespaces/system/dns_zones/example.com/rrsets/groupA", http.StatusBadRequest, "")

	rrSet := RRSet{
		Description: "lego",
		TTL:         60,
		TXTRecord: &TXTRecord{
			Name:   "wwww",
			Values: []string{"txt"},
		},
	}

	_, err := client.CreateRRSet(context.Background(), "example.com", "groupA", rrSet)
	require.Error(t, err)
}

func TestClient_Get(t *testing.T) {
	client := setupTest(t, "GET /api/config/dns/namespaces/system/dns_zones/example.com/rrsets/groupA/www/TXT", http.StatusOK, "get.json")

	result, err := client.GetRRSet(context.Background(), "example.com", "groupA", "www", "TXT")
	require.NoError(t, err)

	expected := &APIRRSet{
		DNSZoneName: "string",
		GroupName:   "string",
		Namespace:   "string",
		RecordName:  "string",
		Type:        "string",
		RRSet: RRSet{
			Description: "string",
			TXTRecord: &TXTRecord{
				Name:   "string",
				Values: []string{"string"},
			},
		},
	}

	assert.Equal(t, expected, result)
}

func TestClient_Get_not_found(t *testing.T) {
	client := setupTest(t, "GET /api/config/dns/namespaces/system/dns_zones/example.com/rrsets/groupA/www/TXT", http.StatusNotFound, "error_404.json")

	result, err := client.GetRRSet(context.Background(), "example.com", "groupA", "www", "TXT")
	require.NoError(t, err)

	assert.Nil(t, result)
}

func TestClient_Get_error(t *testing.T) {
	client := setupTest(t, "GET /api/config/dns/namespaces/system/dns_zones/example.com/rrsets/groupA/www/TXT", http.StatusBadRequest, "")

	_, err := client.GetRRSet(context.Background(), "example.com", "groupA", "www", "TXT")
	require.Error(t, err)
}

func TestClient_Delete(t *testing.T) {
	client := setupTest(t, "DELETE /api/config/dns/namespaces/system/dns_zones/example.com/rrsets/groupA/www/TXT", http.StatusOK, "get.json")

	result, err := client.DeleteRRSet(context.Background(), "example.com", "groupA", "www", "TXT")
	require.NoError(t, err)

	expected := &APIRRSet{
		DNSZoneName: "string",
		GroupName:   "string",
		Namespace:   "string",
		RecordName:  "string",
		Type:        "string",
		RRSet: RRSet{
			Description: "string",
			TXTRecord: &TXTRecord{
				Name:   "string",
				Values: []string{"string"},
			},
		},
	}

	assert.Equal(t, expected, result)
}

func TestClient_Delete_error(t *testing.T) {
	client := setupTest(t, "DELETE /api/config/dns/namespaces/system/dns_zones/example.com/rrsets/groupA/www/TXT", http.StatusBadRequest, "")

	_, err := client.DeleteRRSet(context.Background(), "example.com", "groupA", "www", "TXT")
	require.Error(t, err)
}

func TestClient_Replace(t *testing.T) {
	client := setupTest(t, "PUT /api/config/dns/namespaces/system/dns_zones/example.com/rrsets/groupA/www/TXT", http.StatusOK, "get.json")

	rrSet := RRSet{
		Description: "lego",
		TTL:         60,
		TXTRecord: &TXTRecord{
			Name:   "wwww",
			Values: []string{"txt"},
		},
	}

	result, err := client.ReplaceRRSet(context.Background(), "example.com", "groupA", "www", "TXT", rrSet)
	require.NoError(t, err)

	expected := &APIRRSet{
		DNSZoneName: "string",
		GroupName:   "string",
		Namespace:   "string",
		RecordName:  "string",
		Type:        "string",
		RRSet: RRSet{
			Description: "string",
			TXTRecord: &TXTRecord{
				Name:   "string",
				Values: []string{"string"},
			},
		},
	}

	assert.Equal(t, expected, result)
}

func TestClient_Replace_error(t *testing.T) {
	client := setupTest(t, "PUT /api/config/dns/namespaces/system/dns_zones/example.com/rrsets/groupA/www/TXT", http.StatusBadRequest, "")

	rrSet := RRSet{
		Description: "lego",
		TTL:         60,
		TXTRecord: &TXTRecord{
			Name:   "wwww",
			Values: []string{"txt"},
		},
	}

	_, err := client.ReplaceRRSet(context.Background(), "example.com", "groupA", "www", "TXT", rrSet)
	require.Error(t, err)
}
