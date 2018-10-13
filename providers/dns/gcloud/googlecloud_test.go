package gcloud

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/dns/v1"
)

var (
	liveTest                            bool
	envTestProject                      string
	envTestServiceAccountFile           string
	envTestGoogleApplicationCredentials string
	envTestDomain                       string
)

func init() {
	envTestProject = os.Getenv("GCE_PROJECT")
	envTestServiceAccountFile = os.Getenv("GCE_SERVICE_ACCOUNT_FILE")
	envTestGoogleApplicationCredentials = os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	envTestDomain = os.Getenv("GCE_DOMAIN")

	_, err := google.DefaultClient(context.Background(), dns.NdevClouddnsReadwriteScope)
	if err == nil && len(envTestProject) > 0 && len(envTestDomain) > 0 {
		liveTest = true
	}
}

func restoreEnv() {
	os.Setenv("GCE_PROJECT", envTestProject)
	os.Setenv("GCE_SERVICE_ACCOUNT_FILE", envTestServiceAccountFile)
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", envTestGoogleApplicationCredentials)
}

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "invalid credentials",
			envVars: map[string]string{
				"GCE_PROJECT":              "123",
				"GCE_SERVICE_ACCOUNT_FILE": "",
				// as Travis run on GCE, we have to alter env
				"GOOGLE_APPLICATION_CREDENTIALS": "not-a-secret-file",
			},
			expected: "googlecloud: unable to get Google Cloud client: google: error getting credentials using GOOGLE_APPLICATION_CREDENTIALS environment variable: open not-a-secret-file: no such file or directory",
		},
		{
			desc: "missing project",
			envVars: map[string]string{
				"GCE_PROJECT":              "",
				"GCE_SERVICE_ACCOUNT_FILE": "",
			},
			expected: "googlecloud: project name missing",
		},
		{
			desc: "success",
			envVars: map[string]string{
				"GCE_PROJECT":              "",
				"GCE_SERVICE_ACCOUNT_FILE": "fixtures/gce_account_service_file.json",
			},
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
		project  string
		expected string
	}{
		{
			desc:     "invalid project",
			project:  "123",
			expected: "googlecloud: unable to create Google Cloud DNS service: client is nil",
		},
		{
			desc:     "missing project",
			expected: "googlecloud: unable to create Google Cloud DNS service: client is nil",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer restoreEnv()
			os.Unsetenv("GCE_PROJECT")
			os.Unsetenv("GCE_SERVICE_ACCOUNT_FILE")

			config := NewDefaultConfig()
			config.Project = test.project

			p, err := NewDNSProviderConfig(config)

			if len(test.expected) == 0 {
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
	if !liveTest {
		t.Skip("skipping live test")
	}

	provider, err := NewDNSProviderCredentials(envTestProject)
	require.NoError(t, err)

	err = provider.Present(envTestDomain, "", "123d==")
	require.NoError(t, err)
}

func TestLivePresentMultiple(t *testing.T) {
	if !liveTest {
		t.Skip("skipping live test")
	}

	provider, err := NewDNSProviderCredentials(envTestProject)
	require.NoError(t, err)

	// Check that we're able to create multiple entries
	err = provider.Present(envTestDomain, "1", "123d==")
	require.NoError(t, err)

	err = provider.Present(envTestDomain, "2", "123d==")
	require.NoError(t, err)
}

func TestLiveCleanUp(t *testing.T) {
	if !liveTest {
		t.Skip("skipping live test")
	}

	provider, err := NewDNSProviderCredentials(envTestProject)
	require.NoError(t, err)

	time.Sleep(1 * time.Second)

	err = provider.CleanUp(envTestDomain, "", "123d==")
	require.NoError(t, err)
}
