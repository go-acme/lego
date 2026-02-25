package internal

import (
	"net/http/httptest"
	"net/url"
	"testing"

	servermock2 "github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock2.Builder[*Client] {
	return servermock2.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client := NewClient("token", "secret")
			client.HTTPClient = server.Client()
			client.baseURL, _ = url.Parse(server.URL)

			return client, nil
		},
		servermock2.CheckHeader().WithJSONHeaders().
			WithBasicAuth("token", "secret"),
	)
}

func TestClient_CreateTXTRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /domains/1/dns",
			servermock2.ResponseFromFixture("create_record.json"),
			servermock2.CheckRequestJSONBodyFromFixture("create_record-request.json")).
		Build(t)

	err := client.CreateTXTRecord(t.Context(), &Domain{ID: 1}, "example.com", "txtTXTtxt")
	require.NoError(t, err)
}

func TestClient_DeleteTXTRecord(t *testing.T) {
	client := mockBuilder().
		Route("GET /domains/1/dns",
			servermock2.ResponseFromFixture("delete_record.json")).
		Route("DELETE /domains/1/dns/1", nil).
		Build(t)

	err := client.DeleteTXTRecord(t.Context(), &Domain{ID: 1}, "example.com", "txtTXTtxt")
	require.NoError(t, err)
}

func TestClient_getDNSRecordByHostData(t *testing.T) {
	client := mockBuilder().
		Route("GET /domains/1/dns",
			servermock2.ResponseFromFixture("getDnsRecords.json")).
		Build(t)

	record, err := client.getDNSRecordByHostData(t.Context(), Domain{ID: 1}, "example.com", "txtTXTtxt")
	require.NoError(t, err)

	expected := &DNSRecord{
		ID:   1,
		Type: "TXT",
		Host: "example.com",
		Data: "txtTXTtxt",
		TTL:  3600,
	}

	assert.Equal(t, expected, record)
}

func TestClient_GetDomainByName(t *testing.T) {
	client := mockBuilder().
		Route("GET /domains/",
			servermock2.ResponseFromFixture("getDomains.json")).
		Build(t)

	domain, err := client.GetDomainByName(t.Context(), "example.com")
	require.NoError(t, err)

	expected := &Domain{
		Name:           "example.com",
		ID:             1,
		ExpiryDate:     "2019-08-24",
		Nameservers:    []string{"ns1.hyp.net", "ns2.hyp.net", "ns3.hyp.net"},
		RegisteredDate: "2019-08-24",
		Registrant:     "Ola Nordmann",
		Renew:          true,
		Services: Service{
			DNS:       true,
			Email:     true,
			Registrar: true,
			Webhotel:  "none",
		},
		Status: "active",
	}

	assert.Equal(t, expected, domain)
}
