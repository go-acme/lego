package dns

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	apiKey    string
	apiSecret string
)

func init() {
	apiSecret = os.Getenv("EXOSCALE_API_SECRET")
	apiKey = os.Getenv("EXOSCALE_API_KEY")
}

func restoreExoscaleEnv() {
	os.Setenv("EXOSCALE_API_KEY", apiKey)
	os.Setenv("EXOSCALE_API_SECRET", apiSecret)
}

func TestKnownDNSProviderSuccess(t *testing.T) {
	os.Setenv("EXOSCALE_API_KEY", "abc")
	os.Setenv("EXOSCALE_API_SECRET", "123")
	provider, err := NewDNSChallengeProviderByName("exoscale")
	assert.NoError(t, err)
	assert.NotNil(t, provider)
	restoreExoscaleEnv()
}

func TestKnownDNSProviderError(t *testing.T) {
	os.Setenv("EXOSCALE_API_KEY", "")
	os.Setenv("EXOSCALE_API_SECRET", "")
	_, err := NewDNSChallengeProviderByName("exoscale")
	assert.Error(t, err)
	restoreExoscaleEnv()
}

func TestUnknownDNSProvider(t *testing.T) {
	_, err := NewDNSChallengeProviderByName("foobar")
	assert.Error(t, err)
}
