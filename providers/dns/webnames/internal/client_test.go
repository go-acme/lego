package internal

import (
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client := NewClient("secret")
			client.baseURL = server.URL
			client.HTTPClient = server.Client()

			return client, nil
		},
		servermock.CheckHeader().
			WithContentTypeFromURLEncoded(),
	)
}

func TestClient_AddTXTRecord(t *testing.T) {
	testCases := []struct {
		desc     string
		filename string
		require  require.ErrorAssertionFunc
	}{
		{
			desc:     "ok",
			filename: "ok.json",
			require:  require.NoError,
		},
		{
			desc:     "error",
			filename: "error.json",
			require:  require.Error,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			client := mockBuilder().
				Route("POST /",
					servermock.ResponseFromFixture(test.filename),
					servermock.CheckForm().Strict().
						With("domain", "example.com").
						With("type", "TXT").
						With("record", "foo:txtTXTtxt").
						With("action", "add").
						With("apikey", "secret"),
				).
				Build(t)

			domain := "example.com"
			subDomain := "foo"
			content := "txtTXTtxt"

			err := client.AddTXTRecord(t.Context(), domain, subDomain, content)
			test.require(t, err)
		})
	}
}

func TestClient_RemoveTxtRecord(t *testing.T) {
	testCases := []struct {
		desc     string
		filename string
		require  require.ErrorAssertionFunc
	}{
		{
			desc:     "ok",
			filename: "ok.json",
			require:  require.NoError,
		},
		{
			desc:     "error",
			filename: "error.json",
			require:  require.Error,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			client := mockBuilder().
				Route("POST /",
					servermock.ResponseFromFixture(test.filename),
					servermock.CheckForm().Strict().
						With("domain", "example.com").
						With("type", "TXT").
						With("record", "foo:txtTXTtxt").
						With("action", "delete").
						With("apikey", "secret"),
				).
				Build(t)

			domain := "example.com"
			subDomain := "foo"
			content := "txtTXTtxt"

			err := client.RemoveTXTRecord(t.Context(), domain, subDomain, content)
			test.require(t, err)
		})
	}
}
