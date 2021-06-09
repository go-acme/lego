package lightsail

import (
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/go-acme/lego/v4/platform/tester"
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

func makeProvider(serverURL string) (*DNSProvider, error) {
	config := &aws.Config{
		Credentials: credentials.NewStaticCredentials("abc", "123", " "),
		Endpoint:    aws.String(serverURL),
		Region:      aws.String("mock-region"),
		MaxRetries:  aws.Int(1),
	}

	sess, err := session.NewSession(config)
	if err != nil {
		return nil, err
	}

	conf := NewDefaultConfig()

	client := lightsail.New(sess)
	return &DNSProvider{client: client, config: conf}, nil
}

func TestCredentialsFromEnv(t *testing.T) {
	defer envTest.RestoreEnv()
	envTest.ClearEnv()

	_ = os.Setenv(envAwsAccessKeyID, "123")
	_ = os.Setenv(envAwsSecretAccessKey, "123")
	_ = os.Setenv(envAwsRegion, "us-east-1")

	config := &aws.Config{
		CredentialsChainVerboseErrors: aws.Bool(true),
	}

	sess, err := session.NewSession(config)
	require.NoError(t, err)

	_, err = sess.Config.Credentials.Get()
	require.NoError(t, err, "Expected credentials to be set from environment")
}

func TestDNSProvider_Present(t *testing.T) {
	mockResponses := map[string]MockResponse{
		"/": {StatusCode: 200, Body: ""},
	}

	serverURL := newMockServer(t, mockResponses)

	provider, err := makeProvider(serverURL)
	require.NoError(t, err)

	domain := "example.com"
	keyAuth := "123456d=="

	err = provider.Present(domain, "", keyAuth)
	require.NoError(t, err, "Expected Present to return no error")
}
