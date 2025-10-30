package liquidweb

import (
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/liquidweb/liquidweb-go/network"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(
	EnvURL,
	EnvUsername,
	EnvPassword,
	EnvZone).
	WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "minimum-success",
			envVars: map[string]string{
				EnvUsername: "blars",
				EnvPassword: "tacoman",
			},
		},
		{
			desc: "set-everything",
			envVars: map[string]string{
				EnvURL:      "https://storm.com",
				EnvUsername: "blars",
				EnvPassword: "tacoman",
				EnvZone:     "blars.com",
			},
		},
		{
			desc:     "missing credentials",
			envVars:  map[string]string{},
			expected: "liquidweb: some credentials information are missing: LIQUID_WEB_USERNAME,LIQUID_WEB_PASSWORD",
		},
		{
			desc: "missing username",
			envVars: map[string]string{
				EnvPassword: "tacoman",
				EnvZone:     "blars.com",
			},
			expected: "liquidweb: some credentials information are missing: LIQUID_WEB_USERNAME",
		},
		{
			desc: "missing password",
			envVars: map[string]string{
				EnvUsername: "blars",
				EnvZone:     "blars.com",
			},
			expected: "liquidweb: some credentials information are missing: LIQUID_WEB_PASSWORD",
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
				require.NotNil(t, p.recordIDs)
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
		zone     string
		expected string
	}{
		{
			desc:     "success",
			username: "acme",
			password: "secret",
			zone:     "example.com",
		},
		{
			desc:     "missing credentials",
			username: "",
			password: "",
			zone:     "",
			expected: "liquidweb: could not create Liquid Web API client: provided username is empty",
		},
		{
			desc:     "missing username",
			username: "",
			password: "secret",
			zone:     "example.com",
			expected: "liquidweb: could not create Liquid Web API client: provided username is empty",
		},
		{
			desc:     "missing password",
			username: "acme",
			password: "",
			zone:     "example.com",
			expected: "liquidweb: could not create Liquid Web API client: provided password is empty",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.Username = test.username
			config.Password = test.password
			config.Zone = test.zone

			p, err := NewDNSProviderConfig(config)

			if test.expected == "" {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.config)
				require.NotNil(t, p.client)
				require.NotNil(t, p.recordIDs)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestDNSProvider_Present(t *testing.T) {
	provider := mockProvider(t)

	err := provider.Present("tacoman.com", "", "")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockProvider(t, network.DNSRecord{
		Name:   "_acme-challenge.tacoman.com",
		RData:  "123d==",
		Type:   "TXT",
		TTL:    300,
		ID:     1234567,
		ZoneID: 42,
	})

	provider.recordIDs["123d=="] = 1234567

	err := provider.CleanUp("tacoman.com.", "123d==", "")
	require.NoError(t, err)
}

func TestDNSProvider(t *testing.T) {
	testCases := []struct {
		desc          string
		initRecs      []network.DNSRecord
		domain        string
		token         string
		keyAuth       string
		present       bool
		expPresentErr string
		cleanup       bool
	}{
		{
			desc:    "expected successful",
			domain:  "tacoman.com",
			token:   "123",
			keyAuth: "456",
			present: true,
			cleanup: true,
		},
		{
			desc:    "other successful",
			domain:  "banana.com",
			token:   "123",
			keyAuth: "456",
			present: true,
			cleanup: true,
		},
		{
			desc:          "zone not on account",
			domain:        "huckleberry.com",
			token:         "123",
			keyAuth:       "456",
			present:       true,
			expPresentErr: "no valid zone in account for certificate '_acme-challenge.huckleberry.com'",
			cleanup:       false,
		},
		{
			desc:    "ssl for domain",
			domain:  "sundae.cherry.com",
			token:   "5847953",
			keyAuth: "34872934",
			present: true,
			cleanup: true,
		},
		{
			desc:    "complicated domain",
			domain:  "always.money.stand.banana.com",
			token:   "5847953",
			keyAuth: "there is always money in the banana stand",
			present: true,
			cleanup: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			provider := mockProvider(t, test.initRecs...)

			if test.present {
				err := provider.Present(test.domain, test.token, test.keyAuth)
				if test.expPresentErr == "" {
					require.NoError(t, err)
				} else {
					require.ErrorContains(t, err, test.expPresentErr)
				}
			}

			if test.cleanup {
				err := provider.CleanUp(test.domain, test.token, test.keyAuth)
				require.NoError(t, err)
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

	time.Sleep(2 * time.Second)

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}
