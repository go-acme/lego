package internal

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/clientmock"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *clientmock.Builder[*Client] {
	return clientmock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client := NewClient("user", "secret")
			client.baseURL, _ = url.Parse(server.URL)

			return client, nil
		},
		clientmock.CheckHeader().
			WithContentTypeFromURLEncoded(),
	)
}

func TestRemoveRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /zone/remove_record",
			clientmock.ResponseFromFixture("remove_record.json"),
			clientmock.CheckForm().Strict().
				With("input_data", `{"domains":[{"dname":"test.ru"}],"subdomain":"_acme-challenge","content":"txttxttxt","record_type":"TXT","output_content_type":"plain"}`).
				With("username", "user").
				With("password", "secret").
				With("input_format", "json")).
		Build(t)

	err := client.RemoveTxtRecord(t.Context(), "test.ru", "_acme-challenge", "txttxttxt")
	require.NoError(t, err)
}

func TestRemoveRecord_errors(t *testing.T) {
	testCases := []struct {
		desc     string
		domain   string
		response string
		expected string
	}{
		{
			desc:     "authentication failed",
			domain:   "test.ru",
			response: "remove_record_error_auth.json",
			expected: "API error: NO_AUTH: No authorization mechanism selected",
		},
		{
			desc:     "domain error",
			domain:   "",
			response: "remove_record_error_domain.json",
			expected: "API error: NO_DOMAIN: domain_name not given or empty",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			client := mockBuilder().
				Route("POST /zone/remove_record", clientmock.ResponseFromFixture(test.response)).
				Build(t)

			err := client.RemoveTxtRecord(t.Context(), test.domain, "_acme-challenge", "txttxttxt")
			require.EqualError(t, err, test.expected)
		})
	}
}

func TestAddTXTRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /zone/add_txt",
			clientmock.ResponseFromFixture("add_txt_record.json"),
			clientmock.CheckForm().Strict().
				With("input_data", `{"domains":[{"dname":"test.ru"}],"subdomain":"_acme-challenge","text":"txttxttxt","output_content_type":"plain"}`).
				With("username", "user").
				With("password", "secret").
				With("input_format", "json")).
		Build(t)

	err := client.AddTXTRecord(t.Context(), "test.ru", "_acme-challenge", "txttxttxt")
	require.NoError(t, err)
}

func TestAddTXTRecord_errors(t *testing.T) {
	testCases := []struct {
		desc     string
		domain   string
		response string
		expected string
	}{
		{
			desc:     "authentication failed",
			domain:   "test.ru",
			response: "add_txt_record_error_auth.json",
			expected: "API error: NO_AUTH: No authorization mechanism selected",
		},
		{
			desc:     "domain error",
			domain:   "",
			response: "add_txt_record_error_domain.json",
			expected: "API error: NO_DOMAIN: domain_name not given or empty",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			client := mockBuilder().
				Route("POST /zone/add_txt", clientmock.ResponseFromFixture(test.response)).
				Build(t)

			err := client.AddTXTRecord(t.Context(), test.domain, "_acme-challenge", "txttxttxt")
			require.EqualError(t, err, test.expected)
		})
	}
}
