package allinkl

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/go-acme/lego/v4/providers/dns/allinkl/internal"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvLogin, EnvPassword).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvLogin:    "user",
				EnvPassword: "secret",
			},
		},
		{
			desc: "missing credentials: account name",
			envVars: map[string]string{
				EnvLogin:    "",
				EnvPassword: "secret",
			},
			expected: "allinkl: some credentials information are missing: ALL_INKL_LOGIN",
		},
		{
			desc: "missing credentials: api key",
			envVars: map[string]string{
				EnvLogin:    "user",
				EnvPassword: "",
			},
			expected: "allinkl: some credentials information are missing: ALL_INKL_PASSWORD",
		},
		{
			desc: "missing credentials: all",
			envVars: map[string]string{
				EnvLogin:    "",
				EnvPassword: "",
			},
			expected: "allinkl: some credentials information are missing: ALL_INKL_LOGIN,ALL_INKL_PASSWORD",
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
		login    string
		password string
		expected string
	}{
		{
			desc:     "success",
			login:    "user",
			password: "secret",
		},
		{
			desc:     "missing account name",
			password: "secret",
			expected: "allinkl: missing credentials",
		},
		{
			desc:     "missing api key",
			login:    "user",
			expected: "allinkl: missing credentials",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.Login = test.login
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
			config.Login = "user"
			config.Password = "secret"
			config.HTTPClient = server.Client()

			p, err := NewDNSProviderConfig(config)
			if err != nil {
				return nil, err
			}

			p.client.BaseURL, _ = url.Parse(server.URL)
			p.identifier.BaseURL, _ = url.Parse(server.URL)

			return p, err
		},
	).Route("POST /KasAuth.php",
		servermock.ResponseFromInternal("auth.xml"),
		servermock.CheckRequestBodyFromInternal("auth-request.xml").
			IgnoreWhitespace(),
	)
}

func extractKasRequest(reader io.Reader) (*internal.KasRequest, error) {
	type ReqEnvelope struct {
		XMLName xml.Name `xml:"Envelope"`
		Body    struct {
			KasAPI struct {
				Params string `xml:"Params"`
			} `xml:"KasApi"`
		} `xml:"Body"`
	}

	raw, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	reqEnvelope := ReqEnvelope{}

	err = xml.Unmarshal(raw, &reqEnvelope)
	if err != nil {
		return nil, err
	}

	var kReq internal.KasRequest

	err = json.Unmarshal([]byte(reqEnvelope.Body.KasAPI.Params), &kReq)
	if err != nil {
		return nil, err
	}

	return &kReq, nil
}

func TestDNSProvider_Present(t *testing.T) {
	provider := mockBuilder().
		Route("POST /KasApi.php",
			http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				kReq, err := extractKasRequest(req.Body)
				if err != nil {
					http.Error(rw, err.Error(), http.StatusBadRequest)
					return
				}

				switch kReq.Action {
				case "get_dns_settings":
					params := kReq.RequestParams.(map[string]any)

					if params["zone_host"] == "_acme-challenge.example.com." {
						servermock.ResponseFromInternal("get_dns_settings_not_found.xml").ServeHTTP(rw, req)
					} else {
						servermock.ResponseFromInternal("get_dns_settings.xml").ServeHTTP(rw, req)
					}

				case "add_dns_settings":
					servermock.ResponseFromInternal("add_dns_settings.xml").ServeHTTP(rw, req)

				default:
					http.Error(rw, fmt.Sprintf("unknown action: %v", kReq.Action), http.StatusBadRequest)
				}
			}),
		).
		Build(t)

	err := provider.Present("example.com", "abc", "123d==")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockBuilder().
		Route("POST /KasApi.php",
			servermock.ResponseFromInternal("delete_dns_settings.xml"),
			servermock.CheckRequestBodyFromInternal("delete_dns_settings-request.xml").
				IgnoreWhitespace()).
		Build(t)

	provider.recordIDs["abc"] = "57347450"

	err := provider.CleanUp("example.com", "abc", "123d==")
	require.NoError(t, err)
}
