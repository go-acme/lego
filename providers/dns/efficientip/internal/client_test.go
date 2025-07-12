package internal

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			srvURL, _ := url.Parse(server.URL)

			client := NewClient(srvURL.Host, "user", "secret")
			client.HTTPClient = server.Client()
			client.baseURL, _ = url.Parse(server.URL)

			return client, nil
		},
		servermock.CheckHeader().WithJSONHeaders().
			WithBasicAuth("user", "secret"),
	)
}

func TestListRecords(t *testing.T) {
	client := mockBuilder().
		Route("GET /dns_rr_list", servermock.ResponseFromFixture("dns_rr_list.json")).
		Build(t)

	records, err := client.ListRecords(t.Context())
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
	client := mockBuilder().
		Route("GET /dns_rr_info", servermock.ResponseFromFixture("dns_rr_info.json"),
			servermock.CheckQueryParameter().Strict().
				With("rr_id", "239")).
		Build(t)

	record, err := client.GetRecord(t.Context(), "239")
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
	client := mockBuilder().
		Route("POST /dns_rr_add",
			servermock.ResponseFromFixture("dns_rr_add.json").WithStatusCode(http.StatusCreated),
			servermock.CheckRequestJSONBody(`{"dns_name":"dns.smart","dnsview_name":"external","rr_name":"test.example.com","rr_type":"TXT","value1":"test"}`)).
		Build(t)

	r := ResourceRecord{
		RRName:      "test.example.com",
		RRType:      "TXT",
		Value1:      "test",
		DNSName:     "dns.smart",
		DNSViewName: "external",
	}

	resp, err := client.AddRecord(t.Context(), r)
	require.NoError(t, err)

	expected := &BaseOutput{RetOID: "239"}

	assert.Equal(t, expected, resp)
}

func TestDeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /dns_rr_delete", servermock.ResponseFromFixture("dns_rr_delete.json"),
			servermock.CheckQueryParameter().Strict().
				With("rr_id", "251")).
		Build(t)

	resp, err := client.DeleteRecord(t.Context(), DeleteInputParameters{RRID: "251"})
	require.NoError(t, err)

	expected := &BaseOutput{RetOID: "251"}

	assert.Equal(t, expected, resp)
}

func TestDeleteRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /dns_rr_delete",
			servermock.ResponseFromFixture("dns_rr_delete-error.json").WithStatusCode(http.StatusBadRequest)).
		Build(t)

	_, err := client.DeleteRecord(t.Context(), DeleteInputParameters{RRID: "251"})
	require.ErrorAs(t, err, &APIError{})
}
