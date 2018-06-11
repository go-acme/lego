package lightsail

import (
	"net/http/httptest"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/stretchr/testify/assert"
)

var (
	lightsailSecret string
	lightsailKey    string
	lightsailZone   string
)

func init() {
	lightsailKey = os.Getenv("AWS_ACCESS_KEY_ID")
	lightsailSecret = os.Getenv("AWS_SECRET_ACCESS_KEY")
}

func restoreEnv() {
	os.Setenv("AWS_ACCESS_KEY_ID", lightsailKey)
	os.Setenv("AWS_SECRET_ACCESS_KEY", lightsailSecret)
	os.Setenv("AWS_REGION", "us-east-1")
	os.Setenv("AWS_HOSTED_ZONE_ID", lightsailZone)
}

func makeLightsailProvider(ts *httptest.Server) *DNSProvider {
	config := &aws.Config{
		Credentials: credentials.NewStaticCredentials("abc", "123", " "),
		Endpoint:    aws.String(ts.URL),
		Region:      aws.String("mock-region"),
		MaxRetries:  aws.Int(1),
	}

	client := lightsail.New(session.New(config))
	return &DNSProvider{client: client}
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

func TestLightsailPresent(t *testing.T) {
	mockResponses := MockResponseMap{
		"/": MockResponse{StatusCode: 200, Body: ""},
	}

	ts := newMockServer(t, mockResponses)
	defer ts.Close()

	provider := makeLightsailProvider(ts)

	domain := "example.com"
	keyAuth := "123456d=="

	err := provider.Present(domain, "", keyAuth)
	assert.NoError(t, err, "Expected Present to return no error")
}
