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
	r53AwsSecretAccessKey string
	r53AwsAccessKeyID     string
	r53AwsRegion          string
	r53AwsHostedZoneID    string

	r53AwsMaxRetries         string
	r53AwsTTL                string
	r53AwsPropagationTimeout string
	r53AwsPollingInterval    string
)

func init() {
	r53AwsAccessKeyID = os.Getenv("AWS_ACCESS_KEY_ID")
	r53AwsSecretAccessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	r53AwsRegion = os.Getenv("AWS_REGION")
	r53AwsHostedZoneID = os.Getenv("AWS_HOSTED_ZONE_ID")

	r53AwsMaxRetries = os.Getenv("AWS_MAX_RETRIES")
	r53AwsTTL = os.Getenv("AWS_TTL")
	r53AwsPropagationTimeout = os.Getenv("AWS_PROPAGATION_TIMEOUT")
	r53AwsPollingInterval = os.Getenv("AWS_POLLING_INTERVAL")
}

func restoreEnv() {
	os.Setenv("AWS_ACCESS_KEY_ID", r53AwsAccessKeyID)
	os.Setenv("AWS_SECRET_ACCESS_KEY", r53AwsSecretAccessKey)
	os.Setenv("AWS_REGION", r53AwsRegion)
	os.Setenv("AWS_HOSTED_ZONE_ID", r53AwsHostedZoneID)

	os.Setenv("AWS_MAX_RETRIES", r53AwsMaxRetries)
	os.Setenv("AWS_TTL", r53AwsTTL)
	os.Setenv("AWS_PROPAGATION_TIMEOUT", r53AwsPropagationTimeout)
	os.Setenv("AWS_POLLING_INTERVAL", r53AwsPollingInterval)
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

func makeRoute53Provider(ts *httptest.Server) *DNSProvider {
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
	assert.NoError(t, err, "Expected credentials to be set from environment")

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
	assert.NoError(t, err)

	hostedZoneID, err := provider.getHostedZoneID("whatever")
	assert.NoError(t, err, "HostedZoneID")

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

func TestRoute53Present(t *testing.T) {
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

	provider := makeRoute53Provider(ts)

	domain := "example.com"
	keyAuth := "123456d=="

	err := provider.Present(domain, "", keyAuth)
	assert.NoError(t, err, "Expected Present to return no error")
}
