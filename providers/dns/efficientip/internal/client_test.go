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

func setupTest(t *testing.T, method, pattern string, status int, file string) *Client {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	mux.HandleFunc(pattern, func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != method {
			http.Error(rw, fmt.Sprintf("unsupported method %s", req.Method), http.StatusBadRequest)
			return
		}

		username, password, ok := req.BasicAuth()
		if !ok {
			http.Error(rw, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		if username != "user" {
			http.Error(rw, fmt.Sprintf("username: want %s got %s", username, "user"), http.StatusUnauthorized)
			return
		}

		if password != "secret" {
			http.Error(rw, fmt.Sprintf("password: want %s got %s", password, "secret"), http.StatusUnauthorized)
			return
		}

		open, err := os.Open(filepath.Join("fixtures", file))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		defer func() { _ = open.Close() }()

		rw.WriteHeader(status)
		_, err = io.Copy(rw, open)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	srvURL, _ := url.Parse(server.URL)

	client := NewClient(srvURL.Host, "user", "secret")
	client.HTTPClient = server.Client()
	client.baseURL, _ = url.Parse(server.URL)

	return client
}

func TestListRecords(t *testing.T) {
	client := setupTest(t, http.MethodGet, "/dns_rr_list", http.StatusOK, "dns_rr_list.json")

	ctx := context.Background()

	records, err := client.ListRecords(ctx)
	require.NoError(t, err)

	expected := []ResourceRecord{
		{
			ErrorCode:         "0",
			DelayedCreateTime: "0",
			DelayedDeleteTime: "0",
			DelayedTime:       "0",
			DNSCloud:          "0",
			DNSID:             "3",
			DNSName:           "dns.smart",
			DNSType:           "vdns",
			DNSViewID:         "0",
			DNSViewName:       "#",
			DNSZoneID:         "9",
			DNSZoneIsReverse:  "0",
			DNSZoneIsRpz:      "0",
			DNSZoneName:       "lego.example.com",
			DNSZoneNameUTF:    "lego.example.com",
			DNSZoneSiteName:   "#",
			DNSZoneSortZone:   "lego.example.com",
			DNSZoneType:       "master",
			RRAllValue:        "test1",
			RRAuthGsstsig:     "0",
			RRFullName:        "test.lego.example.com",
			RRFullNameUTF:     "test.lego.example.com",
			RRGlue:            "test",
			RRGlueID:          "21",
			RRID:              "239",
			RRNameID:          "26",
			RRType:            "TXT",
			RRTypeID:          "6",
			RRValueID:         "274",
			TTL:               "3600",
			Value1:            "test1",
			VDNSParentID:      "0",
			VDNSParentName:    "#",
		},
		{
			ErrorCode:         "0",
			DelayedCreateTime: "0",
			DelayedDeleteTime: "0",
			DelayedTime:       "0",
			DNSCloud:          "0",
			DNSID:             "3",
			DNSName:           "dns.smart",
			DNSType:           "vdns",
			DNSViewID:         "0",
			DNSViewName:       "#",
			DNSZoneID:         "9",
			DNSZoneIsReverse:  "0",
			DNSZoneIsRpz:      "0",
			DNSZoneName:       "lego.example.com",
			DNSZoneNameUTF:    "lego.example.com",
			DNSZoneSiteName:   "#",
			DNSZoneSortZone:   "lego.example.com",
			DNSZoneType:       "master",
			RRAllValue:        "test2",
			RRAuthGsstsig:     "0",
			RRFullName:        "test.lego.example.com",
			RRFullNameUTF:     "test.lego.example.com",
			RRGlue:            "test",
			RRGlueID:          "21",
			RRID:              "241",
			RRNameID:          "26",
			RRType:            "TXT",
			RRTypeID:          "6",
			RRValueID:         "275",
			TTL:               "3600",
			Value1:            "test2",
			VDNSParentID:      "0",
			VDNSParentName:    "#",
		},
		{
			ErrorCode:         "0",
			DelayedCreateTime: "0",
			DelayedDeleteTime: "0",
			DelayedTime:       "0",
			DNSCloud:          "0",
			DNSID:             "3",
			DNSName:           "dns.smart",
			DNSType:           "vdns",
			DNSViewID:         "0",
			DNSViewName:       "#",
			DNSZoneID:         "9",
			DNSZoneIsReverse:  "0",
			DNSZoneIsRpz:      "0",
			DNSZoneName:       "lego.example.com",
			DNSZoneNameUTF:    "lego.example.com",
			DNSZoneSiteName:   "#",
			DNSZoneSortZone:   "lego.example.com",
			DNSZoneType:       "master",
			RRAllValue:        "test1",
			RRAuthGsstsig:     "0",
			RRFullName:        "lego.example.com",
			RRFullNameUTF:     "lego.example.com",
			RRGlue:            ".",
			RRGlueID:          "3",
			RRID:              "245",
			RRNameID:          "21",
			RRType:            "TXT",
			RRTypeID:          "6",
			RRValueID:         "274",
			TTL:               "3600",
			Value1:            "test1",
			VDNSParentID:      "0",
			VDNSParentName:    "#",
		},
		{
			ErrorCode:         "0",
			DelayedCreateTime: "0",
			DelayedDeleteTime: "0",
			DelayedTime:       "0",
			DNSCloud:          "0",
			DNSID:             "3",
			DNSName:           "dns.smart",
			DNSType:           "vdns",
			DNSViewID:         "0",
			DNSViewName:       "#",
			DNSZoneID:         "9",
			DNSZoneIsReverse:  "0",
			DNSZoneIsRpz:      "0",
			DNSZoneName:       "lego.example.com",
			DNSZoneNameUTF:    "lego.example.com",
			DNSZoneSiteName:   "#",
			DNSZoneSortZone:   "lego.example.com",
			DNSZoneType:       "master",
			RRAllValue:        "test2",
			RRAuthGsstsig:     "0",
			RRFullName:        "lego.example.com",
			RRFullNameUTF:     "lego.example.com",
			RRGlue:            ".",
			RRGlueID:          "3",
			RRID:              "247",
			RRNameID:          "21",
			RRType:            "TXT",
			RRTypeID:          "6",
			RRValueID:         "275",
			TTL:               "3600",
			Value1:            "test2",
			VDNSParentID:      "0",
			VDNSParentName:    "#",
		},
		{
			ErrorCode:         "0",
			DelayedCreateTime: "0",
			DelayedDeleteTime: "0",
			DelayedTime:       "0",
			DNSCloud:          "0",
			DNSID:             "3",
			DNSName:           "dns.smart",
			DNSType:           "vdns",
			DNSViewID:         "0",
			DNSViewName:       "#",
			DNSZoneID:         "9",
			DNSZoneIsReverse:  "0",
			DNSZoneIsRpz:      "0",
			DNSZoneName:       "lego.example.com",
			DNSZoneNameUTF:    "lego.example.com",
			DNSZoneSiteName:   "#",
			DNSZoneSortZone:   "lego.example.com",
			DNSZoneType:       "master",
			RRAllValue:        "dns.smart, root@lego.example.com, 2023062719, 1200, 600, 1209600, 3600",
			RRAuthGsstsig:     "0",
			RRFullName:        "lego.example.com",
			RRFullNameUTF:     "lego.example.com",
			RRGlue:            ".",
			RRGlueID:          "3",
			RRID:              "201",
			RRNameID:          "21",
			RRType:            "SOA",
			RRTypeID:          "2",
			RRValueID:         "282",
			TTL:               "3600",
			Value1:            "dns.smart",
			Value2:            "root@lego.example.com",
			Value3:            "2023062719",
			Value4:            "1200",
			Value5:            "600",
			Value6:            "1209600",
			Value7:            "3600",
			VDNSParentID:      "0",
			VDNSParentName:    "#",
		},
		{
			ErrorCode:         "0",
			DelayedCreateTime: "0",
			DelayedDeleteTime: "0",
			DelayedTime:       "0",
			DNSCloud:          "0",
			DNSID:             "3",
			DNSName:           "dns.smart",
			DNSType:           "vdns",
			DNSViewID:         "0",
			DNSViewName:       "#",
			DNSZoneID:         "9",
			DNSZoneIsReverse:  "0",
			DNSZoneIsRpz:      "0",
			DNSZoneName:       "lego.example.com",
			DNSZoneNameUTF:    "lego.example.com",
			DNSZoneSiteName:   "#",
			DNSZoneSortZone:   "lego.example.com",
			DNSZoneType:       "master",
			RRAllValue:        "dns.smart",
			RRAuthGsstsig:     "0",
			RRFullName:        "lego.example.com",
			RRFullNameUTF:     "lego.example.com",
			RRGlue:            ".",
			RRGlueID:          "3",
			RRID:              "200",
			RRNameID:          "21",
			RRType:            "NS",
			RRTypeID:          "1",
			RRValueID:         "10",
			TTL:               "3600",
			Value1:            "dns.smart",
			VDNSParentID:      "0",
			VDNSParentName:    "#",
		},
		{
			ErrorCode:         "0",
			DelayedCreateTime: "0",
			DelayedDeleteTime: "0",
			DelayedTime:       "0",
			DNSCloud:          "0",
			DNSID:             "3",
			DNSName:           "dns.smart",
			DNSType:           "vdns",
			DNSViewID:         "0",
			DNSViewName:       "#",
			DNSZoneID:         "9",
			DNSZoneIsReverse:  "0",
			DNSZoneIsRpz:      "0",
			DNSZoneName:       "lego.example.com",
			DNSZoneNameUTF:    "lego.example.com",
			DNSZoneSiteName:   "#",
			DNSZoneSortZone:   "lego.example.com",
			DNSZoneType:       "master",
			RRAllValue:        "127.0.0.1",
			RRAuthGsstsig:     "0",
			RRFullName:        "loopback.lego.example.com",
			RRFullNameUTF:     "loopback.lego.example.com",
			RRGlue:            "loopback",
			RRGlueID:          "17",
			RRID:              "208",
			RRNameID:          "22",
			RRType:            "A",
			RRTypeID:          "3",
			RRValueID:         "237",
			RRValueIP4Addr:    "7f000001",
			RRValueIPAddr:     "7f000001",
			TTL:               "3600",
			Value1:            "127.0.0.1",
			VDNSParentID:      "0",
			VDNSParentName:    "#",
		},
	}

	assert.Equal(t, expected, records)
}

func TestGetRecord(t *testing.T) {
	client := setupTest(t, http.MethodGet, "/dns_rr_info", http.StatusOK, "dns_rr_info.json")

	ctx := context.Background()

	record, err := client.GetRecord(ctx, "239")
	require.NoError(t, err)

	expected := &ResourceRecord{
		ErrorCode:         "0",
		DelayedCreateTime: "0",
		DelayedDeleteTime: "0",
		DelayedTime:       "0",
		DNSCloud:          "0",
		DNSID:             "3",
		DNSName:           "dns.smart",
		DNSType:           "vdns",
		DNSViewID:         "0",
		DNSViewName:       "#",
		DNSZoneID:         "9",
		DNSZoneIsReverse:  "0",
		DNSZoneIsRpz:      "0",
		DNSZoneName:       "lego.example.com",
		DNSZoneNameUTF:    "lego.example.com",
		DNSZoneSiteName:   "#",
		DNSZoneSortZone:   "lego.example.com",
		DNSZoneType:       "master",
		RRAllValue:        "test1",
		RRAuthGsstsig:     "0",
		RRFullName:        "test.lego.example.com",
		RRFullNameUTF:     "test.lego.example.com",
		RRGlue:            "test",
		RRGlueID:          "21",
		RRID:              "239",
		RRNameID:          "26",
		RRType:            "TXT",
		RRTypeID:          "6",
		RRValueID:         "274",
		TTL:               "3600",
		Value1:            "test1",
		VDNSParentID:      "0",
		VDNSParentName:    "#",
	}

	assert.Equal(t, expected, record)
}

func TestAddRecord(t *testing.T) {
	client := setupTest(t, http.MethodPost, "/dns_rr_add", http.StatusCreated, "dns_rr_add.json")

	ctx := context.Background()

	r := ResourceRecord{
		RRName:      "test.example.com",
		RRType:      "TXT",
		Value1:      "test",
		DNSName:     "dns.smart",
		DNSViewName: "external",
	}

	resp, err := client.AddRecord(ctx, r)
	require.NoError(t, err)

	expected := &BaseOutput{RetOID: "239"}

	assert.Equal(t, expected, resp)
}

func TestDeleteRecord(t *testing.T) {
	client := setupTest(t, http.MethodDelete, "/dns_rr_delete", http.StatusOK, "dns_rr_delete.json")

	ctx := context.Background()

	resp, err := client.DeleteRecord(ctx, DeleteInputParameters{RRID: "251"})
	require.NoError(t, err)

	expected := &BaseOutput{RetOID: "251"}

	assert.Equal(t, expected, resp)
}

func TestDeleteRecord_error(t *testing.T) {
	client := setupTest(t, http.MethodDelete, "/dns_rr_delete", http.StatusBadRequest, "dns_rr_delete-error.json")

	ctx := context.Background()

	_, err := client.DeleteRecord(ctx, DeleteInputParameters{RRID: "251"})
	require.ErrorAs(t, err, &APIError{})
}
