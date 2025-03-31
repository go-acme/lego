package internal

import (
	"context"
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

	mux.HandleFunc(pattern, handler)

	client, err := NewClient(server.Client())
	require.NoError(t, err)

	client.baseURL, _ = url.Parse(server.URL)

	return client
}

func writeFixtures(method string, filename string, status int) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != method {
			http.Error(rw, fmt.Sprintf("unsupported method: %s", req.Method), http.StatusMethodNotAllowed)
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
	}
}

func TestClient_GetServices(t *testing.T) {
	client := setupTest(t, "/services",
		writeFixtures(http.MethodGet, "services_GET.xml", http.StatusOK))

	zones, err := client.GetServices(context.Background())
	require.NoError(t, err)

	expected := []Service{
		{
			Admin:        "123/NIC-REG",
			DomainsLimit: "12",
			DomainsNum:   "5",
			Enable:       "true",
			HasPrimary:   "false",
			Name:         "testservice",
			Payer:        "123/NIC-REG",
			Tariff:       "Secondary L",
		},
		{
			Admin:        "123/NIC-REG",
			DomainsLimit: "150",
			DomainsNum:   "10",
			Enable:       "true",
			HasPrimary:   "true",
			Name:         "myservice",
			Payer:        "123/NIC-REG",
			Tariff:       "DNS-master XXL",
			RRLimit:      "7500",
			RRNum:        "1000",
		},
	}

	assert.Equal(t, expected, zones)
}

func TestClient_ListZones(t *testing.T) {
	client := setupTest(t, "/zones",
		writeFixtures(http.MethodGet, "zones_all_GET.xml", http.StatusOK))

	zones, err := client.ListZones(context.Background())
	require.NoError(t, err)

	expected := []Zone{
		{
			Admin:      "123/NIC-REG",
			Enable:     "true",
			HasChanges: "false",
			HasPrimary: "true",
			ID:         "227645",
			IDNName:    "тест.рф",
			Name:       "xn—e1aybc.xn--p1ai",
			Payer:      "123/NIC-REG",
			Service:    "myservice",
		},
		{
			Admin:      "123/NIC-REG",
			Enable:     "true",
			HasChanges: "false",
			HasPrimary: "true",
			ID:         "227642",
			IDNName:    "example.ru",
			Name:       "example.ru",
			Payer:      "123/NIC-REG",
			Service:    "myservice",
		},
		{
			Admin:      "123/NIC-REG",
			Enable:     "true",
			HasChanges: "false",
			HasPrimary: "true",
			ID:         "227643",
			IDNName:    "test.su",
			Name:       "test.su",
			Payer:      "123/NIC-REG",
			Service:    "myservice",
		},
	}

	assert.Equal(t, expected, zones)
}

func TestClient_ListZones_error(t *testing.T) {
	client := setupTest(t, "/zones",
		writeFixtures(http.MethodGet, "errors.xml", http.StatusOK))

	_, err := client.ListZones(context.Background())
	require.ErrorIs(t, err, Error{
		Text: "Access token expired or not found",
		Code: "4097",
	})
}

func TestClient_GetZonesByService(t *testing.T) {
	client := setupTest(t, "/services/test/zones",
		writeFixtures(http.MethodGet, "zones_GET.xml", http.StatusOK))

	zones, err := client.GetZonesByService(context.Background(), "test")
	require.NoError(t, err)

	expected := []Zone{
		{
			Admin:      "123/NIC-REG",
			Enable:     "true",
			HasChanges: "false",
			HasPrimary: "true",
			ID:         "227645",
			IDNName:    "тест.рф",
			Name:       "xn—e1aybc.xn--p1ai",
			Payer:      "123/NIC-REG",
			Service:    "myservice",
		},
		{
			Admin:      "123/NIC-REG",
			Enable:     "true",
			HasChanges: "false",
			HasPrimary: "true",
			ID:         "227642",
			IDNName:    "example.ru",
			Name:       "example.ru",
			Payer:      "123/NIC-REG",
			Service:    "myservice",
		},
		{
			Admin:      "123/NIC-REG",
			Enable:     "true",
			HasChanges: "false",
			HasPrimary: "true",
			ID:         "227643",
			IDNName:    "test.su",
			Name:       "test.su",
			Payer:      "123/NIC-REG",
			Service:    "myservice",
		},
	}

	assert.Equal(t, expected, zones)
}

func TestClient_GetZonesByService_error(t *testing.T) {
	client := setupTest(t, "/services/test/zones",
		writeFixtures(http.MethodGet, "errors.xml", http.StatusOK))

	_, err := client.GetZonesByService(context.Background(), "test")
	require.ErrorIs(t, err, Error{
		Text: "Access token expired or not found",
		Code: "4097",
	})
}

func TestClient_GetRecords(t *testing.T) {
	client := setupTest(t, "/services/test/zones/example.com./records",
		writeFixtures(http.MethodGet, "records_GET.xml", http.StatusOK))

	records, err := client.GetRecords(context.Background(), "test", "example.com.")
	require.NoError(t, err)

	expected := []RR{
		{
			ID:      "210074",
			Name:    "@",
			IDNName: "@",
			TTL:     "",
			Type:    "SOA",
			SOA: &SOA{
				MName: &MName{
					Name:    "ns3-l2.nic.ru.",
					IDNName: "ns3-l2.nic.ru.",
				},
				RName: &RName{
					Name:    "dns.nic.ru.",
					IDNName: "dns.nic.ru.",
				},
				Serial:  "2011112002",
				Refresh: "1440",
				Retry:   "3600",
				Expire:  "2592000",
				Minimum: "600",
			},
		},
		{
			ID:      "210075",
			Name:    "@",
			IDNName: "@",
			Type:    "NS",
			NS: &NS{
				Name:    "ns3-l2.nic.ru.",
				IDNName: "ns3- l2.nic.ru.",
			},
		},
		{
			ID:      "210076",
			Name:    "@",
			IDNName: "@",
			Type:    "NS",
			NS: &NS{
				Name:    "ns4-l2.nic.ru.",
				IDNName: "ns4-l2.nic.ru.",
			},
		},
		{
			ID:      "210077",
			Name:    "@",
			IDNName: "@",
			Type:    "NS",
			NS: &NS{
				Name:    "ns8-l2.nic.ru.",
				IDNName: "ns8- l2.nic.ru.",
			},
		},
	}

	assert.Equal(t, expected, records)
}

func TestClient_GetRecords_error(t *testing.T) {
	client := setupTest(t, "/services/test/zones/example.com./records",
		writeFixtures(http.MethodGet, "errors.xml", http.StatusOK))

	_, err := client.GetRecords(context.Background(), "test", "example.com.")
	require.ErrorIs(t, err, Error{
		Text: "Access token expired or not found",
		Code: "4097",
	})
}

func TestClient_AddRecord(t *testing.T) {
	client := setupTest(t, "/services/test/zones/example.com./records",
		writeFixtures(http.MethodPut, "records_PUT.xml", http.StatusOK))

	rrs := []RR{
		{
			Name: "@",
			Type: "NS",
			NS:   &NS{Name: "ns4-l2.nic.ru."},
		},
		{
			Name: "@",
			Type: "NS",
			NS:   &NS{Name: "ns8-l2.nic.ru."},
		},
	}

	response, err := client.AddRecords(context.Background(), "test", "example.com.", rrs)
	require.NoError(t, err)

	expected := []Zone{
		{
			Admin:      "123/NIC-REG",
			HasChanges: "true",
			ID:         "228095",
			IDNName:    "test.ru",
			Name:       "test.ru",
			Service:    "testservice",
			RR: []RR{
				{
					ID:      "210076",
					Name:    "@",
					IDNName: "@",
					Type:    "NS",
					NS: &NS{
						Name:    "ns4-l2.nic.ru.",
						IDNName: "ns4-l2.nic.ru.",
					},
				},
				{
					ID:      "210077",
					Name:    "@",
					IDNName: "@",
					Type:    "NS",
					NS: &NS{
						Name:    "ns8-l2.nic.ru.",
						IDNName: "ns8-l2.nic.ru.",
					},
				},
			},
		},
	}

	assert.Equal(t, expected, response)
}

func TestClient_AddRecord_error(t *testing.T) {
	client := setupTest(t, "/services/test/zones/example.com./records",
		writeFixtures(http.MethodPut, "errors.xml", http.StatusOK))

	rrs := []RR{
		{
			Name: "@",
			Type: "NS",
			NS:   &NS{Name: "ns4-l2.nic.ru."},
		},
		{
			Name: "@",
			Type: "NS",
			NS:   &NS{Name: "ns8-l2.nic.ru."},
		},
	}

	_, err := client.AddRecords(context.Background(), "test", "example.com.", rrs)
	require.ErrorIs(t, err, Error{
		Text: "Access token expired or not found",
		Code: "4097",
	})
}

func TestClient_DeleteRecord(t *testing.T) {
	client := setupTest(t, "/services/test/zones/example.com./records/123",
		writeFixtures(http.MethodDelete, "record_DELETE.xml", http.StatusUnauthorized))

	err := client.DeleteRecord(context.Background(), "test", "example.com.", "123")
	require.NoError(t, err)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client := setupTest(t, "/services/test/zones/example.com./records/123",
		writeFixtures(http.MethodDelete, "errors.xml", http.StatusUnauthorized))

	err := client.DeleteRecord(context.Background(), "test", "example.com.", "123")
	require.ErrorIs(t, err, Error{
		Text: "Access token expired or not found",
		Code: "4097",
	})
}

func TestClient_CommitZone(t *testing.T) {
	client := setupTest(t, "/services/test/zones/example.com./commit", writeFixtures(http.MethodPost, "commit_POST.xml", http.StatusOK))

	err := client.CommitZone(context.Background(), "test", "example.com.")
	require.NoError(t, err)
}

func TestClient_CommitZone_error(t *testing.T) {
	client := setupTest(t, "/services/test/zones/example.com./commit", writeFixtures(http.MethodPost, "errors.xml", http.StatusOK))

	err := client.CommitZone(context.Background(), "test", "example.com.")
	require.ErrorIs(t, err, Error{
		Text: "Access token expired or not found",
		Code: "4097",
	})
}
