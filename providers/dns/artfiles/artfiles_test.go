package artfiles

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v5/internal/tester"
	servermock2 "github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvUsername, EnvPassword).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvUsername: "user",
				EnvPassword: "secret",
			},
		},
		{
			desc: "missing username",
			envVars: map[string]string{
				EnvUsername: "",
				EnvPassword: "secret",
			},
			expected: "artfiles: some credentials information are missing: ARTFILES_USERNAME",
		},
		{
			desc: "missing password",
			envVars: map[string]string{
				EnvUsername: "user",
				EnvPassword: "",
			},
			expected: "artfiles: some credentials information are missing: ARTFILES_PASSWORD",
		},
		{
			desc:     "missing credentials",
			envVars:  map[string]string{},
			expected: "artfiles: some credentials information are missing: ARTFILES_USERNAME,ARTFILES_PASSWORD",
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
			expected: "artfiles: credentials missing",
		},
		{
			desc:     "missing Example",
			username: "user",
			expected: "artfiles: credentials missing",
		},
		{
			desc:     "missing credentials",
			expected: "artfiles: credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.Username = test.username
			config.Password = test.password

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

func mockBuilder() *servermock2.Builder[*DNSProvider] {
	return servermock2.NewBuilder(
		func(server *httptest.Server) (*DNSProvider, error) {
			config := NewDefaultConfig()
			config.Username = "user"
			config.Password = "secret"
			config.HTTPClient = server.Client()

			p, err := NewDNSProviderConfig(config)
			if err != nil {
				return nil, err
			}

			p.client.BaseURL, _ = url.Parse(server.URL)

			return p, nil
		},
		servermock2.CheckHeader().
			WithBasicAuth("user", "secret"),
	)
}

func TestDNSProvider_Present(t *testing.T) {
	provider := mockBuilder().
		Route("GET /domain/get_domains.html",
			servermock2.ResponseFromInternal("domains.txt"),
		).
		Route("GET /dns/get_dns.html",
			servermock2.ResponseFromInternal("get_dns.json"),
			servermock2.CheckQueryParameter().Strict().
				With("domain", "example.com"),
		).
		Route("POST /dns/set_dns.html",
			servermock2.ResponseFromInternal("set_dns.json"),
			servermock2.CheckQueryParameter().Strict().
				With("TXT", `@ "v=spf1 a mx ~all"
_acme-challenge "TheAcmeChallenge"
_acme-challenge "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY"
_dmarc "v=DMARC1;p=reject;sp=reject;adkim=r;aspf=r;pct=100;rua=mailto:someone@in.mailhardener.com,mailto:postmaster@example.tld;ri=86400;ruf=mailto:someone@in.mailhardener.com,mailto:postmaster@example.tld;fo=1;rf=afrf"
_mta-sts "v=STSv1;id=yyyymmddTHHMMSS;"
_smtp._tls "v=TLSRPTv1;rua=mailto:someone@in.mailhardener.com"
selector._domainkey "v=DKIM1;k=rsa;p=Base64Stuff" "MoreBase64Stuff" "Even++MoreBase64Stuff" "YesMoreBase64Stuff" "And+Yes+Even+MoreBase64Stuff" "Sure++MoreBase64Stuff" "LastBase64Stuff"
selectorecc._domainkey "v=DKIM1;k=ed25519;p=Base64Stuff"`).
				With("domain", "example.com"),
		).
		Build(t)

	err := provider.Present(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockBuilder().
		Route("GET /domain/get_domains.html",
			servermock2.ResponseFromInternal("domains.txt"),
		).
		Route("GET /dns/get_dns.html",
			servermock2.ResponseFromInternal("get_dns.json"),
			servermock2.CheckQueryParameter().Strict().
				With("domain", "example.com"),
		).
		Route("POST /dns/set_dns.html",
			servermock2.ResponseFromInternal("set_dns.json"),
			servermock2.CheckQueryParameter().Strict().
				With("TXT", `@ "v=spf1 a mx ~all"
_acme-challenge "TheAcmeChallenge"
_dmarc "v=DMARC1;p=reject;sp=reject;adkim=r;aspf=r;pct=100;rua=mailto:someone@in.mailhardener.com,mailto:postmaster@example.tld;ri=86400;ruf=mailto:someone@in.mailhardener.com,mailto:postmaster@example.tld;fo=1;rf=afrf"
_mta-sts "v=STSv1;id=yyyymmddTHHMMSS;"
_smtp._tls "v=TLSRPTv1;rua=mailto:someone@in.mailhardener.com"
selector._domainkey "v=DKIM1;k=rsa;p=Base64Stuff" "MoreBase64Stuff" "Even++MoreBase64Stuff" "YesMoreBase64Stuff" "And+Yes+Even+MoreBase64Stuff" "Sure++MoreBase64Stuff" "LastBase64Stuff"
selectorecc._domainkey "v=DKIM1;k=ed25519;p=Base64Stuff"`).
				With("domain", "example.com"),
		).
		Build(t)

	err := provider.CleanUp(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)
}
