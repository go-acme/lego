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
			client.baseURL, _ = url.Parse(server.URL)

			return client, nil
		},
		servermock.CheckHeader().WithJSONHeaders().
			WithAuthorization("user.secret"))
}

func TestClient_ListServices(t *testing.T) {
	client := mockBuilder().
		Route("GET /purchase", servermock.ResponseFromFixture("purchase.json")).
		Build(t)

	services, err := client.ListServices(t.Context())
	require.NoError(t, err)

	expected := []int{2018, 10039, 10128}

	assert.Equal(t, expected, services)
}

func TestClient_ListServices_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /purchase", servermock.ResponseFromFixture("error.json")).
		Build(t)

	_, err := client.ListServices(t.Context())
	require.EqualError(t, err, "code 2: Token di autorizzazione non valido")
}

func TestClient_ListServices_error_status(t *testing.T) {
	client := mockBuilder().
		Route("GET /purchase",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized)).
		Build(t)

	_, err := client.ListServices(t.Context())
	require.EqualError(t, err, "code 2: Token di autorizzazione non valido")
}

func TestClient_GetServiceDetails(t *testing.T) {
	client := mockBuilder().
		Route("GET /purchase/details/123", servermock.ResponseFromFixture("purchase-details.json")).
		Build(t)

	services, err := client.GetServiceDetails(t.Context(), 123)
	require.NoError(t, err)

	expected := &ServiceDetails{ID: 123, Name: "example", DomainID: 456}

	assert.Equal(t, expected, services)
}

func TestClient_GetServiceDetails_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /purchase/details/123", servermock.ResponseFromFixture("error.json")).
		Build(t)

	_, err := client.GetServiceDetails(t.Context(), 123)
	require.EqualError(t, err, "code 2: Token di autorizzazione non valido")
}

func TestClient_GetServiceDetails_error_status(t *testing.T) {
	client := mockBuilder().
		Route("GET /purchase/details/123",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized)).
		Build(t)

	_, err := client.GetServiceDetails(t.Context(), 123)
	require.EqualError(t, err, "code 2: Token di autorizzazione non valido")
}

func TestClient_GetDomainDetails(t *testing.T) {
	client := mockBuilder().
		Route("GET /domain/details/123", servermock.ResponseFromFixture("domain-details.json")).
		Build(t)

	services, err := client.GetDomainDetails(t.Context(), 123)
	require.NoError(t, err)

	expected := &DomainDetails{ID: 123, DomainName: "example.com", DomainNameASCII: "example.com"}

	assert.Equal(t, expected, services)
}

func TestClient_GetDomainDetails_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /domain/details/123", servermock.ResponseFromFixture("error.json")).
		Build(t)

	_, err := client.GetDomainDetails(t.Context(), 123)
	require.EqualError(t, err, "code 2: Token di autorizzazione non valido")
}

func TestClient_GetDomainDetails_error_status(t *testing.T) {
	client := mockBuilder().
		Route("GET /domain/details/123",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized)).
		Build(t)

	_, err := client.GetDomainDetails(t.Context(), 123)
	require.EqualError(t, err, "code 2: Token di autorizzazione non valido")
}

func TestClient_CreateRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /dns_record/store/123", servermock.ResponseFromFixture("dns_record-store.json")).
		Build(t)

	services, err := client.CreateRecord(t.Context(), 123, Record{})
	require.NoError(t, err)

	expected := 2255674

	assert.Equal(t, expected, services)
}

func TestClient_CreateRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /dns_record/store/123", servermock.ResponseFromFixture("error.json")).
		Build(t)

	_, err := client.CreateRecord(t.Context(), 123, Record{})
	require.EqualError(t, err, "code 2: Token di autorizzazione non valido")
}

func TestClient_CreateRecord_error_status(t *testing.T) {
	client := mockBuilder().
		Route("POST /dns_record/store/123",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized)).
		Build(t)

	_, err := client.CreateRecord(t.Context(), 123, Record{})
	require.EqualError(t, err, "code 2: Token di autorizzazione non valido")
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /dns_record/remove/123/456", servermock.ResponseFromFixture("dns_record-remove.json")).
		Build(t)

	err := client.DeleteRecord(t.Context(), 123, 456)
	require.NoError(t, err)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /dns_record/remove/123/456", servermock.ResponseFromFixture("error.json")).
		Build(t)

	err := client.DeleteRecord(t.Context(), 123, 456)
	require.EqualError(t, err, "code 2: Token di autorizzazione non valido")
}

func TestClient_DeleteRecord_error_status(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /dns_record/remove/123/456",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized)).
		Build(t)

	err := client.DeleteRecord(t.Context(), 123, 456)
	require.EqualError(t, err, "code 2: Token di autorizzazione non valido")
}

func TestTTLRounder(t *testing.T) {
	testCases := []struct {
		desc     string
		value    int
		expected int
	}{
		{
			desc:     "lower than 3600",
			value:    123,
			expected: 3600,
		},
		{
			desc:     "lower than 14400",
			value:    12341,
			expected: 14400,
		},
		{
			desc:     "lower than 28800",
			value:    28341,
			expected: 28800,
		},
		{
			desc:     "lower than 57600",
			value:    56600,
			expected: 57600,
		},
		{
			desc:     "rounded to 86400",
			value:    86000,
			expected: 86400,
		},
		{
			desc:     "default",
			value:    100000,
			expected: 3600,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			ttl := TTLRounder(test.value)

			assert.Equal(t, test.expected, ttl)
		})
	}
}
