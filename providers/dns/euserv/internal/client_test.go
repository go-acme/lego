package internal

import (
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client := NewClient()

			client.BaseURL = server.URL
			client.HTTPClient = server.Client()

			return client, nil
		},
		servermock.CheckHeader().
			WithJSONHeaders(),
	)
}

func TestClient_RemoveRecord(t *testing.T) {
	client := mockBuilder().
		Route("GET /",
			servermock.ResponseFromFixture("dns_remove.json"),
			servermock.CheckQueryParameter().Strict().
				With("subaction", "kc2_domain_dns_remove").
				With("dns_record_id", "a1b2c3").
				With("lang_id", "2").
				With("method", "json"),
		).
		Build(t)

	err := client.RemoveRecord(t.Context(), "a1b2c3")
	require.NoError(t, err)
}

func TestClient_SetRecord(t *testing.T) {
	client := mockBuilder().
		Route("GET /",
			servermock.ResponseFromFixture("dns_set.json"),
			servermock.CheckQueryParameter().Strict().
				With("subaction", "kc2_domain_dns_set").
				With("dom_id", "a1").
				With("dns_record_id", "b2").
				With("subdomain", "foo").
				With("type", "TXT").
				With("content", "txtTXTtxt").
				With("ttl", "120").
				With("lang_id", "2").
				With("method", "json"),
		).
		Build(t)

	request := SetRecordRequest{
		DomainID:  "a1",
		ID:        "b2",
		Subdomain: "foo",
		Type:      "TXT",
		Content:   "txtTXTtxt",
		TTL:       120,
	}

	err := client.SetRecord(t.Context(), request)
	require.NoError(t, err)
}

func TestClient_GetRecords(t *testing.T) {
	client := mockBuilder().
		Route("GET /",
			servermock.ResponseFromFixture("dns_get_records.json"),
			servermock.CheckQueryParameter().Strict().
				With("subaction", "kc2_domain_dns_get_records").
				With("dns_records_load_keyword", "bar").
				With("dns_records_load_only_for_dom_id", "a1").
				With("dns_records_load_page", "6").
				With("dns_records_load_subdomain", "foo").
				With("dns_records_load_type", "TXT").
				With("lang_id", "2").
				With("method", "json"),
		).
		Build(t)

	request := GetRecordsRequest{
		Page:         "6",
		OnlyForDomID: "a1",
		Keyword:      "bar",
		Subdomain:    "foo",
		Type:         "TXT",
	}

	domains, err := client.GetRecords(t.Context(), request)
	require.NoError(t, err)

	expected := []Domain{{
		ID:                      APIValue{Value: "1234"},
		Domain:                  APIValue{Value: "mydomain.com"},
		NameserversUsingDefault: APIValue{Value: "1"},
		DNSRecordCount:          APIValue{Value: "2"},
		DNSRecords: []DNSRecord{
			{
				ID:        APIValue{Value: "123456"},
				Name:      APIValue{Value: "*.mydomain.com"},
				Type:      APIValue{Value: "A"},
				Subdomain: APIValue{Value: "*"},
				Content:   APIValue{Value: "XX.XX.XX.XX"},
				TTL:       APIValue{Value: "86400"},
				Priority:  APIValue{Value: ""},
			},
			{
				ID:        APIValue{Value: "234567"},
				Name:      APIValue{Value: "mydomain.com"},
				Type:      APIValue{Value: "MX"},
				Subdomain: APIValue{Value: ""},
				Content:   APIValue{Value: "mx10.kundencontroller.de"},
				TTL:       APIValue{Value: "86400"},
				Priority:  APIValue{Value: "10"},
			},
		},
	}}

	assert.Equal(t, expected, domains)
}
