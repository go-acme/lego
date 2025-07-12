package internal

import (
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupClient(server *httptest.Server) (*Client, error) {
	client := NewClient("secret")
	client.baseURL = server.URL
	client.HTTPClient = server.Client()

	return client, nil
}

func TestClient_AddTXTRecord(t *testing.T) {
	client := servermock.NewBuilder[*Client](setupClient).
		Route("GET /", servermock.RawStringResponse("OK"),
			servermock.CheckQueryParameter().Strict().
				With("clear", "false").
				With("domains", "com").
				With("token", "secret").
				With("txt", "value")).
		Build(t)

	err := client.AddTXTRecord(t.Context(), "example.com", "value")
	require.NoError(t, err)
}

func TestClient_RemoveTXTRecord(t *testing.T) {
	client := servermock.NewBuilder[*Client](setupClient).
		Route("GET /", servermock.RawStringResponse("OK"),
			servermock.CheckQueryParameter().Strict().
				With("clear", "true").
				With("domains", "com").
				With("token", "secret").
				With("txt", "")).
		Build(t)

	err := client.RemoveTXTRecord(t.Context(), "example.com")
	require.NoError(t, err)
}

func Test_getMainDomain(t *testing.T) {
	testCases := []struct {
		desc     string
		domain   string
		expected string
	}{
		{
			desc:     "empty",
			domain:   "",
			expected: "",
		},
		{
			desc:     "missing sub domain",
			domain:   "duckdns.org",
			expected: "",
		},
		{
			desc:     "explicit domain: sub domain",
			domain:   "_acme-challenge.sub.duckdns.org",
			expected: "sub.duckdns.org",
		},
		{
			desc:     "explicit domain: subsub domain",
			domain:   "_acme-challenge.my.sub.duckdns.org",
			expected: "sub.duckdns.org",
		},
		{
			desc:     "explicit domain: subsubsub domain",
			domain:   "_acme-challenge.my.sub.sub.duckdns.org",
			expected: "sub.duckdns.org",
		},
		{
			desc:     "only subname: sub domain",
			domain:   "_acme-challenge.sub",
			expected: "sub",
		},
		{
			desc:     "only subname: subsub domain",
			domain:   "_acme-challenge.my.sub",
			expected: "sub",
		},
		{
			desc:     "only subname: subsubsub domain",
			domain:   "_acme-challenge.my.sub.sub",
			expected: "sub",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			wDomain := getMainDomain(test.domain)
			assert.Equal(t, test.expected, wDomain)
		})
	}
}
