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

type signerMock struct{}

func (s signerMock) GetJWT() (string, error) {
	return "", nil
}

func TestClient_GetZones(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	mux.Handle("/dns/loc123/project/proj123/zone", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(rw, fmt.Sprintf("unsupported method: %s", req.Method), http.StatusMethodNotAllowed)
			return
		}

		file, err := os.Open(filepath.Join(".", "fixtures", "zones.json"))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = io.Copy(rw, file)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	}))

	passport := &Passport{
		SubjectID: "/iam/project/proj123/sa/xxxxxxx",
	}

	client, err := NewClient(server.URL, "loc123", passport)
	require.NoError(t, err)

	client.signer = signerMock{}

	zones, err := client.GetZones()
	require.NoError(t, err)

	expected := []Zone{
		{
			ID:      "zoneA",
			Name:    "example.org",
			DNSName: "example.org",
			FQDN:    "example.org.",
			URI:     "",
		},
		{
			ID:      "zoneB",
			Name:    "example.com",
			DNSName: "example.com",
			FQDN:    "example.com.",
			URI:     "",
		},
	}

	assert.Equal(t, expected, zones)
}
