package ucloud

import (
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/go-acme/lego/v5/internal/tester"
	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(
	EnvPrivateKey,
	EnvPublicKey,
	EnvRegion,
	EnvProjectID,
).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvPrivateKey: "xxx",
				EnvPublicKey:  "yyy",
			},
		},
		{
			desc: "missing private key",
			envVars: map[string]string{
				EnvPrivateKey: "",
				EnvPublicKey:  "yyy",
			},
			expected: "ucloud: some credentials information are missing: UCLOUD_PRIVATE_KEY",
		},
		{
			desc: "missing public key",
			envVars: map[string]string{
				EnvPrivateKey: "xxx",
				EnvPublicKey:  "",
			},
			expected: "ucloud: some credentials information are missing: UCLOUD_PUBLIC_KEY",
		},
		{
			desc:     "missing credentials",
			envVars:  map[string]string{},
			expected: "ucloud: some credentials information are missing: UCLOUD_PUBLIC_KEY,UCLOUD_PRIVATE_KEY",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer envTest.RestoreEnv()

			envTest.ClearEnv()

			envTest.Apply(test.envVars)

			p, err := NewDNSProvider()

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

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc       string
		privateKey string
		publicKey  string
		expected   string
	}{
		{
			desc:       "success",
			privateKey: "xxx",
			publicKey:  "yyy",
		},
		{
			desc:      "missing private key",
			publicKey: "yyy",
			expected:  "ucloud: credentials missing",
		},
		{
			desc:       "missing public key",
			privateKey: "xxx",
			expected:   "ucloud: credentials missing",
		},
		{
			desc:     "missing credentials",
			expected: "ucloud: credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.PrivateKey = test.privateKey
			config.PublicKey = test.publicKey

			p, err := NewDNSProviderConfig(config)

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

func TestLivePresent(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()

	provider, err := NewDNSProvider()
	require.NoError(t, err)

	err = provider.Present(t.Context(), envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}

func TestLiveCleanUp(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()

	provider, err := NewDNSProvider()
	require.NoError(t, err)

	err = provider.CleanUp(t.Context(), envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}

func mockBuilder() *servermock.Builder[*DNSProvider] {
	return servermock.NewBuilder(
		func(server *httptest.Server) (*DNSProvider, error) {
			config := NewDefaultConfig()
			config.PrivateKey = "privkey"
			config.PublicKey = "pubkey"

			config.baseURL = server.URL

			p, err := NewDNSProviderConfig(config)
			if err != nil {
				return nil, err
			}

			p.client.SetTransport(server.Client().Transport)

			return p, nil
		},
		servermock.CheckHeader().
			WithRegexp("U-Timestamp-Ms", `\d+`).
			WithContentTypeFromURLEncoded(),
	)
}

func TestDNSProvider_Present(t *testing.T) {
	provider := mockBuilder().
		Route("/",
			subRouter().
				Route(matchAction("UdnrDomainDNSAdd"),
					servermock.ResponseFromInternal("udnrDomainDNSAdd.json"),
					servermock.CheckQueryParameter().Strict().
						With("Action", "UdnrDomainDNSAdd"),
					servermock.CheckForm().Strict().
						WithRegexp("Action", "UdnrDomainDNSAdd").
						With("Dn", "example.com").
						With("RecordName", "_acme-challenge.example.com").
						With("TTL", "600").
						With("DnsType", "TXT").
						With("Content", "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY").
						With("PublicKey", "pubkey").
						WithRegexp("Signature", ".+"),
				),
		).
		Build(t)

	err := provider.Present(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockBuilder().
		Route("/",
			subRouter().
				Route(matchAction("UdnrDomainDNSQuery"),
					servermock.ResponseFromInternal("udnrDomainDNSQuery.json"),
					servermock.CheckQueryParameter().Strict().
						With("Action", "UdnrDomainDNSQuery"),
					servermock.CheckForm().Strict().
						WithRegexp("Action", "UdnrDomainDNSQuery").
						With("Dn", "example.com").
						With("PublicKey", "pubkey").
						WithRegexp("Signature", ".+"),
				).
				Route(matchAction("UdnrDeleteDnsRecord"),
					servermock.ResponseFromInternal("udnrDeleteDNSRecord.json"),
					servermock.CheckQueryParameter().Strict().
						With("Action", "UdnrDeleteDnsRecord"),
					servermock.CheckForm().Strict().
						WithRegexp("Action", "UdnrDeleteDnsRecord").
						With("Dn", "example.com").
						With("RecordName", "_acme-challenge.example.com").
						With("DnsType", "TXT").
						With("Content", "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY").
						With("PublicKey", "pubkey").
						WithRegexp("Signature", ".+"),
				),
		).
		Build(t)

	err := provider.CleanUp(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)
}

func matchAction(action string) func(req *http.Request) bool {
	return func(req *http.Request) bool {
		return req.URL.Query().Get("Action") == action
	}
}

// NOTE(ldez): this idea can be reused.
type subRoute struct {
	matcher func(req *http.Request) bool
	handler http.Handler
}

// NOTE(ldez): this idea can be reused.
type sr struct {
	subRoutes []*subRoute
}

func subRouter() *sr {
	return &sr{}
}

func (s *sr) Route(matcher func(req *http.Request) bool, handler http.Handler, chain ...servermock.Link) *sr {
	for _, link := range slices.Backward(chain) {
		handler = link.Bind(handler)
	}

	s.subRoutes = append(s.subRoutes, &subRoute{
		matcher: matcher,
		handler: handler,
	})

	return s
}

func (s *sr) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	for _, route := range s.subRoutes {
		if route.matcher(req) {
			route.handler.ServeHTTP(rw, req)
			return
		}
	}

	rw.WriteHeader(http.StatusNotFound)
}
