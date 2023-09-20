package hostingde

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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

func setupTest(t *testing.T, pattern string, handler http.HandlerFunc) *Client {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := NewClient("secret")
	client.HTTPClient = server.Client()
	client.BaseURL, _ = url.Parse(server.URL)

	mux.HandleFunc(pattern, handler)

	return client
}

func writeFixture(rw http.ResponseWriter, filename string) {
	file, err := os.Open(filepath.Join("fixtures", filename))
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	defer func() { _ = file.Close() }()

	_, _ = io.Copy(rw, file)
}

func TestClient_ListZoneConfigs(t *testing.T) {
	client := setupTest(t, "/zoneConfigsFind", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(rw, fmt.Sprintf("unsupported method %s", req.Method), http.StatusBadRequest)
			return
		}

		raw, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		body := string(bytes.TrimSpace(raw))
		if body != `{"authToken":"secret","filter":{"field":"zoneName","value":"example.com"},"limit":1,"page":1}` {
			http.Error(rw, fmt.Sprintf("unexpected body: got %s", body), http.StatusBadRequest)
			return
		}

		writeFixture(rw, "zoneConfigsFind.json")
	})

	zonesFind := ZoneConfigsFindRequest{
		Filter: Filter{Field: "zoneName", Value: "example.com"},
		Limit:  1,
		Page:   1,
	}

	zoneResponse, err := client.ListZoneConfigs(context.Background(), zonesFind)
	require.NoError(t, err)

	expected := &ZoneResponse{
		Limit:        10,
		Page:         1,
		TotalEntries: 15,
		TotalPages:   2,
		Type:         "FindZoneConfigsResult",
		Data: []ZoneConfig{{
			ID:                    "123",
			AccountID:             "456",
			Status:                "s",
			Name:                  "n",
			NameUnicode:           "u",
			MasterIP:              "m",
			Type:                  "t",
			EMailAddress:          "e",
			ZoneTransferWhitelist: []string{"a", "b"},
			LastChangeDate:        "l",
			DNSServerGroupID:      "g",
			DNSSecMode:            "m",
			SOAValues: &SOAValues{
				Refresh:     1,
				Retry:       2,
				Expire:      3,
				TTL:         4,
				NegativeTTL: 5,
			},
			TemplateValues: json.RawMessage(nil),
		}},
	}

	assert.Equal(t, expected, zoneResponse)
}

func TestClient_ListZoneConfigs_error(t *testing.T) {
	client := setupTest(t, "/zoneConfigsFind", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(rw, fmt.Sprintf("unsupported method %s", req.Method), http.StatusBadRequest)
			return
		}

		writeFixture(rw, "zoneConfigsFind_error.json")
	})

	zonesFind := ZoneConfigsFindRequest{
		Filter: Filter{Field: "zoneName", Value: "example.com"},
		Limit:  1,
		Page:   1,
	}

	_, err := client.ListZoneConfigs(context.Background(), zonesFind)
	require.Error(t, err)
}

func TestClient_UpdateZone(t *testing.T) {
	client := setupTest(t, "/zoneUpdate", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(rw, fmt.Sprintf("unsupported method %s", req.Method), http.StatusBadRequest)
			return
		}

		raw, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		body := string(bytes.TrimSpace(raw))
		if body != `{"authToken":"secret","zoneConfig":{"id":"123","accountId":"456","status":"s","name":"n","nameUnicode":"u","masterIp":"m","type":"t","emailAddress":"e","zoneTransferWhitelist":["a","b"],"lastChangeDate":"l","dnsServerGroupId":"g","dnsSecMode":"m","soaValues":{"refresh":1,"retry":2,"expire":3,"ttl":4,"negativeTtl":5}},"recordsToAdd":null,"recordsToDelete":[{"name":"_acme-challenge.example.com","type":"TXT","content":"\"txt\""}]}` {
			http.Error(rw, fmt.Sprintf("unexpected body: got %s", body), http.StatusBadRequest)
			return
		}

		writeFixture(rw, "zoneUpdate.json")
	})

	request := ZoneUpdateRequest{
		ZoneConfig: ZoneConfig{
			ID:                    "123",
			AccountID:             "456",
			Status:                "s",
			Name:                  "n",
			NameUnicode:           "u",
			MasterIP:              "m",
			Type:                  "t",
			EMailAddress:          "e",
			ZoneTransferWhitelist: []string{"a", "b"},
			LastChangeDate:        "l",
			DNSServerGroupID:      "g",
			DNSSecMode:            "m",
			SOAValues: &SOAValues{
				Refresh:     1,
				Retry:       2,
				Expire:      3,
				TTL:         4,
				NegativeTTL: 5,
			},
		},
		RecordsToDelete: []DNSRecord{{
			Type:    "TXT",
			Name:    "_acme-challenge.example.com",
			Content: `"txt"`,
		}},
	}

	response, err := client.UpdateZone(context.Background(), request)
	require.NoError(t, err)

	expected := &Zone{
		Records: []DNSRecord{{
			ID:               "123",
			ZoneID:           "456",
			RecordTemplateID: "789",
			Name:             "n",
			Type:             "TXT",
			Content:          "txt",
			TTL:              120,
			Priority:         5,
			LastChangeDate:   "d",
		}},
		ZoneConfig: ZoneConfig{
			ID:                    "123",
			AccountID:             "456",
			Status:                "s",
			Name:                  "n",
			NameUnicode:           "u",
			MasterIP:              "m",
			Type:                  "t",
			EMailAddress:          "e",
			ZoneTransferWhitelist: []string{"a", "b"},
			LastChangeDate:        "l",
			DNSServerGroupID:      "g",
			DNSSecMode:            "m",
			SOAValues: &SOAValues{
				Refresh:     1,
				Retry:       2,
				Expire:      3,
				TTL:         4,
				NegativeTTL: 5,
			},
		},
	}

	assert.Equal(t, expected, response)
}

func TestClient_UpdateZone_error(t *testing.T) {
	client := setupTest(t, "/zoneUpdate", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(rw, fmt.Sprintf("unsupported method %s", req.Method), http.StatusBadRequest)
			return
		}

		writeFixture(rw, "zoneUpdate_error.json")
	})

	request := ZoneUpdateRequest{
		ZoneConfig: ZoneConfig{
			ID:                    "123",
			AccountID:             "456",
			Status:                "s",
			Name:                  "n",
			NameUnicode:           "u",
			MasterIP:              "m",
			Type:                  "t",
			EMailAddress:          "e",
			ZoneTransferWhitelist: []string{"a", "b"},
			LastChangeDate:        "l",
			DNSServerGroupID:      "g",
			DNSSecMode:            "m",
			SOAValues: &SOAValues{
				Refresh:     1,
				Retry:       2,
				Expire:      3,
				TTL:         4,
				NegativeTTL: 5,
			},
		},
		RecordsToDelete: []DNSRecord{{
			Type:    "TXT",
			Name:    "_acme-challenge.example.com",
			Content: `"txt"`,
		}},
	}

	_, err := client.UpdateZone(context.Background(), request)
	require.Error(t, err)
}
