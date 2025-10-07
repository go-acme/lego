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
			client, err := NewClient("secret")
			if err != nil {
				return nil, err
			}

			client.BaseURL, _ = url.Parse(server.URL)
			client.HTTPClient = server.Client()

			return client, nil
		},
		servermock.CheckHeader().
			WithAccept("application/json").
			With("X-Api-Key", "secret"),
	)
}

func TestClient_ListDomains(t *testing.T) {
	client := mockBuilder().
		Route("GET /domains",
			servermock.ResponseFromFixture("list_domains.json"),
			servermock.CheckQueryParameter().Strict().
				With("domain-name", "example.com")).
		Build(t)

	domains, err := client.ListDomains(t.Context(), "example.com")
	require.NoError(t, err)

	expected := map[string]Domain{
		"2976": {DomainName: "example.com", RegistrationDate: "12/09/2021", ExpirationDate: "12/09/2024", Status: "active"},
		"2977": {DomainName: "example.org", RegistrationDate: "01/10/2021", ExpirationDate: "01/10/2024", Status: "active"},
		"2978": {DomainName: "example.net", RegistrationDate: "21/08/2025", ExpirationDate: "-", Status: "active"},
	}

	assert.Equal(t, expected, domains)
}

func TestClient_ListDomains_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /domains",
			servermock.Noop().WithStatusCode(http.StatusBadRequest)).
		Build(t)

	_, err := client.ListDomains(t.Context(), "example.com")
	require.EqualError(t, err, "unexpected status code: [status code: 400] body: ")
}

func TestClient_ListDomains_api_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /domains",
			servermock.ResponseFromFixture("error.json")).
		Build(t)

	_, err := client.ListDomains(t.Context(), "example.com")
	require.EqualError(t, err, "unexpected status: error: missing required fields (type, name, ttl)")
}

func TestClient_ListDNSRecords(t *testing.T) {
	client := mockBuilder().
		Route("POST /domains/dns-records/list",
			servermock.ResponseFromFixture("list_dns_records.json"),
			servermock.CheckHeader().
				WithContentType("application/x-www-form-urlencoded"),
			servermock.CheckForm().Strict().
				With("order-id", "abc").
				With("types[]", "TXT")).
		Build(t)

	records, err := client.ListDNSRecords(t.Context(), "abc", "TXT")
	require.NoError(t, err)

	expected := []Record{
		{ID: 15, Type: "A", Name: "example.com.", TTL: 14400, Value: "203.0.113.10"},
		{ID: 22, Type: "MX", Name: "example.com.", TTL: 14400, Value: "10 mail.example.com."},
		{ID: 31, Type: "TXT", Name: "_dmarc.example.com.", TTL: 300, Value: "v=DMARC1; p=none; rua=mailto:dmarc@example.com"},
	}

	assert.Equal(t, expected, records)
}

func TestClient_ListDNSRecords_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /domains/dns-records/list",
			servermock.Noop().WithStatusCode(http.StatusBadRequest)).
		Build(t)

	_, err := client.ListDNSRecords(t.Context(), "abc", "TXT")
	require.EqualError(t, err, "unexpected status code: [status code: 400] body: ")
}

func TestClient_ListDNSRecords_api_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /domains/dns-records/list",
			servermock.ResponseFromFixture("error.json")).
		Build(t)

	_, err := client.ListDNSRecords(t.Context(), "abc", "TXT")
	require.EqualError(t, err, "unexpected status: error: missing required fields (type, name, ttl)")
}

func TestClient_AddDNSRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /domains/dns-records/add",
			servermock.ResponseFromFixture("add_dns_record.json"),
			servermock.CheckHeader().
				WithContentType("application/x-www-form-urlencoded"),
			servermock.CheckForm().Strict().
				With("order-id", "abc").
				With("name", "example.com.").
				With("ttl", "120").
				With("type", "TXT").
				With("value", "txtTXTtxt")).
		Build(t)

	record := Record{
		Type:  "TXT",
		Name:  "example.com.",
		TTL:   120,
		Value: "txtTXTtxt",
	}

	result, err := client.AddDNSRecord(t.Context(), "abc", record)
	require.NoError(t, err)

	expected := &Record{
		Type:  "A",
		Name:  "example.com.",
		TTL:   14400,
		Value: "203.0.113.10",
	}

	assert.Equal(t, expected, result)
}

func TestClient_AddDNSRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /domains/dns-records/add",
			servermock.Noop().WithStatusCode(http.StatusBadRequest)).
		Build(t)

	record := Record{
		Type:  "TXT",
		Name:  "example.com.",
		TTL:   120,
		Value: "txtTXTtxt",
	}

	_, err := client.AddDNSRecord(t.Context(), "abc", record)
	require.EqualError(t, err, "unexpected status code: [status code: 400] body: ")
}

func TestClient_AddDNSRecord_api_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /domains/dns-records/add",
			servermock.ResponseFromFixture("error.json")).
		Build(t)

	record := Record{
		Type:  "TXT",
		Name:  "example.com.",
		TTL:   120,
		Value: "txtTXTtxt",
	}

	_, err := client.AddDNSRecord(t.Context(), "abc", record)
	require.EqualError(t, err, "unexpected status: error: missing required fields (type, name, ttl)")
}

func TestClient_DeleteDNSRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /domains/dns-records/delete",
			servermock.ResponseFromFixture("delete_dns_record.json"),
			servermock.CheckHeader().
				WithContentType("application/x-www-form-urlencoded"),
			servermock.CheckForm().Strict().
				With("order-id", "abc").
				With("line", "123")).
		Build(t)

	result, err := client.DeleteDNSRecord(t.Context(), "abc", 123)
	require.NoError(t, err)

	expected := &DeletedRecordInfo{
		Count: 1,
		Lines: []int{15},
	}

	assert.Equal(t, expected, result)
}

func TestClient_DeleteDNSRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /domains/dns-records/delete",
			servermock.Noop().WithStatusCode(http.StatusBadRequest)).
		Build(t)

	_, err := client.DeleteDNSRecord(t.Context(), "abc", 123)
	require.EqualError(t, err, "unexpected status code: [status code: 400] body: ")
}

func TestClient_DeleteDNSRecord_api_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /domains/dns-records/delete",
			servermock.ResponseFromFixture("error.json")).
		Build(t)

	_, err := client.DeleteDNSRecord(t.Context(), "abc", 123)
	require.EqualError(t, err, "unexpected status: error: missing required fields (type, name, ttl)")
}
