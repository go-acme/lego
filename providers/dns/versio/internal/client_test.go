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
			client := NewClient("user", "secret")
			client.HTTPClient = server.Client()
			client.BaseURL, _ = url.Parse(server.URL)

			return client, nil
		},
		servermock.CheckHeader().WithJSONHeaders().
			WithBasicAuth("user", "secret"))
}

func TestClient_GetDomain(t *testing.T) {
	client := mockBuilder().
		Route("GET /domains/example.com",
			servermock.ResponseFromFixture("get-domain.json"),
			servermock.CheckQueryParameter().Strict().
				With("show_dns_records", "true")).
		Build(t)

	records, err := client.GetDomain(t.Context(), "example.com")
	require.NoError(t, err)

	expected := &DomainInfoResponse{DomainInfo: DomainInfo{DNSRecords: []Record{
		{Type: "MX", Name: "example.com", Value: "fallback.axc.eu", Priority: 20, TTL: 3600},
		{Type: "TXT", Name: "example.com", Value: "\"v=spf1 a mx ip4:127.0.0.1 a:spf.spamexperts.axc.nl ~all\"", Priority: 0, TTL: 3600},
		{Type: "A", Name: "example.com", Value: "185.13.227.159", Priority: 0, TTL: 14400},
		{Type: "A", Name: "ftp.example.com", Value: "185.13.227.159", Priority: 0, TTL: 14400},
		{Type: "A", Name: "localhost.example.com", Value: "185.13.227.159", Priority: 0, TTL: 14400},
		{Type: "A", Name: "pop.example.com", Value: "185.13.227.159", Priority: 0, TTL: 14400},
		{Type: "A", Name: "smtp.example.com", Value: "185.13.227.159", Priority: 0, TTL: 14400},
		{Type: "A", Name: "www.example.com", Value: "185.13.227.159", Priority: 0, TTL: 14400},
		{Type: "A", Name: "dev.example.com", Value: "185.13.227.159", Priority: 0, TTL: 14400},
		{Type: "A", Name: "_domainkey.domain.com.example.com", Value: "185.13.227.159", Priority: 0, TTL: 14400},
		{Type: "MX", Name: "example.com", Value: "spamfilter2.axc.eu", Priority: 0, TTL: 3600},
		{Type: "A", Name: "redirect.example.com", Value: "localhost", Priority: 10, TTL: 14400},
	}}}

	assert.Equal(t, expected, records)
}

func TestClient_GetDomain_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /domains/example.com",
			servermock.ResponseFromFixture("get-domain-error.json").
				WithStatusCode(http.StatusUnauthorized)).
		Build(t)

	_, err := client.GetDomain(t.Context(), "example.com")
	require.ErrorAs(t, err, &ErrorMessage{})
}

func TestClient_UpdateDomain(t *testing.T) {
	client := mockBuilder().
		Route("POST /domains/example.com/update",
			servermock.ResponseFromFixture("update-domain.json"),
			servermock.CheckRequestJSONBodyFromFixture("update-domain-request.json")).
		Build(t)

	msg := &DomainInfo{DNSRecords: []Record{
		{Type: "MX", Name: "example.com", Value: "fallback.axc.eu", Priority: 20, TTL: 3600},
		{Type: "TXT", Name: "example.com", Value: "\"v=spf1 a mx ip4:127.0.0.1 a:spf.spamexperts.axc.nl ~all\"", Priority: 0, TTL: 3600},
		{Type: "A", Name: "example.com", Value: "185.13.227.159", Priority: 0, TTL: 14400},
		{Type: "A", Name: "ftp.example.com", Value: "185.13.227.159", Priority: 0, TTL: 14400},
		{Type: "A", Name: "localhost.example.com", Value: "185.13.227.159", Priority: 0, TTL: 14400},
		{Type: "A", Name: "pop.example.com", Value: "185.13.227.159", Priority: 0, TTL: 14400},
		{Type: "A", Name: "smtp.example.com", Value: "185.13.227.159", Priority: 0, TTL: 14400},
		{Type: "A", Name: "www.example.com", Value: "185.13.227.159", Priority: 0, TTL: 14400},
		{Type: "A", Name: "dev.example.com", Value: "185.13.227.159", Priority: 0, TTL: 14400},
		{Type: "A", Name: "_domainkey.domain.com.example.com", Value: "185.13.227.159", Priority: 0, TTL: 14400},
		{Type: "MX", Name: "example.com", Value: "spamfilter2.axc.eu", Priority: 0, TTL: 3600},
		{Type: "A", Name: "redirect.example.com", Value: "localhost", Priority: 10, TTL: 14400},
	}}

	records, err := client.UpdateDomain(t.Context(), "example.com", msg)
	require.NoError(t, err)

	expected := &DomainInfoResponse{DomainInfo: DomainInfo{DNSRecords: []Record{
		{Type: "MX", Name: "example.com", Value: "fallback.axc.eu", Priority: 20, TTL: 3600},
		{Type: "TXT", Name: "example.com", Value: "\"v=spf1 a mx ip4:127.0.0.1 a:spf.spamexperts.axc.nl ~all\"", Priority: 0, TTL: 3600},
		{Type: "A", Name: "example.com", Value: "185.13.227.159", Priority: 0, TTL: 14400},
		{Type: "A", Name: "ftp.example.com", Value: "185.13.227.159", Priority: 0, TTL: 14400},
		{Type: "A", Name: "localhost.example.com", Value: "185.13.227.159", Priority: 0, TTL: 14400},
		{Type: "A", Name: "pop.example.com", Value: "185.13.227.159", Priority: 0, TTL: 14400},
		{Type: "A", Name: "smtp.example.com", Value: "185.13.227.159", Priority: 0, TTL: 14400},
		{Type: "A", Name: "www.example.com", Value: "185.13.227.159", Priority: 0, TTL: 14400},
		{Type: "A", Name: "dev.example.com", Value: "185.13.227.159", Priority: 0, TTL: 14400},
		{Type: "A", Name: "_domainkey.domain.com.example.com", Value: "185.13.227.159", Priority: 0, TTL: 14400},
		{Type: "MX", Name: "example.com", Value: "spamfilter2.axc.eu", Priority: 0, TTL: 3600},
		{Type: "A", Name: "redirect.example.com", Value: "localhost", Priority: 10, TTL: 14400},
	}}}

	assert.Equal(t, expected, records)
}

func TestClient_UpdateDomain_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /domains/example.com/update",
			servermock.ResponseFromFixture("update-domain-error.json").
				WithStatusCode(http.StatusUnauthorized)).
		Build(t)

	msg := &DomainInfo{DNSRecords: []Record{
		{Type: "MX", Name: "example.com", Value: "fallback.axc.eu", Priority: 20, TTL: 3600},
		{Type: "TXT", Name: "example.com", Value: "\"v=spf1 a mx ip4:127.0.0.1 a:spf.spamexperts.axc.nl ~all\"", Priority: 0, TTL: 3600},
		{Type: "A", Name: "example.com", Value: "185.13.227.159", Priority: 0, TTL: 14400},
		{Type: "A", Name: "ftp.example.com", Value: "185.13.227.159", Priority: 0, TTL: 14400},
		{Type: "A", Name: "localhost.example.com", Value: "185.13.227.159", Priority: 0, TTL: 14400},
		{Type: "A", Name: "pop.example.com", Value: "185.13.227.159", Priority: 0, TTL: 14400},
		{Type: "A", Name: "smtp.example.com", Value: "185.13.227.159", Priority: 0, TTL: 14400},
		{Type: "A", Name: "www.example.com", Value: "185.13.227.159", Priority: 0, TTL: 14400},
		{Type: "A", Name: "dev.example.com", Value: "185.13.227.159", Priority: 0, TTL: 14400},
		{Type: "A", Name: "_domainkey.domain.com.example.com", Value: "185.13.227.159", Priority: 0, TTL: 14400},
		{Type: "MX", Name: "example.com", Value: "spamfilter2.axc.eu", Priority: 0, TTL: 3600},
		{Type: "A", Name: "redirect.example.com", Value: "localhost", Priority: 10, TTL: 14400},
	}}

	_, err := client.UpdateDomain(t.Context(), "example.com", msg)
	require.ErrorAs(t, err, &ErrorMessage{})
}
