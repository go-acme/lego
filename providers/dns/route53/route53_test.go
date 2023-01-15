package route53

import (
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const envDomain = "R53_DOMAIN"

var envTest = tester.NewEnvTest(
	EnvAccessKeyID,
	EnvSecretAccessKey,
	EnvRegion,
	EnvHostedZoneID,
	EnvMaxRetries,
	EnvTTL,
	EnvPropagationTimeout,
	EnvPollingInterval).
	WithDomain(envDomain).
	WithLiveTestRequirements(EnvAccessKeyID, EnvSecretAccessKey, EnvRegion, envDomain)

func makeTestProvider(t *testing.T, serverURL string) *DNSProvider {
	t.Helper()

	config := &aws.Config{
		Credentials: credentials.NewStaticCredentials("abc", "123", " "),
		Endpoint:    aws.String(serverURL),
		Region:      aws.String("mock-region"),
		MaxRetries:  aws.Int(1),
	}

	sess, err := session.NewSession(config)
	require.NoError(t, err)

	return &DNSProvider{
		client: route53.New(sess),
		config: NewDefaultConfig(),
	}
}

func Test_loadCredentials_FromEnv(t *testing.T) {
	defer envTest.RestoreEnv()
	envTest.ClearEnv()

	_ = os.Setenv(EnvAccessKeyID, "123")
	_ = os.Setenv(EnvSecretAccessKey, "456")
	_ = os.Setenv(EnvRegion, "us-east-1")

	config := &aws.Config{
		CredentialsChainVerboseErrors: aws.Bool(true),
	}

	sess, err := session.NewSession(config)
	require.NoError(t, err)

	value, err := sess.Config.Credentials.Get()
	require.NoError(t, err, "Expected credentials to be set from environment")

	expected := credentials.Value{
		AccessKeyID:     "123",
		SecretAccessKey: "456",
		SessionToken:    "",
		ProviderName:    "EnvConfigCredentials",
	}
	assert.Equal(t, expected, value)
}

func Test_loadRegion_FromEnv(t *testing.T) {
	defer envTest.RestoreEnv()
	envTest.ClearEnv()

	os.Setenv(EnvRegion, route53.CloudWatchRegionUsEast1)

	sess, err := session.NewSession(aws.NewConfig())
	require.NoError(t, err)

	region := aws.StringValue(sess.Config.Region)
	assert.Equal(t, route53.CloudWatchRegionUsEast1, region, "Region")
}

func Test_getHostedZoneID_FromEnv(t *testing.T) {
	defer envTest.RestoreEnv()
	envTest.ClearEnv()

	expectedZoneID := "zoneID"

	os.Setenv(EnvHostedZoneID, expectedZoneID)

	provider, err := NewDNSProvider()
	require.NoError(t, err)

	hostedZoneID, err := provider.getHostedZoneID("whatever")
	require.NoError(t, err, "HostedZoneID")

	assert.Equal(t, expectedZoneID, hostedZoneID)
}

func TestNewDefaultConfig(t *testing.T) {
	defer envTest.RestoreEnv()

	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected *Config
	}{
		{
			desc: "default configuration",
			expected: &Config{
				MaxRetries:         5,
				TTL:                10,
				PropagationTimeout: 2 * time.Minute,
				PollingInterval:    4 * time.Second,
			},
		},
		{
			desc: "",
			envVars: map[string]string{
				EnvMaxRetries:         "10",
				EnvTTL:                "99",
				EnvPropagationTimeout: "60",
				EnvPollingInterval:    "60",
				EnvHostedZoneID:       "abc123",
			},
			expected: &Config{
				MaxRetries:         10,
				TTL:                99,
				PropagationTimeout: 60 * time.Second,
				PollingInterval:    60 * time.Second,
				HostedZoneID:       "abc123",
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			envTest.ClearEnv()
			for key, value := range test.envVars {
				os.Setenv(key, value)
			}

			config := NewDefaultConfig()

			assert.Equal(t, test.expected, config)
		})
	}
}

func TestDNSProvider_Present(t *testing.T) {
	mockResponses := MockResponseMap{
		"/2013-04-01/hostedzonesbyname":         {StatusCode: 200, Body: ListHostedZonesByNameResponse},
		"/2013-04-01/hostedzone/ABCDEFG/rrset/": {StatusCode: 200, Body: ChangeResourceRecordSetsResponse},
		"/2013-04-01/change/123456":             {StatusCode: 200, Body: GetChangeResponse},
		"/2013-04-01/hostedzone/ABCDEFG/rrset?name=_acme-challenge.example.com.&type=TXT": {
			StatusCode: 200,
			Body:       "",
		},
	}

	serverURL := setupTest(t, mockResponses)

	defer envTest.RestoreEnv()
	envTest.ClearEnv()
	provider := makeTestProvider(t, serverURL)

	domain := "example.com"
	keyAuth := "123456d=="

	err := provider.Present(domain, "", keyAuth)
	require.NoError(t, err, "Expected Present to return no error")
}

func TestCreateSession(t *testing.T) {
	testCases := []struct {
		desc             string
		env              map[string]string
		config           *Config
		wantCreds        credentials.Value
		wantDefaultChain bool
		wantRegion       string
		wantErr          string
	}{
		{
			desc:    "config is nil",
			wantErr: "config is nil",
		},
		{
			desc:    "session token without access key id or secret access key",
			config:  &Config{SessionToken: "foo"},
			wantErr: "SessionToken must be supplied with AccessKeyID and SecretAccessKey",
		},
		{
			desc:    "access key id without secret access key",
			config:  &Config{AccessKeyID: "foo"},
			wantErr: "AccessKeyID and SecretAccessKey must be supplied together",
		},
		{
			desc:    "access key id without secret access key",
			config:  &Config{SecretAccessKey: "foo"},
			wantErr: "AccessKeyID and SecretAccessKey must be supplied together",
		},
		{
			desc:             "credentials from default chain",
			config:           &Config{},
			wantDefaultChain: true,
		},
		{
			desc: "static credentials",
			config: &Config{
				AccessKeyID:     "one",
				SecretAccessKey: "two",
			},
			wantCreds: credentials.Value{
				AccessKeyID:     "one",
				SecretAccessKey: "two",
				SessionToken:    "",
				ProviderName:    credentials.StaticProviderName,
			},
		},
		{
			desc: "static credentials with session token",
			config: &Config{
				AccessKeyID:     "one",
				SecretAccessKey: "two",
				SessionToken:    "three",
			},
			wantCreds: credentials.Value{
				AccessKeyID:     "one",
				SecretAccessKey: "two",
				SessionToken:    "three",
				ProviderName:    credentials.StaticProviderName,
			},
		},
		{
			desc:   "region from env",
			config: &Config{},
			env: map[string]string{
				"AWS_REGION": "foo",
			},
			wantDefaultChain: true,
			wantRegion:       "foo",
		},
		{
			desc: "static region",
			config: &Config{
				Region: "one",
			},
			env: map[string]string{
				"AWS_REGION": "foo",
			},
			wantDefaultChain: true,
			wantRegion:       "one",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer envTest.RestoreEnv()
			envTest.ClearEnv()

			envTest.Apply(test.env)

			sess, err := createSession(test.config)
			requireErr(t, err, test.wantErr)

			if err != nil {
				return
			}

			gotCreds, err := sess.Config.Credentials.Get()

			if test.wantDefaultChain {
				assert.NotEqual(t, credentials.StaticProviderName, gotCreds.ProviderName)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.wantCreds, gotCreds)
			}

			if test.wantRegion != "" {
				assert.Equal(t, test.wantRegion, aws.StringValue(sess.Config.Region))
			}
		})
	}
}

func requireErr(t *testing.T, err error, wantErr string) {
	t.Helper()

	switch {
	case err != nil && wantErr == "":
		// force the assertion error.
		require.NoError(t, err)

	case err == nil && wantErr != "":
		// force the assertion error.
		require.EqualError(t, err, wantErr)

	case err != nil && wantErr != "":
		require.EqualError(t, err, wantErr)
	}
}
