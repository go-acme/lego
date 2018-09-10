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

func makeRoute53Provider(ts *httptest.Server) *DNSProvider {
	config := &aws.Config{
		Credentials: credentials.NewStaticCredentials("abc", "123", " "),
		Endpoint:    aws.String(ts.URL),
		Region:      aws.String("mock-region"),
		MaxRetries:  aws.Int(1),
	}

	client := route53.New(session.New(config))
	cfg := NewDefaultConfig()
	return &DNSProvider{client: client, config: cfg}
}

func TestCredentialsFromEnv(t *testing.T) {
	defer restoreEnv()
	os.Setenv("AWS_ACCESS_KEY_ID", "123")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "123")
	os.Setenv("AWS_REGION", "us-east-1")

	config := &aws.Config{
		CredentialsChainVerboseErrors: aws.Bool(true),
	}

	sess := session.New(config)
	_, err := sess.Config.Credentials.Get()
	assert.NoError(t, err, "Expected credentials to be set from environment")
}

func TestRegionFromEnv(t *testing.T) {
	defer restoreEnv()
	os.Setenv("AWS_REGION", "us-east-1")

	sess := session.New(aws.NewConfig())
	assert.Equal(t, "us-east-1", aws.StringValue(sess.Config.Region), "Expected Region to be set from environment")
}

func TestHostedZoneIDFromEnv(t *testing.T) {
	defer restoreEnv()

	const testZoneID = "testzoneid"
	os.Setenv("AWS_HOSTED_ZONE_ID", testZoneID)

	provider, err := NewDNSProvider()
	assert.NoError(t, err, "Expected no error constructing DNSProvider")

	fqdn, err := provider.getHostedZoneID("whatever")
	assert.NoError(t, err, "Expected FQDN to be resolved to environment variable value")

	assert.Equal(t, testZoneID, fqdn)
}

func TestConfigFromEnv(t *testing.T) {
	defer restoreEnv()

	config := NewDefaultConfig()
	assert.Equal(t, config.TTL, 10, "Expected TTL to be use the default")

	os.Setenv("AWS_MAX_RETRIES", "10")
	os.Setenv("AWS_TTL", "99")
	os.Setenv("AWS_PROPAGATION_TIMEOUT", "60")
	os.Setenv("AWS_POLLING_INTERVAL", "60")
	const zoneID = "abc123"
	os.Setenv("AWS_HOSTED_ZONE_ID", zoneID)

	config = NewDefaultConfig()
	assert.Equal(t, config.MaxRetries, 10, "Expected PropagationTimeout to be configured from the environment")
	assert.Equal(t, config.TTL, 99, "Expected TTL to be configured from the environment")
	assert.Equal(t, config.PropagationTimeout, time.Second*60, "Expected PropagationTimeout to be configured from the environment")
	assert.Equal(t, config.PollingInterval, time.Second*60, "Expected PollingInterval to be configured from the environment")
	assert.Equal(t, config.HostedZoneID, zoneID, "Expected HostedZoneID to be configured from the environment")
}

func TestRoute53Present(t *testing.T) {
	mockResponses := MockResponseMap{
		"/2013-04-01/hostedzonesbyname":         MockResponse{StatusCode: 200, Body: ListHostedZonesByNameResponse},
		"/2013-04-01/hostedzone/ABCDEFG/rrset/": MockResponse{StatusCode: 200, Body: ChangeResourceRecordSetsResponse},
		"/2013-04-01/change/123456":             MockResponse{StatusCode: 200, Body: GetChangeResponse},
	}

	ts := newMockServer(t, mockResponses)
	defer ts.Close()

	provider := makeRoute53Provider(ts)

	domain := "example.com"
	keyAuth := "123456d=="

	err := provider.Present(domain, "", keyAuth)
	assert.NoError(t, err, "Expected Present to return no error")
}
