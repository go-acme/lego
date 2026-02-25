package internal

import (
	"net/http/httptest"
	"testing"

	servermock2 "github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/assert"
)

func setupClient(server *httptest.Server) (*Client, error) {
	client := NewClient(map[string]string{"example.com": "secret"})
	client.baseURL = server.URL
	client.HTTPClient = server.Client()

	return client, nil
}

func TestClient_UpdateTxtRecord(t *testing.T) {
	testCases := []struct {
		code     string
		expected assert.ErrorAssertionFunc
	}{
		{
			code:     codeGood,
			expected: assert.NoError,
		},
		{
			code:     codeNoChg + ` "0123456789abcdef"`,
			expected: assert.NoError,
		},
		{
			code:     codeAbuse,
			expected: assert.Error,
		},
		{
			code:     codeBadAgent,
			expected: assert.Error,
		},
		{
			code:     codeBadAuth,
			expected: assert.Error,
		},
		{
			code:     codeNoHost,
			expected: assert.Error,
		},
		{
			code:     codeNotFqdn,
			expected: assert.Error,
		},
	}

	for _, test := range testCases {
		t.Run(test.code, func(t *testing.T) {
			t.Parallel()

			client := servermock2.NewBuilder[*Client](setupClient, servermock2.CheckHeader().WithContentTypeFromURLEncoded()).
				Route("POST /",
					servermock2.RawStringResponse(test.code),
					servermock2.CheckForm().Strict().
						With("hostname", "_acme-challenge.example.com").
						With("password", "secret").
						With("txt", "foo")).
				Build(t)

			err := client.UpdateTxtRecord(t.Context(), "example.com", "_acme-challenge.example.com", "foo")
			test.expected(t, err)
		})
	}
}
