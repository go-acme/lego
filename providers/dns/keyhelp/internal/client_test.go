package internal

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client, err := NewClient(server.URL, "secret")
			if err != nil {
				return nil, err
			}

			client.HTTPClient = server.Client()

			return client, nil
		},
		servermock.CheckHeader().
			With(APIKeyHeader, "secret").
			WithJSONHeaders(),
	)
}

func TestClient_ListDomains(t *testing.T) {
	client := mockBuilder().
		Route("GET /api/v2/domains",
			servermock.ResponseFromFixture("get_domains.json"),
			servermock.CheckQueryParameter().
				With("sort", "domain_utf8").
				Strict()).
		Build(t)

	domains, err := client.ListDomains(t.Context())
	require.NoError(t, err)

	expected := []Domain{{
		ID:             8,
		UserID:         4,
		ParentDomainID: 0,
		Status:         1,
		Domain:         "example.com",
		DomainUTF8:     "example.com",
		IsEmailDomain:  true,
	}}

	assert.Equal(t, expected, domains)
}

func TestClient_ListDomains_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /api/v2/domains",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized)).
		Build(t)

	_, err := client.ListDomains(t.Context())

	require.EqualError(t, err, "401 Unauthorized: API key is missing or invalid.")
}

func TestClient_ListDomainRecords(t *testing.T) {
	client := mockBuilder().
		Route("GET /api/v2/dns/123",
			servermock.ResponseFromFixture("get_domain_records.json")).
		Build(t)

	domainRecords, err := client.ListDomainRecords(t.Context(), 123)
	require.NoError(t, err)

	expected := &DomainRecords{
		DkimRecord: `default._domainkey IN TXT ( "v=DKIM1; k=rsa; s=email; " "...DKIM KEY..." )`,
		Records: &Records{
			Soa: &SOARecord{
				TTL:       86400,
				PrimaryNs: "ns.example.com.",
				RName:     "root.example.com.",
				Refresh:   14400,
				Retry:     1800,
				Expire:    604800,
				Minimum:   3600,
			},
			Other: []Record{{
				Host:  "@",
				TTL:   86400,
				Type:  "A",
				Value: "192.168.178.1",
			}},
		},
	}

	assert.Equal(t, expected, domainRecords)
}

func TestClient_ListDomainRecords_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /api/v2/dns/8",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized)).
		Build(t)

	_, err := client.ListDomainRecords(t.Context(), 8)

	require.EqualError(t, err, "401 Unauthorized: API key is missing or invalid.")
}

func TestClient_UpdateDomainRecords(t *testing.T) {
	client := mockBuilder().
		Route("PUT /api/v2/dns/8",
			servermock.ResponseFromFixture("update_domain_records.json"),
			servermock.CheckRequestJSONBodyFromFixture("update_domain_records-request.json")).
		Build(t)

	records := DomainRecords{
		DkimRecord: `default._domainkey IN TXT ( "v=DKIM1; k=rsa; s=email; " "...DKIM KEY..." )`,
		Records: &Records{
			Soa: &SOARecord{
				TTL:       86400,
				PrimaryNs: "ns.example.com.",
				RName:     "root.example.com.",
				Refresh:   14400,
				Retry:     1800,
				Expire:    604800,
				Minimum:   3600,
			},
			Other: []Record{
				{
					Host:  "@",
					TTL:   86400,
					Type:  "A",
					Value: "192.168.178.1",
				},
				{
					Host:  "_acme-challenge",
					TTL:   120,
					Type:  "TXT",
					Value: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
				},
			},
		},
	}

	domainID, err := client.UpdateDomainRecords(t.Context(), 8, records)
	require.NoError(t, err)

	expected := &DomainID{ID: 8}

	assert.Equal(t, expected, domainID)
}

func TestClient_UpdateDomainRecords_error(t *testing.T) {
	client := mockBuilder().
		Route("PUT /api/v2/dns/123",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized)).
		Build(t)

	records := DomainRecords{}

	_, err := client.UpdateDomainRecords(t.Context(), 123, records)

	require.EqualError(t, err, "401 Unauthorized: API key is missing or invalid.")
}
