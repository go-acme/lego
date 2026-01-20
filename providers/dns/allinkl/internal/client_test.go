package internal

import (
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupClient(server *httptest.Server) (*Client, error) {
	client := NewClient("user")
	client.BaseURL, _ = url.Parse(server.URL)
	client.HTTPClient = server.Client()

	client.maxElapsedTime = 1 * time.Second

	return client, nil
}

func TestClient_GetDNSSettings(t *testing.T) {
	client := servermock.NewBuilder[*Client](setupClient).
		Route("POST /KasApi.php", servermock.ResponseFromFixture("get_dns_settings.xml"),
			servermock.CheckRequestBodyFromFixture("get_dns_settings-request.xml").
				IgnoreWhitespace()).
		Build(t)

	records, err := client.GetDNSSettings(mockContext(t), "example.com", "")
	require.NoError(t, err)

	expected := []ReturnInfo{
		{
			ID:         "57297429",
			Zone:       "example.org",
			Name:       "",
			Type:       "A",
			Data:       "10.0.0.1",
			Changeable: "Y",
			Aux:        0,
		},
		{
			ID:         int64(0),
			Zone:       "example.org",
			Name:       "",
			Type:       "NS",
			Data:       "ns5.kasserver.com.",
			Changeable: "N",
			Aux:        0,
		},
		{
			ID:         int64(0),
			Zone:       "example.org",
			Name:       "",
			Type:       "NS",
			Data:       "ns6.kasserver.com.",
			Changeable: "N",
			Aux:        0,
		},
		{
			ID:         "57297479",
			Zone:       "example.org",
			Name:       "*",
			Type:       "A",
			Data:       "10.0.0.1",
			Changeable: "Y",
			Aux:        0,
		},
		{
			ID:         "57297481",
			Zone:       "example.org",
			Name:       "",
			Type:       "MX",
			Data:       "user.kasserver.com.",
			Changeable: "Y",
			Aux:        10,
		},
		{
			ID:         "57297483",
			Zone:       "example.org",
			Name:       "",
			Type:       "TXT",
			Data:       "v=spf1 mx a ?all",
			Changeable: "Y",
			Aux:        0,
		},
		{
			ID:         "57297485",
			Zone:       "example.org",
			Name:       "_dmarc",
			Type:       "TXT",
			Data:       "v=DMARC1; p=none;",
			Changeable: "Y",
			Aux:        0,
		},
	}

	assert.Equal(t, expected, records)
}

func TestClient_GetDNSSettings_error_flood_protection(t *testing.T) {
	client := servermock.NewBuilder[*Client](setupClient).
		Route("POST /KasApi.php",
			servermock.ResponseFromFixture("flood_protection.xml"),
		).
		Build(t)

	assert.Zero(t, client.floodTime)

	_, err := client.GetDNSSettings(mockContext(t), "example.com", "")
	require.EqualError(t, err, "KasApi: SOAP-ENV:Server: flood_protection: 0.0688529014587")

	assert.NotZero(t, client.floodTime)
}

func TestClient_AddDNSSettings(t *testing.T) {
	client := servermock.NewBuilder[*Client](setupClient).
		Route("POST /KasApi.php", servermock.ResponseFromFixture("add_dns_settings.xml"),
			servermock.CheckRequestBodyFromFixture("add_dns_settings-request.xml").
				IgnoreWhitespace()).
		Build(t)

	record := DNSRequest{
		ZoneHost:   "42cnc.de.",
		RecordType: "TXT",
		RecordName: "lego",
		RecordData: "abcdefgh",
	}

	recordID, err := client.AddDNSSettings(mockContext(t), record)
	require.NoError(t, err)

	assert.Equal(t, "57347444", recordID)
}

func TestClient_DeleteDNSSettings(t *testing.T) {
	client := servermock.NewBuilder[*Client](setupClient).
		Route("POST /KasApi.php", servermock.ResponseFromFixture("delete_dns_settings.xml"),
			servermock.CheckRequestBodyFromFixture("delete_dns_settings-request.xml").
				IgnoreWhitespace()).
		Build(t)

	r, err := client.DeleteDNSSettings(mockContext(t), "57347450")
	require.NoError(t, err)

	assert.Equal(t, "TRUE", r)
}
