package lightsail

import (
	"context"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	envAwsNamespace = "AWS_"

	envAwsAccessKeyID     = envAwsNamespace + "ACCESS_KEY_ID"
	envAwsSecretAccessKey = envAwsNamespace + "SECRET_ACCESS_KEY"
	envAwsRegion          = envAwsNamespace + "REGION"
	envAwsHostedZoneID    = envAwsNamespace + "HOSTED_ZONE_ID"
)

var envTest = tester.NewEnvTest(
	envAwsAccessKeyID,
	envAwsSecretAccessKey,
	envAwsRegion,
	envAwsHostedZoneID).
	WithDomain(EnvDNSZone).
	WithLiveTestRequirements(envAwsAccessKeyID, envAwsSecretAccessKey, EnvDNSZone)

type endpointResolverMock struct {
	endpoint string
}

func (e endpointResolverMock) ResolveEndpoint(_, _ string, _ ...interface{}) (aws.Endpoint, error) {
	return aws.Endpoint{URL: e.endpoint}, nil
}

func makeProvider(serverURL string) *DNSProvider {
	config := aws.Config{
		Credentials:                 credentials.NewStaticCredentialsProvider("abc", "123", " "),
		Region:                      "mock-region",
		EndpointResolverWithOptions: endpointResolverMock{endpoint: serverURL},
		RetryMaxAttempts:            1,
	}

	return &DNSProvider{
		client: lightsail.NewFromConfig(config),
		config: NewDefaultConfig(),
	}
}

func TestCredentialsFromEnv(t *testing.T) {
	defer envTest.RestoreEnv()
	envTest.ClearEnv()

	_ = os.Setenv(envAwsAccessKeyID, "123")
	_ = os.Setenv(envAwsSecretAccessKey, "123")
	_ = os.Setenv(envAwsRegion, "us-east-1")

	ctx := context.Background()
	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	require.NoError(t, err)

	cs, err := cfg.Credentials.Retrieve(ctx)
	require.NoError(t, err, "Expected credentials to be set from environment")

	expected := aws.Credentials{
		AccessKeyID:     "123",
		SecretAccessKey: "123",
		Source:          "EnvConfigCredentials",
	}
	assert.Equal(t, expected, cs)
}

func TestDNSProvider_Present(t *testing.T) {
	mockResponses := map[string]MockResponse{
		"/": {StatusCode: 200, Body: ""},
	}

	serverURL := newMockServer(t, mockResponses)

	provider := makeProvider(serverURL)

	domain := "example.com"
	keyAuth := "123456d=="

	err := provider.Present(domain, "", keyAuth)
	require.NoError(t, err, "Expected Present to return no error")
}
