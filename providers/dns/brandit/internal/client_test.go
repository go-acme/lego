package internal

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T, file string) *Client {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		open, err := os.Open(file)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		defer func() { _ = open.Close() }()

		rw.WriteHeader(http.StatusOK)
		_, err = io.Copy(rw, open)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	}))
	t.Cleanup(server.Close)

	client, err := NewClient("test_user", "apiKey")
	require.NoError(t, err)

	client.HTTPClient = server.Client()
	client.BaseURL = server.URL

	return client
}

func TestClient_StatusDomain(t *testing.T) {
	client := setupTest(t, "./fixtures/status-domain.json")

	domain, err := client.StatusDomain("example.com")
	require.NoError(t, err)

	expected := &StatusDomain{
		Response: StatusResponse{
			RenewalMode:                []string{"DEFAULT"},
			Status:                     []string{"clientTransferProhibited"},
			TransferLock:               []int{1},
			Registrar:                  []string{"brandit"},
			PaidUntilDate:              []string{"2021-12-15 05:00:00.0"},
			Nameserver:                 []string{"NS1.RRPPROXY.NET", "NS2.RRPPROXY.NET"},
			RegistrationExpirationDate: []string{"2021-12-15 05:00:00.0"},
			Domain:                     []string{"example.com"},
			RenewalDate:                []string{"2024-01-19 05:00:00.0"},
			UpdatedDate:                []string{"2022-12-16 08:01:27.0"},
			BillingContact:             []string{"example"},
			XDomainRoID:                []string{"example"},
			AdminContact:               []string{"example"},
			TechContact:                []string{"example"},
			DomainIDN:                  []string{"example.com"},
			CreatedDate:                []string{"2016-12-16 05:00:00.0"},
			RegistrarTransferDate:      []string{"2021-12-09 05:17:42.0"},
			Zone:                       []string{"com"},
			Auth:                       []string{"example"},
			UpdatedBy:                  []string{"example"},
			RoID:                       []string{"example"},
			OwnerContact:               []string{"example"},
			CreatedBy:                  []string{"example"},
			TransferMode:               []string{"auto"},
		},
		Code:   200,
		Status: "success",
		Error:  "",
	}

	assert.Equal(t, expected, domain)
}

func TestClient_ListRecords(t *testing.T) {
	client := setupTest(t, "./fixtures/list-records.json")

	resp, err := client.ListRecords("example", "example.com")
	require.NoError(t, err)

	expected := &ListRecords{
		Response: ListRecordsResponse{
			Limit:  []int{100},
			Column: []string{"rr"},
			Count:  []int{1},
			First:  []int{0},
			Total:  []int{1},
			RR:     []string{"example.com. 600 IN TXT txttxttxt"},
			Last:   []int{0},
		},
		Code:   200,
		Status: "success",
		Error:  "",
	}

	assert.Equal(t, expected, resp)
}

func TestClient_AddRecord(t *testing.T) {
	client := setupTest(t, "./fixtures/add-record.json")

	testRecord := Record{
		ID:      2565,
		Type:    "TXT",
		Name:    "example.com",
		Content: "txttxttxt",
		TTL:     600,
	}
	resp, err := client.AddRecord("example.com", "test", "2565", testRecord)
	require.NoError(t, err)

	expected := &AddRecord{
		Response: AddRecordResponse{
			ZoneType: []string{"com"},
			Signed:   []int{1},
		},
		Record: "example.com 600 IN TXT txttxttxt",
		Code:   200,
		Status: "success",
		Error:  "",
	}

	assert.Equal(t, expected, resp)
}

func TestClient_DeleteRecord(t *testing.T) {
	client := setupTest(t, "./fixtures/delete-record.json")

	resp, err := client.DeleteRecord("example.com", "test", "example.com 600 IN TXT txttxttxt", "2374")
	require.NoError(t, err)

	expected := &DeleteRecord{
		Code:   200,
		Status: "success",
		Error:  "",
	}

	assert.Equal(t, expected, resp)
}
