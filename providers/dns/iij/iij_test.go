package iij

import (
	"testing"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "TESTDOMAIN"

var envTest = tester.NewEnvTest(
	EnvAPIAccessKey,
	EnvAPISecretKey,
	EnvDoServiceCode).
	WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvAPIAccessKey:  "A",
				EnvAPISecretKey:  "B",
				EnvDoServiceCode: "C",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				EnvAPIAccessKey:  "",
				EnvAPISecretKey:  "",
				EnvDoServiceCode: "",
			},
			expected: "iij: some credentials information are missing: IIJ_API_ACCESS_KEY,IIJ_API_SECRET_KEY,IIJ_DO_SERVICE_CODE",
		},
		{
			desc: "missing api access key",
			envVars: map[string]string{
				EnvAPIAccessKey:  "",
				EnvAPISecretKey:  "B",
				EnvDoServiceCode: "C",
			},
			expected: "iij: some credentials information are missing: IIJ_API_ACCESS_KEY",
		},
		{
			desc: "missing secret key",
			envVars: map[string]string{
				EnvAPIAccessKey:  "A",
				EnvAPISecretKey:  "",
				EnvDoServiceCode: "C",
			},
			expected: "iij: some credentials information are missing: IIJ_API_SECRET_KEY",
		},
		{
			desc: "missing do service code",
			envVars: map[string]string{
				EnvAPIAccessKey:  "A",
				EnvAPISecretKey:  "B",
				EnvDoServiceCode: "",
			},
			expected: "iij: some credentials information are missing: IIJ_DO_SERVICE_CODE",
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
				require.NotNil(t, p.api)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc          string
		accessKey     string
		secretKey     string
		doServiceCode string
		expected      string
	}{
		{
			desc:          "success",
			accessKey:     "A",
			secretKey:     "B",
			doServiceCode: "C",
		},
		{
			desc:     "missing credentials",
			expected: "iij: credentials missing",
		},
		{
			desc:          "missing access key",
			accessKey:     "",
			secretKey:     "B",
			doServiceCode: "C",
			expected:      "iij: credentials missing",
		},
		{
			desc:          "missing secret key",
			accessKey:     "A",
			secretKey:     "",
			doServiceCode: "C",
			expected:      "iij: credentials missing",
		},
		{
			desc:          "missing do service code",
			accessKey:     "A",
			secretKey:     "B",
			doServiceCode: "",
			expected:      "iij: credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.AccessKey = test.accessKey
			config.SecretKey = test.secretKey
			config.DoServiceCode = test.doServiceCode

			p, err := NewDNSProviderConfig(config)

			if test.expected == "" {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.config)
				require.NotNil(t, p.api)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestSplitDomain(t *testing.T) {
	testCases := []struct {
		desc          string
		domain        string
		zones         []string
		expectedOwner string
		expectedZone  string
	}{
		{
			desc:          "domain equals zone",
			domain:        "domain.com",
			zones:         []string{"domain.com"},
			expectedOwner: "_acme-challenge",
			expectedZone:  "domain.com",
		},
		{
			desc:          "with a subdomain",
			domain:        "my.domain.com",
			zones:         []string{"domain.com"},
			expectedOwner: "_acme-challenge.my",
			expectedZone:  "domain.com",
		},
		{
			desc:          "with a subdomain in a zone",
			domain:        "my.sub.domain.com",
			zones:         []string{"sub.domain.com", "domain.com"},
			expectedOwner: "_acme-challenge.my",
			expectedZone:  "sub.domain.com",
		},
		{
			desc:          "with a sub-subdomain",
			domain:        "my.sub.domain.com",
			zones:         []string{"domain1.com", "domain.com"},
			expectedOwner: "_acme-challenge.my.sub",
			expectedZone:  "domain.com",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			owner, zone, err := splitDomain(test.domain, test.zones)
			require.NoError(t, err)

			assert.Equal(t, test.expectedOwner, owner)
			assert.Equal(t, test.expectedZone, zone)
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
