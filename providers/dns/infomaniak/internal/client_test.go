package internal

import (
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/clientmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *clientmock.Builder[*Client] {
	return clientmock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client, err := New(OAuthStaticAccessToken(server.Client(), "token"), server.URL)
			if err != nil {
				return nil, err
			}

			return client, nil
		},
		clientmock.CheckHeader().WithJSONHeaders().
			WithAuthorization("Bearer token"))
}

func TestClient_CreateDNSRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /1/domain/666/dns/record",
			clientmock.RawStringResponse(`{"result":"success","data": "123"}`),
			clientmock.CheckRequestJSONBodyFromFile("create_dns_record-request.json")).
		Build(t)

	domain := &DNSDomain{
		ID:           666,
		CustomerName: "test",
	}

	record := Record{
		Source: "foo",
		Target: "txtxtxttxt",
		Type:   "TXT",
		TTL:    60,
	}

	recordID, err := client.CreateDNSRecord(t.Context(), domain, record)
	require.NoError(t, err)

	assert.Equal(t, "123", recordID)
}

func TestClient_GetDomainByName(t *testing.T) {
	client := mockBuilder().
		Route("GET /1/product",
			clientmock.ResponseFromFixture("get_domain_name.json"),
			clientmock.CheckQueryParameter().Strict().
				WithRegexp("customer_name", `.+\.example\.com`).
				With("service_name", "domain")).
		Build(t)

	domain, err := client.GetDomainByName(t.Context(), "one.two.three.example.com.")
	require.NoError(t, err)

	expected := &DNSDomain{ID: 123, CustomerName: "two.three.example.com"}
	assert.Equal(t, expected, domain)
}

func TestClient_DeleteDNSRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /1/domain/123/dns/record/456",
			clientmock.RawStringResponse(`{"result":"success"}`)).
		Build(t)

	err := client.DeleteDNSRecord(t.Context(), 123, "456")
	require.NoError(t, err)
}
