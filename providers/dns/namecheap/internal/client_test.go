package internal

import (
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupClient(server *httptest.Server) (*Client, error) {
	client := NewClient("user", "secret", "127.0.0.1")
	client.HTTPClient = server.Client()
	client.BaseURL = server.URL

	return client, nil
}

func TestClient_GetHosts(t *testing.T) {
	client := servermock.NewBuilder[*Client](setupClient).
		Route("GET /",
			servermock.ResponseFromFixture("getHosts.xml"),
			servermock.CheckQueryParameter().Strict().
				With("ApiKey", "secret").
				With("ApiUser", "user").
				With("ClientIp", "127.0.0.1").
				With("Command", "namecheap.domains.dns.getHosts").
				With("SLD", "foo").
				With("TLD", "example.com").
				With("UserName", "user"),
		).
		Build(t)

	hosts, err := client.GetHosts(t.Context(), "foo", "example.com")
	require.NoError(t, err)

	expected := []Record{
		{Type: "A", Name: "@", Address: "1.2.3.4", MXPref: "10", TTL: "1800"},
		{Type: "A", Name: "www", Address: "122.23.3.7", MXPref: "10", TTL: "1800"},
	}

	assert.Equal(t, expected, hosts)
}

func TestClient_GetHosts_error(t *testing.T) {
	client := servermock.NewBuilder[*Client](setupClient).
		Route("GET /",
			servermock.ResponseFromFixture("getHosts_errorBadAPIKey1.xml")).
		Build(t)

	_, err := client.GetHosts(t.Context(), "foo", "example.com")
	require.ErrorAs(t, err, &apiError{})
}

func TestClient_SetHosts(t *testing.T) {
	client := servermock.NewBuilder[*Client](setupClient, servermock.CheckHeader().WithContentTypeFromURLEncoded()).
		Route("POST /",
			servermock.ResponseFromFixture("setHosts.xml"),
			servermock.CheckForm().Strict().
				With("ApiKey", "secret").
				With("ApiUser", "user").
				With("ClientIp", "127.0.0.1").
				With("Command", "namecheap.domains.dns.setHosts").
				With("SLD", "foo").
				With("TLD", "example.com").
				With("UserName", "user").
				// entry 1
				With("HostName1", "_acme-challenge.test.example.com").
				With("RecordType1", "TXT").
				With("Address1", "txtTXTtxt").
				With("MXPref1", "10").
				With("TTL1", "120").
				// entry 2
				With("HostName2", "_acme-challenge.test.example.org").
				With("RecordType2", "TXT").
				With("Address2", "txtTXTtxt").
				With("MXPref2", "10").
				With("TTL2", "120"),
		).
		Build(t)

	records := []Record{
		{Name: "_acme-challenge.test.example.com", Type: "TXT", Address: "txtTXTtxt", MXPref: "10", TTL: "120"},
		{Name: "_acme-challenge.test.example.org", Type: "TXT", Address: "txtTXTtxt", MXPref: "10", TTL: "120"},
	}

	err := client.SetHosts(t.Context(), "foo", "example.com", records)
	require.NoError(t, err)
}

func TestClient_SetHosts_error(t *testing.T) {
	client := servermock.NewBuilder[*Client](setupClient).
		Route("POST /",
			servermock.ResponseFromFixture("setHosts_errorBadAPIKey1.xml")).
		Build(t)

	records := []Record{
		{Name: "_acme-challenge.test.example.com", Type: "TXT", Address: "txtTXTtxt", MXPref: "10", TTL: "120"},
		{Name: "_acme-challenge.test.example.org", Type: "TXT", Address: "txtTXTtxt", MXPref: "10", TTL: "120"},
	}

	err := client.SetHosts(t.Context(), "foo", "example.com", records)
	require.ErrorAs(t, err, &apiError{})
}
