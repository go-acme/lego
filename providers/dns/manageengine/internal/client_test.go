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
			client := NewClient(server.Client())

			client.baseURL, _ = url.Parse(server.URL)

			return client, nil
		},
		servermock.CheckHeader().
			WithAccept("application/json"))
}

func TestClient_GetAllZones(t *testing.T) {
	client := mockBuilder().
		Route("GET /dns/domain", servermock.ResponseFromFixture("zone_domains_all.json")).
		Build(t)

	groups, err := client.GetAllZones(t.Context())
	require.NoError(t, err)

	expected := []Zone{
		{
			ZoneID:        1,
			ZoneName:      "test.com.",
			ZoneTTL:       500,
			ZoneTargeting: true,
			Refresh:       43200,
			Retry:         3600,
			Expiry:        1209600,
			Minimum:       180,
			Org:           2,
			NsID:          1,
			Serial:        2022042206,
			Nss:           []string{"ns11.zns-53.com.", "ns21.zns-53.net.", "ns31.zns-53.com.", "ns41.zns-53.net."},
		},
		{
			ZoneID:   2,
			ZoneName: "yourdomain.com.",
			ZoneTTL:  1000,
			Refresh:  43200,
			Retry:    3600,
			Expiry:   1209600,
			Minimum:  180,
			Org:      2,
			Vanity:   true,
			NsID:     1,
			Serial:   2022040608,
			Nss:      []string{"ns11.yourdomain.com.", "ns21.yourdomain.net.", "ns31.yourdomain.com.", "ns41.yourdomain.net."},
		},
		{
			ZoneID:   20,
			ZoneName: "hello45.com.",
			ZoneTTL:  3000,
			Refresh:  43200,
			Retry:    3600,
			Expiry:   1209600,
			Minimum:  180,
			Org:      2,
			NsID:     1,
			Serial:   2022040711,
			Nss:      []string{"ns11.zns-53.com.", "ns21.zns-53.net.", "ns31.zns-53.com.", "ns41.zns-53.net."},
		},
		{
			ZoneID:        22,
			ZoneName:      "zohoaccl.com.",
			ZoneTTL:       300,
			ZoneTargeting: true,
			Refresh:       43200,
			Retry:         3600,
			Expiry:        1209600,
			Minimum:       180,
			Org:           2,
			NsID:          1,
			Serial:        2022042206,
			Nss:           []string{"ns11.zns-53.com.", "ns21.zns-53.net.", "ns31.zns-53.com.", "ns41.zns-53.net."},
		},
		{
			ZoneID:        23,
			ZoneName:      "zohocal.com.",
			ZoneTTL:       300,
			ZoneTargeting: true,
			Refresh:       43200,
			Retry:         3600,
			Expiry:        1209600,
			Minimum:       180,
			Org:           2,
			NsID:          1,
			Serial:        2022041310,
			Nss:           []string{"ns11.zns-53.com.", "ns21.zns-53.net.", "ns31.zns-53.com.", "ns41.zns-53.net."},
		},
	}

	assert.Equal(t, expected, groups)
}

func TestClient_GetAllZones_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /dns/domain",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized)).
		Build(t)

	_, err := client.GetAllZones(t.Context())
	require.Error(t, err)

	require.EqualError(t, err, "[status code: 401] Authentication credentials were not provided.")
}

func TestClient_GetAllZoneRecords(t *testing.T) {
	client := mockBuilder().
		Route("GET /dns/domain/4/records/SPF_TXT", servermock.ResponseFromFixture("zone_records_all.json")).
		Build(t)

	groups, err := client.GetAllZoneRecords(t.Context(), 4)
	require.NoError(t, err)

	expected := []ZoneRecord{
		{
			ZoneID:           4,
			SpfTxtDomainID:   6,
			DomainName:       "spftest.example.com.",
			DomainTTL:        300,
			DomainLocationID: 1,
			RecordType:       "SPF",
			Records: []Record{{
				ID:       1,
				Values:   []string{"necwcltpwxbz-noelget3jush-vop2xxvapot3eyq_0"},
				DomainID: 6,
			}},
		},
		{
			ZoneID:           4,
			SpfTxtDomainID:   13,
			DomainName:       "txt.example.com.",
			DomainTTL:        300,
			DomainLocationID: 1,
			RecordType:       "TXT",
			Records: []Record{{
				ID:       1,
				Values:   []string{"v=spf1include:transmail.netinclude:example.com~all", "c-68e3oc4trm8w7piplscg7vgojmtkjrnrabr4king8"},
				DomainID: 13,
			}},
		},
	}

	assert.Equal(t, expected, groups)
}

func TestClient_GetAllZoneRecords_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /dns/domain/4/records/SPF_TXT",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized)).
		Build(t)

	_, err := client.GetAllZoneRecords(t.Context(), 4)
	require.Error(t, err)

	require.EqualError(t, err, "[status code: 401] Authentication credentials were not provided.")
}

func TestClient_DeleteZoneRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /dns/domain/4/records/SPF_TXT/6", servermock.ResponseFromFixture("zone_record_delete.json")).
		Build(t)

	err := client.DeleteZoneRecord(t.Context(), 4, 6)
	require.NoError(t, err)
}

func TestClient_DeleteZoneRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /dns/domain/4/records/SPF_TXT/6",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized)).
		Build(t)

	err := client.DeleteZoneRecord(t.Context(), 4, 6)
	require.Error(t, err)

	require.EqualError(t, err, "[status code: 401] Authentication credentials were not provided.")
}

func TestClient_CreateZoneRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /dns/domain/4/records/SPF_TXT/",
			servermock.ResponseFromFixture("zone_record_create.json"),
			servermock.CheckHeader().
				WithContentTypeFromURLEncoded(),
			servermock.CheckForm().Strict().
				With("config", `[{"zone_id":1,"spf_txt_domain_id":2,"domain_name":"example.com","domain_ttl":120,"domain_location_id":3,"record_type":"TXT","records":[{"record_id":123,"value":["value1"],"domain_id":1}]}]
`)).
		Build(t)

	record := ZoneRecord{
		ZoneID:           1,
		SpfTxtDomainID:   2,
		DomainName:       "example.com",
		DomainTTL:        120,
		DomainLocationID: 3,
		RecordType:       "TXT",
		Records: []Record{
			{
				ID:       123,
				Values:   []string{"value1"},
				Disabled: false,
				DomainID: 1,
			},
		},
	}

	err := client.CreateZoneRecord(t.Context(), 4, record)
	require.NoError(t, err)
}

func TestClient_CreateZoneRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /dns/domain/4/records/SPF_TXT/",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized),
			servermock.CheckHeader().
				WithContentTypeFromURLEncoded()).
		Build(t)

	record := ZoneRecord{}

	err := client.CreateZoneRecord(t.Context(), 4, record)
	require.Error(t, err)

	require.EqualError(t, err, "[status code: 401] Authentication credentials were not provided.")
}

func TestClient_CreateZoneRecord_error_bad_request(t *testing.T) {
	client := mockBuilder().
		Route("POST /dns/domain/4/records/SPF_TXT/",
			servermock.ResponseFromFixture("error_bad_request.json").
				WithStatusCode(http.StatusBadRequest),
			servermock.CheckHeader().
				WithContentTypeFromURLEncoded()).
		Build(t)

	record := ZoneRecord{}

	err := client.CreateZoneRecord(t.Context(), 4, record)
	require.Error(t, err)

	require.EqualError(t, err, "[status code: 400] Invalid record format, Record should be in list.")
}

func TestClient_UpdateZoneRecord(t *testing.T) {
	client := mockBuilder().
		Route("PUT /dns/domain/4/records/SPF_TXT/6/",
			servermock.ResponseFromFixture("zone_record_update.json"),
			servermock.CheckHeader().
				WithContentTypeFromURLEncoded(),
			servermock.CheckForm().Strict().
				With("config", `[{"zone_id":4,"spf_txt_domain_id":6,"records":null}]
`)).
		Build(t)

	record := ZoneRecord{
		SpfTxtDomainID: 6,
		ZoneID:         4,
	}

	err := client.UpdateZoneRecord(t.Context(), record)
	require.NoError(t, err)
}

func TestClient_UpdateZoneRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("PUT /dns/domain/4/records/SPF_TXT/6/",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized),
			servermock.CheckHeader().
				WithContentTypeFromURLEncoded()).
		Build(t)

	record := ZoneRecord{
		SpfTxtDomainID: 6,
		ZoneID:         4,
	}

	err := client.UpdateZoneRecord(t.Context(), record)
	require.Error(t, err)

	require.EqualError(t, err, "[status code: 401] Authentication credentials were not provided.")
}
