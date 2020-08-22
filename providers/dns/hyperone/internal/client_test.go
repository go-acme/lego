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

func TestClient_FindRecordset(t *testing.T) {
	server := createTestServer(t, "GET", "/dns/loc123/project/proj123/zone/321/recordset", "recordset.json")
	client := getTestClient(t, server.URL)

	recordset, err := client.FindRecordset("321", "SOA", "example.com.")
	require.NoError(t, err)

	expected := &Recordset{
		ID:         "123456789abcd",
		Name:       "example.com.",
		RecordType: "SOA",
		TTL:        1800,
	}

	assert.Equal(t, expected, recordset)
}

func TestClient_GetRecords(t *testing.T) {
	server := createTestServer(t, "GET", "/dns/loc123/project/proj123/zone/321/recordset/322/record", "record.json")
	client := getTestClient(t, server.URL)

	records, err := client.GetRecords("321", "322")
	require.NoError(t, err)

	expected := []Record{
		{
			ID:      "135128352183572dd",
			Content: "pns.hyperone.com. hostmaster.hyperone.com. 1 15 180 1209600 1800",
			Enabled: true,
		},
	}

	assert.Equal(t, expected, records)
}

func TestClient_FindZone(t *testing.T) {
	server := createTestServer(t, "GET", "/dns/loc123/project/proj123/zone", "zones.json")
	client := getTestClient(t, server.URL)

	zone, err := client.FindZone("example.com")
	require.NoError(t, err)

	expected := &Zone{
		ID:      "zoneB",
		Name:    "example.com",
		DNSName: "example.com",
		FQDN:    "example.com.",
		URI:     "",
	}

	assert.Equal(t, expected, zone)
}

func TestClient_GetZones(t *testing.T) {
	server := createTestServer(t, "GET", "/dns/loc123/project/proj123/zone", "zones.json")
	client := getTestClient(t, server.URL)

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

func createTestServer(t *testing.T, method, path, fixtureName string) *httptest.Server {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)

	mux.Handle(path, http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != method {
			http.Error(rw, fmt.Sprintf("unsupported method: %s", req.Method), http.StatusMethodNotAllowed)
			return
		}

		file, err := os.Open(filepath.Join(".", "fixtures", fixtureName))
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

	t.Cleanup(server.Close)

	return server
}

func getTestClient(t *testing.T, serverURL string) *Client {
	passport := &Passport{
		SubjectID: "/iam/project/proj123/sa/xxxxxxx",
	}

	client, err := NewClient(serverURL, "loc123", passport)
	require.NoError(t, err)

	client.signer = signerMock{}

	return client
}
