package internal

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupClient(server *httptest.Server) (*Client, error) {
	client := NewClient("secret")
	client.baseURL, _ = url.Parse(server.URL)
	client.zoneEndpoint = server.URL
	client.HTTPClient = server.Client()

	return client, nil
}

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](setupClient,
		servermock.CheckHeader().WithJSONHeaders().
			WithBasicAuth("api", "secret"))
}

func TestGetRecords(t *testing.T) {
	client := mockBuilder().
		Route("GET /zones/47c0ecf6c91243308c649ad1d2d618dd/dnsrecords",
			servermock.ResponseFromFixture("records-GET.json")).
		Build(t)

	records, err := client.GetRecords(t.Context(), "47c0ecf6c91243308c649ad1d2d618dd", &GetRecordsParameters{DNSType: "TXT", Content: `"test"'`})
	require.NoError(t, err)

	excepted := &GetRecordsResponse{Data: []Record{
		{
			Type:    "A",
			Host:    "example.com",
			Content: "188.114.97.3",
			ID:      "812bee17a0b440b0bd5ee099a78b839c",
		},
		{
			Type:    "A",
			Host:    "example.com",
			Content: "188.114.96.3",
			ID:      "90e6029da45d4a36bf31056cf85d0cab",
		},
		{
			Type:    "AAAA",
			Host:    "example.com",
			Content: "2a06:98c1:3121::7",
			ID:      "0ac0320da0d24b5ca4f1648986a17340",
		},
		{
			Type:    "AAAA",
			Host:    "example.com",
			Content: "2a06:98c1:3120::7",
			ID:      "c91599694aea413498a0b3cd0a54a585",
		},
		{
			Type:    "A",
			Host:    "www",
			Content: "188.114.96.7",
			ID:      "c21f974992d549499f92e768bc468374",
		},
		{
			Type:    "A",
			Host:    "www",
			Content: "188.114.97.7",
			ID:      "90c3c1f05dca426893f10f122d18ad7a",
		},
		{
			Type:    "AAAA",
			Host:    "www",
			Content: "2a06:98c1:3121::",
			ID:      "379ab0ac0e434bc9aee5287e497f88a5",
		},
		{
			Type:    "AAAA",
			Host:    "www",
			Content: "2a06:98c1:3120::",
			ID:      "a1c4f9e50ba74791a4d70dc96999474c",
		},
	}, Count: 8}

	assert.Equal(t, excepted, records)
}

func TestGetRecords_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /zones/47c0ecf6c91243308c649ad1d2d618dd/dnsrecords",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized)).
		Build(t)

	_, err := client.GetRecords(t.Context(), "47c0ecf6c91243308c649ad1d2d618dd", &GetRecordsParameters{DNSType: "TXT", Content: `"test"'`})
	require.Error(t, err)
}

func TestGetRecord(t *testing.T) {
	client := mockBuilder().
		Route("GET /zones/47c0ecf6c91243308c649ad1d2d618dd/dnsrecords/812bee17a0b440b0bd5ee099a78b839c",
			servermock.ResponseFromFixture("record-GET.json")).
		Build(t)

	record, err := client.GetRecord(t.Context(), "47c0ecf6c91243308c649ad1d2d618dd", "812bee17a0b440b0bd5ee099a78b839c")
	require.NoError(t, err)

	excepted := &Record{
		Type:    "A",
		Host:    "example.com",
		Content: "188.114.97.3",
		ID:      "812bee17a0b440b0bd5ee099a78b839c",
	}

	assert.Equal(t, excepted, record)
}

func TestGetRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /zones/47c0ecf6c91243308c649ad1d2d618dd/dnsrecords",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized)).
		Build(t)

	_, err := client.GetRecord(t.Context(), "47c0ecf6c91243308c649ad1d2d618dd", "812bee17a0b440b0bd5ee099a78b839c")
	require.Error(t, err)
}

func TestCreateRecord(t *testing.T) {
	client := mockBuilder().
		Route("PUT /zones/47c0ecf6c91243308c649ad1d2d618dd/dnsrecords",
			servermock.ResponseFromFixture("record-PUT.json").
				WithStatusCode(http.StatusCreated)).
		Build(t)

	r := Record{
		Type:    "TXT",
		Host:    "test",
		Content: "test",
		TTL:     120,
	}

	record, err := client.CreateRecord(t.Context(), "47c0ecf6c91243308c649ad1d2d618dd", r)
	require.NoError(t, err)

	excepted := &Record{
		Type:    "A",
		Host:    "example.com",
		Content: "188.114.97.3",
		ID:      "812bee17a0b440b0bd5ee099a78b839c",
	}

	assert.Equal(t, excepted, record)
}

func TestCreateRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("PUT /zones/47c0ecf6c91243308c649ad1d2d618dd/dnsrecords",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized)).
		Build(t)

	r := Record{
		Type:    "TXT",
		Host:    "test",
		Content: "test",
		TTL:     120,
	}

	_, err := client.CreateRecord(t.Context(), "47c0ecf6c91243308c649ad1d2d618dd", r)
	require.Error(t, err)
}

func TestEditRecord(t *testing.T) {
	client := mockBuilder().
		Route("PATCH /zones/47c0ecf6c91243308c649ad1d2d618dd/dnsrecords/eebc813de2f94d67b09d91e10e2d65c2",
			servermock.ResponseFromFixture("record-PATCH.json")).
		Build(t)

	record, err := client.EditRecord(t.Context(), "47c0ecf6c91243308c649ad1d2d618dd", "eebc813de2f94d67b09d91e10e2d65c2", Record{
		Content: "foo",
	})
	require.NoError(t, err)

	excepted := &Record{
		Type:    "A",
		Host:    "example.com",
		Content: "188.114.97.3",
		ID:      "812bee17a0b440b0bd5ee099a78b839c",
	}

	assert.Equal(t, excepted, record)
}

func TestEditRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("PATCH /zones/47c0ecf6c91243308c649ad1d2d618dd/dnsrecords/eebc813de2f94d67b09d91e10e2d65c2",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized)).
		Build(t)

	_, err := client.EditRecord(t.Context(), "47c0ecf6c91243308c649ad1d2d618dd", "eebc813de2f94d67b09d91e10e2d65c2", Record{
		Content: "foo",
	})
	require.Error(t, err)
}

func TestDeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /zones/47c0ecf6c91243308c649ad1d2d618dd/dnsrecords/653464211b7447a1bee6b8fcb9fb86df",
			servermock.ResponseFromFixture("record-DELETE.json")).
		Build(t)

	err := client.DeleteRecord(t.Context(), "47c0ecf6c91243308c649ad1d2d618dd", "653464211b7447a1bee6b8fcb9fb86df")
	require.NoError(t, err)
}

func TestDeleteRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /zones/47c0ecf6c91243308c649ad1d2d618dd/dnsrecords/653464211b7447a1bee6b8fcb9fb86df",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized)).
		Build(t)

	err := client.DeleteRecord(t.Context(), "47c0ecf6c91243308c649ad1d2d618dd", "653464211b7447a1bee6b8fcb9fb86df")
	require.Error(t, err)
}

func TestGetZones(t *testing.T) {
	client := servermock.NewBuilder[*Client](setupClient,
		servermock.CheckHeader().
			WithBasicAuth("api", "secret"),
	).
		Route("GET /", servermock.ResponseFromFixture("service-cdn-zones.json")).
		Build(t)

	zones, err := client.GetZones(t.Context())
	require.NoError(t, err)

	excepted := []Zone{{
		ID:               "47c0ecf6c91243308c649ad1d2d618dd",
		Tags:             []string{},
		ContextID:        "47c0ecf6c91243308c649ad1d2d618dd",
		ContextType:      "CDN",
		HumanReadable:    "example.com",
		Serial:           "2301449956",
		CreationTime:     1679090659902,
		CreationTimeDate: time.Date(2023, time.March, 17, 22, 4, 19, 902000000, time.UTC),
		Status:           "active",
		IsMoved:          true,
		Paused:           false,
		ServiceType:      "CDN",
		Limbo:            false,
		TeamName:         "test",
		TeamID:           "640ef58496738d38fa7246a4",
		MyTeam:           true,
		RoleName:         "owner",
		IsBoard:          true,
		BoardRole:        []string{"owner"},
	}}

	assert.Equal(t, excepted, zones)
}

func TestGetZones_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /", servermock.ResponseFromFixture("error.json").
			WithStatusCode(http.StatusUnauthorized)).
		Build(t)

	_, err := client.GetZones(t.Context())
	require.Error(t, err)
}
