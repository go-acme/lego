package httpreq

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/require"
)

var envTest = tester.NewEnvTest(EnvEndpoint, EnvMode, EnvUsername, EnvPassword)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvEndpoint: "http://localhost:8090",
			},
		},
		{
			desc: "invalid URL",
			envVars: map[string]string{
				EnvEndpoint: ":",
			},
			expected: `httpreq: parse ":": missing protocol scheme`,
		},
		{
			desc: "missing endpoint",
			envVars: map[string]string{
				EnvEndpoint: "",
			},
			expected: "httpreq: some credentials information are missing: HTTPREQ_ENDPOINT",
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
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc     string
		endpoint *url.URL
		expected string
	}{
		{
			desc:     "success",
			endpoint: mustParse("http://localhost:8090"),
		},
		{
			desc:     "missing endpoint",
			expected: "httpreq: the endpoint is missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.Endpoint = test.endpoint

			p, err := NewDNSProviderConfig(config)

			if test.expected == "" {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.config)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestNewDNSProvider_Present(t *testing.T) {
	envTest.RestoreEnv()

	testCases := []struct {
		desc          string
		builder       *servermock.Builder[*DNSProvider]
		expectedError string
	}{
		{
			desc: "success",
			builder: mockBuilder("").
				Route("/present",
					servermock.RawStringResponse("lego"),
					servermock.CheckRequestJSONBody(`{"fqdn":"_acme-challenge.domain.","value":"LHDhK3oGRvkiefQnx7OOczTY5Tic_xZ6HcMOc_gmtoM"}`)),
		},
		{
			desc: "success with path prefix",
			builder: mockBuilderWithPathPrefix("", "/api/acme/").
				Route("/api/acme/present",
					servermock.RawStringResponse("lego"),
					servermock.CheckRequestJSONBody(`{"fqdn":"_acme-challenge.domain.","value":"LHDhK3oGRvkiefQnx7OOczTY5Tic_xZ6HcMOc_gmtoM"}`)),
		},
		{
			desc:          "error",
			builder:       mockBuilder(""),
			expectedError: "httpreq: unexpected status code: [status code: 404] body: 404 page not found",
		},
		{
			desc: "success raw mode",
			builder: mockBuilder("RAW").
				Route("/present",
					servermock.RawStringResponse("lego"),
					servermock.CheckRequestBody(`{"domain":"domain","token":"token","keyAuth":"key"}`)),
		},
		{
			desc:          "error raw mode",
			builder:       mockBuilder("RAW"),
			expectedError: "httpreq: unexpected status code: [status code: 404] body: 404 page not found",
		},
		{
			desc: "basic auth fail",
			builder: mockBuilderWithBasicAuth("nope", "nope").
				Route("/present", servermock.Noop()),
			expectedError: `httpreq: unexpected status code: [status code: 400] body: invalid credentials: got [username: "nope", password: "nope"], want [username: "user", password: "secret"]`,
		},
		{
			desc: "basic auth success",
			builder: mockBuilderWithBasicAuth("user", "secret").
				Route("/present",
					servermock.RawStringResponse("lego"),
					servermock.CheckRequestJSONBody(`{"fqdn":"_acme-challenge.domain.","value":"LHDhK3oGRvkiefQnx7OOczTY5Tic_xZ6HcMOc_gmtoM"}`)),
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			p := test.builder.Build(t)

			err := p.Present("domain", "token", "key")
			if test.expectedError == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, test.expectedError)
			}
		})
	}
}

func TestNewDNSProvider_Cleanup(t *testing.T) {
	envTest.RestoreEnv()

	testCases := []struct {
		desc          string
		builder       *servermock.Builder[*DNSProvider]
		expectedError string
	}{
		{
			desc: "success",
			builder: mockBuilder("").
				Route("/cleanup",
					servermock.RawStringResponse("lego"),
					servermock.CheckRequestJSONBody(`{"fqdn":"_acme-challenge.domain.","value":"LHDhK3oGRvkiefQnx7OOczTY5Tic_xZ6HcMOc_gmtoM"}`)),
		},
		{
			desc:          "error",
			builder:       mockBuilder(""),
			expectedError: "httpreq: unexpected status code: [status code: 404] body: 404 page not found",
		},
		{
			desc: "success raw mode",
			builder: mockBuilder("RAW").
				Route("/cleanup",
					servermock.RawStringResponse("lego"),
					servermock.CheckRequestBody(`{"domain":"domain","token":"token","keyAuth":"key"}`)),
		},
		{
			desc:          "error raw mode",
			builder:       mockBuilder("RAW"),
			expectedError: "httpreq: unexpected status code: [status code: 404] body: 404 page not found",
		},
		{
			desc: "basic auth fail",
			builder: mockBuilderWithBasicAuth("test", "example").
				Route("/cleanup", servermock.Noop()),
			expectedError: `httpreq: unexpected status code: [status code: 400] body: invalid credentials: got [username: "test", password: "example"], want [username: "user", password: "secret"]`,
		},
		{
			desc: "basic auth success",
			builder: mockBuilderWithBasicAuth("user", "secret").
				Route("/cleanup",
					servermock.RawStringResponse("lego"),
					servermock.CheckRequestJSONBody(`{"fqdn":"_acme-challenge.domain.","value":"LHDhK3oGRvkiefQnx7OOczTY5Tic_xZ6HcMOc_gmtoM"}`)),
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			p := test.builder.Build(t)

			err := p.CleanUp("domain", "token", "key")
			if test.expectedError == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, test.expectedError)
			}
		})
	}
}

func mockBuilder(mode string) *servermock.Builder[*DNSProvider] {
	return servermock.NewBuilder(
		func(server *httptest.Server) (*DNSProvider, error) {
			config := NewDefaultConfig()
			config.HTTPClient = server.Client()
			config.Endpoint, _ = url.Parse(server.URL)
			config.Mode = mode

			return NewDNSProviderConfig(config)
		})
}

func mockBuilderWithPathPrefix(mode, prefix string) *servermock.Builder[*DNSProvider] {
	return servermock.NewBuilder(
		func(server *httptest.Server) (*DNSProvider, error) {
			config := NewDefaultConfig()
			config.HTTPClient = server.Client()
			config.Endpoint, _ = url.Parse(server.URL + prefix)
			config.Mode = mode

			return NewDNSProviderConfig(config)
		})
}

func mockBuilderWithBasicAuth(username, password string) *servermock.Builder[*DNSProvider] {
	return servermock.NewBuilder(
		func(server *httptest.Server) (*DNSProvider, error) {
			config := NewDefaultConfig()
			config.HTTPClient = server.Client()
			config.Endpoint, _ = url.Parse(server.URL)
			config.Username = username
			config.Password = password

			return NewDNSProviderConfig(config)
		},
		servermock.CheckHeader().WithBasicAuth("user", "secret"))
}

func mustParse(rawURL string) *url.URL {
	uri, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}
	return uri
}
