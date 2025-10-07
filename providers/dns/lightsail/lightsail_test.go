package lightsail

import (
	"net/http/httptest"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/go-acme/lego/v4/platform/tester/servermock"
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

func TestCredentialsFromEnv(t *testing.T) {
	defer envTest.RestoreEnv()
	envTest.ClearEnv()

	_ = os.Setenv(envAwsAccessKeyID, "123")
	_ = os.Setenv(envAwsSecretAccessKey, "123")
	_ = os.Setenv(envAwsRegion, "us-east-1")

	ctx := t.Context()
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
	provider := servermock.NewBuilder(
		func(server *httptest.Server) (*DNSProvider, error) {
			return &DNSProvider{
				client: lightsail.NewFromConfig(aws.Config{
					HTTPClient:       server.Client(),
					Credentials:      credentials.NewStaticCredentialsProvider("abc", "123", " "),
					Region:           "mock-region",
					BaseEndpoint:     aws.String(server.URL),
					RetryMaxAttempts: 1,
				}),
				config: NewDefaultConfig(),
			}, nil
		}).
		Route("POST /", nil).
		Build(t)

	domain := "example.com"
	keyAuth := "123456d=="

	err := provider.Present(domain, "", keyAuth)
	require.NoError(t, err)
}
