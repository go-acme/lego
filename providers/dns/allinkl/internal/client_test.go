package internal

import (
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/clientmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupClient(server *httptest.Server) (*Client, error) {
	client := NewClient("user")
	client.baseURL = server.URL
	client.HTTPClient = server.Client()

	return client, nil
}

func TestClient_GetDNSSettings(t *testing.T) {
	client := clientmock.NewBuilder[*Client](setupClient).
		Route("POST /", clientmock.ResponseFromFixture("get_dns_settings.xml"),
			clientmock.CheckRequestBodyFromFile("get_dns_settings-request.xml")).
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

func TestClient_AddDNSSettings(t *testing.T) {
	client := clientmock.NewBuilder[*Client](setupClient).
		Route("POST /", clientmock.ResponseFromFixture("add_dns_settings.xml"),
			clientmock.CheckRequestBodyFromFile("add_dns_settings-request.xml")).
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
	client := clientmock.NewBuilder[*Client](setupClient).
		Route("POST /", clientmock.ResponseFromFixture("delete_dns_settings.xml"),
			clientmock.CheckRequestBodyFromFile("delete_dns_settings-request.xml")).
		Build(t)

	r, err := client.DeleteDNSSettings(mockContext(t), "57347450")
	require.NoError(t, err)

	assert.Equal(t, "TRUE", r)
}
