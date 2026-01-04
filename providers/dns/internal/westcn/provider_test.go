package westcn

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/require"
)

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc     string
		username string
		password string
		expected string
	}{
		{
			desc:     "success",
			username: "user",
			password: "secret",
		},
		{
			desc:     "missing username",
			password: "secret",
			expected: "credentials missing",
		},
		{
			desc:     "missing password",
			username: "user",
			expected: "credentials missing",
		},
		{
			desc:     "missing credentials",
			expected: "credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := &Config{}
			config.Username = test.username
			config.Password = test.password

			p, err := NewDNSProviderConfig(config, "")

			if test.expected == "" {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.config)
				require.NotNil(t, p.client)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func mockBuilder() *servermock.Builder[*DNSProvider] {
	return servermock.NewBuilder(
		func(server *httptest.Server) (*DNSProvider, error) {
			config := &Config{
				Username:           "user",
				Password:           "secret",
				PropagationTimeout: 10 * time.Second,
				PollingInterval:    1 * time.Second,
				TTL:                120,
				HTTPClient:         server.Client(),
			}

			p, err := NewDNSProviderConfig(config, server.URL)
			if err != nil {
				return nil, err
			}

			return p, nil
		},
		servermock.CheckHeader().
			WithContentTypeFromURLEncoded())
}

func TestDNSProvider_Present(t *testing.T) {
	provider := mockBuilder().
		Route("POST /domain/",
			servermock.ResponseFromInternal("adddnsrecord.json").
				WithHeader("Content-Type", "application/json", "Charset=gb2312"),
			servermock.CheckQueryParameter().Strict().
				With("act", "adddnsrecord"),
			servermock.CheckForm().UsePostForm().Strict().
				With("domain", "example.com").
				With("host", "_acme-challenge").
				With("ttl", "120").
				With("type", "TXT").
				With("value", "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY").
				// With("act", "adddnsrecord").
				With("username", "user").
				WithRegexp("time", `\d+`).
				WithRegexp("token", `[a-z0-9]{32}`),
		).
		Build(t)

	err := provider.Present("example.com", "abc", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockBuilder().
		Route("POST /domain/",
			servermock.ResponseFromInternal("deldnsrecord.json").
				WithHeader("Content-Type", "application/json", "Charset=gb2312"),
			servermock.CheckQueryParameter().Strict().
				With("act", "deldnsrecord"),
			servermock.CheckForm().UsePostForm().Strict().
				With("id", "123").
				With("domain", "example.com").
				With("username", "user").
				WithRegexp("time", `\d+`).
				WithRegexp("token", `[a-z0-9]{32}`),
		).
		Build(t)

	provider.recordIDs["abc"] = 123

	err := provider.CleanUp("example.com", "abc", "123d==")
	require.NoError(t, err)
}
