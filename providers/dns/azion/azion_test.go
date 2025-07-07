package azion

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/aziontech/azionapi-go-sdk/idns"
	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvPersonalToken).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvPersonalToken: "token",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				EnvPersonalToken: "",
			},
			expected: "azion: some credentials information are missing: AZION_PERSONAL_TOKEN",
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
		token    string
		expected string
	}{
		{
			desc:  "success",
			token: "token",
		},
		{
			desc:     "missing credentials",
			expected: "azion: missing credentials",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.PersonalToken = test.token

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

func TestDNSProvider_findZone(t *testing.T) {
	provider, mux := setupTest(t)
	mux.HandleFunc("GET /intelligent_dns", writeFixtureHandler("zones.json"))

	testCases := []struct {
		desc     string
		fqdn     string
		expected *idns.Zone
	}{
		{
			desc: "apex",
			fqdn: "example.com.",
			expected: &idns.Zone{
				Id:     idns.PtrInt32(1),
				Domain: idns.PtrString("example.com"),
			},
		},
		{
			desc: "sub domain",
			fqdn: "sub.example.com.",
			expected: &idns.Zone{
				Id:     idns.PtrInt32(2),
				Domain: idns.PtrString("sub.example.com"),
			},
		},
		{
			desc: "long sub domain",
			fqdn: "_acme-challenge.api.sub.example.com.",
			expected: &idns.Zone{
				Id:     idns.PtrInt32(2),
				Domain: idns.PtrString("sub.example.com"),
			},
		},
		{
			desc: "long sub domain, apex",
			fqdn: "_acme-challenge.test.example.com.",
			expected: &idns.Zone{
				Id:     idns.PtrInt32(1),
				Domain: idns.PtrString("example.com"),
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			zone, err := provider.findZone(context.Background(), test.fqdn)
			require.NoError(t, err)

			assert.Equal(t, test.expected, zone)
		})
	}
}

func TestDNSProvider_findZone_error(t *testing.T) {
	testCases := []struct {
		desc     string
		fqdn     string
		response string
		expected string
	}{
		{
			desc:     "no parent zone found",
			fqdn:     "_acme-challenge.example.org.",
			response: "zones.json",
			expected: `zone not found (fqdn: "_acme-challenge.example.org.")`,
		},
		{
			desc:     "empty zones list",
			fqdn:     "example.com.",
			response: "zones_empty.json",
			expected: `zone not found (fqdn: "example.com.")`,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			provider, mux := setupTest(t)
			mux.HandleFunc("GET /intelligent_dns", writeFixtureHandler(test.response))

			zone, err := provider.findZone(context.Background(), test.fqdn)
			require.EqualError(t, err, test.expected)

			assert.Nil(t, zone)
		})
	}
}

func setupTest(t *testing.T) (*DNSProvider, *http.ServeMux) {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)

	config := NewDefaultConfig()
	config.PersonalToken = "secret"

	provider, err := NewDNSProviderConfig(config)
	require.NoError(t, err)

	clientConfig := provider.client.GetConfig()
	clientConfig.HTTPClient = server.Client()
	clientConfig.Servers = idns.ServerConfigurations{
		{
			URL:         server.URL,
			Description: "Production",
		},
	}

	return provider, mux
}

func writeFixtureHandler(filename string) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "application/json")

		file, err := os.Open(filepath.Join("fixtures", filename))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		defer func() { _ = file.Close() }()

		_, _ = io.Copy(rw, file)
	}
}
