package jdcloud

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(
	EnvAccessKeyID,
	EnvAccessKeySecret,
	EnvRegionID,
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
				EnvAccessKeyID:     "abc123",
				EnvAccessKeySecret: "secret",
			},
		},
		{
			desc: "missing access key ID",
			envVars: map[string]string{
				EnvAccessKeyID:     "",
				EnvAccessKeySecret: "secret",
			},
			expected: "jdcloud: some credentials information are missing: JDCLOUD_ACCESS_KEY_ID",
		},
		{
			desc: "missing access key secret",
			envVars: map[string]string{
				EnvAccessKeyID:     "abc123",
				EnvAccessKeySecret: "",
			},
			expected: "jdcloud: some credentials information are missing: JDCLOUD_ACCESS_KEY_SECRET",
		},
		{
			desc:     "missing credentials",
			envVars:  map[string]string{},
			expected: "jdcloud: some credentials information are missing: JDCLOUD_ACCESS_KEY_ID,JDCLOUD_ACCESS_KEY_SECRET",
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
		desc            string
		accessKeyID     string
		accessKeySecret string
		expected        string
	}{
		{
			desc:            "success",
			accessKeyID:     "abc123",
			accessKeySecret: "secret",
		},
		{
			desc:            "missing access key ID",
			accessKeySecret: "secret",
			expected:        "jdcloud: missing credentials",
		},
		{
			desc:        "missing access key secret",
			accessKeyID: "abc123",
			expected:    "jdcloud: missing credentials",
		},
		{
			desc:     "missing credentials",
			expected: "jdcloud: missing credentials",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.AccessKeyID = test.accessKeyID
			config.AccessKeySecret = test.accessKeySecret
			config.RegionID = "cn-north-1"

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

	err = provider.Present(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}

func TestLiveCleanUp(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()

	provider, err := NewDNSProvider()
	require.NoError(t, err)

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}

func mockBuilder() *servermock.Builder[*DNSProvider] {
	return servermock.NewBuilder(
		func(server *httptest.Server) (*DNSProvider, error) {
			config := NewDefaultConfig()
			config.AccessKeyID = "abc123"
			config.AccessKeySecret = "secret"
			config.RegionID = "cn-north-1"

			p, err := NewDNSProviderConfig(config)
			if err != nil {
				return nil, err
			}

			serverURL, _ := url.Parse(server.URL)

			p.client.Config.SetEndpoint(net.JoinHostPort(serverURL.Hostname(), serverURL.Port()))
			p.client.Config.SetScheme(serverURL.Scheme)
			p.client.Config.SetTimeout(server.Client().Timeout)

			return p, nil
		},
	)
}

func TestDNSProvider_Present(t *testing.T) {
	provider := mockBuilder().
		Route("GET /v2/regions/cn-north-1/domain",
			http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				pageNumber := req.URL.Query().Get("pageNumber")

				servermock.ResponseFromFixture(
					fmt.Sprintf("describe_domains_page%s.json", pageNumber),
				).ServeHTTP(rw, req)
			}),
			servermock.CheckQueryParameter().Strict().
				With("domainName", "example.com").
				WithRegexp("pageNumber", `(1|2)`).
				With("pageSize", "10"),
			servermock.CheckHeader().
				WithRegexp("Authorization",
					`JDCLOUD2-HMAC-SHA256 Credential=abc123/\d{8}/cn-north-1/domainservice/jdcloud2_request, SignedHeaders=content-type;host;x-jdcloud-date;x-jdcloud-nonce, Signature=\w+`).
				WithRegexp("X-Jdcloud-Date", `\d{8}T\d{6}Z`).
				WithRegexp("X-Jdcloud-Nonce", `[\w-]+`),
		).
		Route("POST /v2/regions/cn-north-1/domain/20/ResourceRecord",
			servermock.ResponseFromFixture("create_record.json"),
			servermock.CheckRequestJSONBodyFromFixture("create_record-request.json"),
			servermock.CheckHeader().
				WithRegexp("Authorization",
					`JDCLOUD2-HMAC-SHA256 Credential=abc123/\d{8}/cn-north-1/domainservice/jdcloud2_request, SignedHeaders=content-type;host;x-jdcloud-date;x-jdcloud-nonce, Signature=\w+`).
				WithRegexp("X-Jdcloud-Date", `\d{8}T\d{6}Z`).
				WithRegexp("X-Jdcloud-Nonce", `[\w-]+`),
		).
		Build(t)

	err := provider.Present("example.com", "abc", "123d==")
	require.NoError(t, err)

	require.Len(t, provider.domainIDs, 1)
	require.Len(t, provider.recordIDs, 1)

	assert.Equal(t, 20, provider.domainIDs["abc"])
	assert.Equal(t, 123, provider.recordIDs["abc"])
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockBuilder().
		Route("DELETE /v2/regions/cn-north-1/domain/20/ResourceRecord/123",
			servermock.ResponseFromFixture("delete_record.json"),
			servermock.CheckHeader().
				WithRegexp("Authorization",
					`JDCLOUD2-HMAC-SHA256 Credential=abc123/\d{8}/cn-north-1/domainservice/jdcloud2_request, SignedHeaders=content-type;host;x-jdcloud-date;x-jdcloud-nonce, Signature=\w+`).
				WithRegexp("X-Jdcloud-Date", `\d{8}T\d{6}Z`).
				WithRegexp("X-Jdcloud-Nonce", `[\w-]+`),
		).
		Build(t)

	provider.domainIDs["abc"] = 20
	provider.recordIDs["abc"] = 123

	err := provider.CleanUp("example.com", "abc", "123d==")
	require.NoError(t, err)
}
