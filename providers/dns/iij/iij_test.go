package iij

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	liveTest                bool
	envTestAPIAccessKey     string
	envTestAPISecretKeyEnv  string
	envTestDoServiceCodeEnv string
	envTestDomain           string
)

func init() {
	envTestAPIAccessKey = os.Getenv("IIJ_API_ACCESS_KEY")
	envTestAPISecretKeyEnv = os.Getenv("IIJ_API_SECRET_KEY")
	envTestDoServiceCodeEnv = os.Getenv("IIJ_DO_SERVICE_CODE")

	envTestDomain = os.Getenv("IIJ_API_TESTDOMAIN")

	if len(envTestAPIAccessKey) > 0 && len(envTestAPISecretKeyEnv) > 0 && len(envTestDoServiceCodeEnv) > 0 && len(envTestDomain) > 0 {
		liveTest = true
	}
}

func restoreEnv() {
	os.Setenv("IIJ_API_ACCESS_KEY", envTestAPIAccessKey)
	os.Setenv("IIJ_API_SECRET_KEY", envTestAPISecretKeyEnv)
	os.Setenv("IIJ_DO_SERVICE_CODE", envTestDoServiceCodeEnv)
}

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				"IIJ_API_ACCESS_KEY":  "A",
				"IIJ_API_SECRET_KEY":  "B",
				"IIJ_DO_SERVICE_CODE": "C",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				"IIJ_API_ACCESS_KEY":  "",
				"IIJ_API_SECRET_KEY":  "",
				"IIJ_DO_SERVICE_CODE": "",
			},
			expected: "iij: some credentials information are missing: IIJ_API_ACCESS_KEY,IIJ_API_SECRET_KEY,IIJ_DO_SERVICE_CODE",
		},
		{
			desc: "missing api access key",
			envVars: map[string]string{
				"IIJ_API_ACCESS_KEY":  "",
				"IIJ_API_SECRET_KEY":  "B",
				"IIJ_DO_SERVICE_CODE": "C",
			},
			expected: "iij: some credentials information are missing: IIJ_API_ACCESS_KEY",
		},
		{
			desc: "missing secret key",
			envVars: map[string]string{
				"IIJ_API_ACCESS_KEY":  "A",
				"IIJ_API_SECRET_KEY":  "",
				"IIJ_DO_SERVICE_CODE": "C",
			},
			expected: "iij: some credentials information are missing: IIJ_API_SECRET_KEY",
		},
		{
			desc: "missing do service code",
			envVars: map[string]string{
				"IIJ_API_ACCESS_KEY":  "A",
				"IIJ_API_SECRET_KEY":  "B",
				"IIJ_DO_SERVICE_CODE": "",
			},
			expected: "iij: some credentials information are missing: IIJ_DO_SERVICE_CODE",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer restoreEnv()
			for key, value := range test.envVars {
				if len(value) == 0 {
					os.Unsetenv(key)
				} else {
					os.Setenv(key, value)
				}
			}

			p, err := NewDNSProvider()

			if len(test.expected) == 0 {
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
			defer restoreEnv()
			os.Unsetenv("IIJ_API_ACCESS_KEY")
			os.Unsetenv("IIJ_API_SECRET_KEY")
			os.Unsetenv("IIJ_DO_SERVICE_CODE")

			config := NewDefaultConfig()
			config.AccessKey = test.accessKey
			config.SecretKey = test.secretKey
			config.DoServiceCode = test.doServiceCode

			p, err := NewDNSProviderConfig(config)

			if len(test.expected) == 0 {
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
			desc:          "with a sub domain",
			domain:        "my.domain.com",
			zones:         []string{"domain.com"},
			expectedOwner: "_acme-challenge.my",
			expectedZone:  "domain.com",
		},
		{
			desc:          "with a sub domain in a zone",
			domain:        "my.sub.domain.com",
			zones:         []string{"sub.domain.com", "domain.com"},
			expectedOwner: "_acme-challenge.my",
			expectedZone:  "sub.domain.com",
		},
		{
			desc:          "with a sub sub domain",
			domain:        "my.sub.domain.com",
			zones:         []string{"domain1.com", "domain.com"},
			expectedOwner: "_acme-challenge.my.sub",
			expectedZone:  "domain.com",
		},
	}

	for _, test := range testCases {
		test := test
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
	if !liveTest {
		t.Skip("skipping live test")
	}

	restoreEnv()
	provider, err := NewDNSProvider()
	require.NoError(t, err)

	err = provider.Present(envTestDomain, "", "123d==")
	require.NoError(t, err)
}

func TestLiveCleanUp(t *testing.T) {
	if !liveTest {
		t.Skip("skipping live test")
	}

	restoreEnv()
	provider, err := NewDNSProvider()
	require.NoError(t, err)

	err = provider.CleanUp(envTestDomain, "", "123d==")
	require.NoError(t, err)
}
