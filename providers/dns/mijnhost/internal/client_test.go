package internal

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const apiKey = "secret"

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client := NewClient(apiKey)
			client.baseURL, _ = url.Parse(server.URL)
			client.HTTPClient = server.Client()

			return client, nil
		},
		servermock.CheckHeader().WithJSONHeaders().
			With(authorizationHeader, apiKey),
	)
}

func TestClient_ListDomains(t *testing.T) {
	client := mockBuilder().
		Route("GET /domains", servermock.ResponseFromFixture("list-domains.json")).
		Build(t)

	domains, err := client.ListDomains(t.Context())
	require.NoError(t, err)

	expected := []Domain{{
		ID:          1000,
		Domain:      "example.com",
		RenewalDate: "2030-01-01",
		Status:      "Active",
		StatusID:    1,
		Tags:        []string{"my-tag"},
	}}

	assert.Equal(t, expected, domains)
}

func TestClient_GetRecords(t *testing.T) {
	client := mockBuilder().
		Route("GET /domains/example.com/dns", servermock.ResponseFromFixture("get-dns-records.json")).
		Build(t)

	records, err := client.GetRecords(t.Context(), "example.com")
	require.NoError(t, err)

	expected := []Record{
		{
			Type:  "A",
			Name:  "example.com.",
			Value: "135.226.123.12",
			TTL:   900,
		},
		{
			Type:  "AAAA",
			Name:  "example.com.",
			Value: "2009:21d0:322:6100::5:c92b",
			TTL:   900,
		},
		{
			Type:  "MX",
			Name:  "example.com.",
			Value: "10 mail.example.com.",
			TTL:   900,
		},
		{
			Type:  "TXT",
			Name:  "example.com.",
			Value: "v=spf1 include:spf.mijn.host ~all",
			TTL:   900,
		},
	}

	assert.Equal(t, expected, records)
}

func TestClient_UpdateRecords(t *testing.T) {
	client := mockBuilder().
		Route("PUT /domains/example.com/dns",
			servermock.ResponseFromFixture("update-dns-records.json"),
			servermock.CheckRequestJSONBody(`{"records":[{"type":"TXT","name":"foo","value":"value1","ttl":120}]}`)).
		Build(t)

	records := []Record{{
		Type:  "TXT",
		Name:  "foo",
		Value: "value1",
		TTL:   120,
	}}

	err := client.UpdateRecords(t.Context(), "example.com", records)
	require.NoError(t, err)
}
