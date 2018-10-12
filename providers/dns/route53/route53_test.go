package route53

import (
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	envTestAwsSecretAccessKey string
	envTestAwsAccessKeyID     string
	envTestAwsRegion          string
	envTestAwsHostedZoneID    string

	envTestAwsMaxRetries         string
	envTestAwsTTL                string
	envTestAwsPropagationTimeout string
	envTestAwsPollingInterval    string
)

func init() {
	envTestAwsAccessKeyID = os.Getenv("AWS_ACCESS_KEY_ID")
	envTestAwsSecretAccessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	envTestAwsRegion = os.Getenv("AWS_REGION")
	envTestAwsHostedZoneID = os.Getenv("AWS_HOSTED_ZONE_ID")

	envTestAwsMaxRetries = os.Getenv("AWS_MAX_RETRIES")
	envTestAwsTTL = os.Getenv("AWS_TTL")
	envTestAwsPropagationTimeout = os.Getenv("AWS_PROPAGATION_TIMEOUT")
	envTestAwsPollingInterval = os.Getenv("AWS_POLLING_INTERVAL")
}

func restoreEnv() {
	os.Setenv("AWS_ACCESS_KEY_ID", envTestAwsAccessKeyID)
	os.Setenv("AWS_SECRET_ACCESS_KEY", envTestAwsSecretAccessKey)
	os.Setenv("AWS_REGION", envTestAwsRegion)
	os.Setenv("AWS_HOSTED_ZONE_ID", envTestAwsHostedZoneID)

	os.Setenv("AWS_MAX_RETRIES", envTestAwsMaxRetries)
	os.Setenv("AWS_TTL", envTestAwsTTL)
	os.Setenv("AWS_PROPAGATION_TIMEOUT", envTestAwsPropagationTimeout)
	os.Setenv("AWS_POLLING_INTERVAL", envTestAwsPollingInterval)
}

func cleanEnv() {
	os.Unsetenv("AWS_ACCESS_KEY_ID")
	os.Unsetenv("AWS_SECRET_ACCESS_KEY")
	os.Unsetenv("AWS_REGION")
	os.Unsetenv("AWS_HOSTED_ZONE_ID")

	os.Unsetenv("AWS_MAX_RETRIES")
	os.Unsetenv("AWS_TTL")
	os.Unsetenv("AWS_PROPAGATION_TIMEOUT")
	os.Unsetenv("AWS_POLLING_INTERVAL")
}

func makeTestProvider(ts *httptest.Server) *DNSProvider {
	config := &aws.Config{
		Credentials: credentials.NewStaticCredentials("abc", "123", " "),
		Endpoint:    aws.String(ts.URL),
		Region:      aws.String("mock-region"),
		MaxRetries:  aws.Int(1),
	}

	sess, err := session.NewSession(config)
	if err != nil {
		panic(err)
	}
	client := route53.New(sess)
	cfg := NewDefaultConfig()
	return &DNSProvider{client: client, config: cfg}
}

func Test_loadCredentials_FromEnv(t *testing.T) {
	defer restoreEnv()
	os.Setenv("AWS_ACCESS_KEY_ID", "123")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "456")
	os.Setenv("AWS_REGION", "us-east-1")

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
	defer restoreEnv()
	os.Setenv("AWS_REGION", route53.CloudWatchRegionUsEast1)

	sess, err := session.NewSession(aws.NewConfig())
	require.NoError(t, err)

	region := aws.StringValue(sess.Config.Region)
	assert.Equal(t, route53.CloudWatchRegionUsEast1, region, "Region")
}

func Test_getHostedZoneID_FromEnv(t *testing.T) {
	defer restoreEnv()

	expectedZoneID := "zoneID"

	os.Setenv("AWS_HOSTED_ZONE_ID", expectedZoneID)

	provider, err := NewDNSProvider()
	require.NoError(t, err)

	hostedZoneID, err := provider.getHostedZoneID("whatever")
	require.NoError(t, err, "HostedZoneID")

	assert.Equal(t, expectedZoneID, hostedZoneID)
}

func TestNewDefaultConfig(t *testing.T) {
	defer restoreEnv()

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
				"AWS_MAX_RETRIES":         "10",
				"AWS_TTL":                 "99",
				"AWS_PROPAGATION_TIMEOUT": "60",
				"AWS_POLLING_INTERVAL":    "60",
				"AWS_HOSTED_ZONE_ID":      "abc123",
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
			cleanEnv()
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

	ts := newMockServer(t, mockResponses)
	defer ts.Close()

	provider := makeTestProvider(ts)

	domain := "example.com"
	keyAuth := "123456d=="

	err := provider.Present(domain, "", keyAuth)
	require.NoError(t, err, "Expected Present to return no error")
}
