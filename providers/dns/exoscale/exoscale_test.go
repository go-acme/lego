package exoscale

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	exoscaleLiveTest  bool
	exoscaleAPIKey    string
	exoscaleAPISecret string
	exoscaleDomain    string
)

func init() {
	exoscaleAPISecret = os.Getenv("EXOSCALE_API_SECRET")
	exoscaleAPIKey = os.Getenv("EXOSCALE_API_KEY")
	exoscaleDomain = os.Getenv("EXOSCALE_DOMAIN")
	if len(exoscaleAPIKey) > 0 && len(exoscaleAPISecret) > 0 && len(exoscaleDomain) > 0 {
		exoscaleLiveTest = true
	}
}

func restoreEnv() {
	os.Setenv("EXOSCALE_API_KEY", exoscaleAPIKey)
	os.Setenv("EXOSCALE_API_SECRET", exoscaleAPISecret)
}

func TestNewDNSProviderValid(t *testing.T) {
	defer restoreEnv()
	os.Setenv("EXOSCALE_API_KEY", "")
	os.Setenv("EXOSCALE_API_SECRET", "")

	config := NewDefaultConfig()
	config.APIKey = "example@example.com"
	config.APISecret = "123"

	_, err := NewDNSProviderConfig(config)
	assert.NoError(t, err)
}

func TestNewDNSProviderValidEnv(t *testing.T) {
	defer restoreEnv()
	os.Setenv("EXOSCALE_API_KEY", "example@example.com")
	os.Setenv("EXOSCALE_API_SECRET", "123")

	_, err := NewDNSProvider()
	assert.NoError(t, err)
}

func TestNewDNSProviderMissingCredErr(t *testing.T) {
	os.Setenv("EXOSCALE_API_KEY", "")
	os.Setenv("EXOSCALE_API_SECRET", "")
	defer restoreEnv()

	_, err := NewDNSProvider()
	assert.EqualError(t, err, "exoscale: some credentials information are missing: EXOSCALE_API_KEY,EXOSCALE_API_SECRET")
}

func TestExtractRootRecordName(t *testing.T) {
	config := NewDefaultConfig()
	config.APIKey = "example@example.com"
	config.APISecret = "123"

	provider, err := NewDNSProviderConfig(config)
	assert.NoError(t, err)

	zone, recordName, err := provider.FindZoneAndRecordName("_acme-challenge.bar.com.", "bar.com")
	assert.NoError(t, err)
	assert.Equal(t, "bar.com", zone)
	assert.Equal(t, "_acme-challenge", recordName)
}

func TestExtractSubRecordName(t *testing.T) {
	config := NewDefaultConfig()
	config.APIKey = "example@example.com"
	config.APISecret = "123"

	provider, err := NewDNSProviderConfig(config)
	assert.NoError(t, err)

	zone, recordName, err := provider.FindZoneAndRecordName("_acme-challenge.foo.bar.com.", "foo.bar.com")
	assert.NoError(t, err)
	assert.Equal(t, "bar.com", zone)
	assert.Equal(t, "_acme-challenge.foo", recordName)
}

func TestLiveExoscalePresent(t *testing.T) {
	if !exoscaleLiveTest {
		t.Skip("skipping live test")
	}

	config := NewDefaultConfig()
	config.APIKey = exoscaleAPIKey
	config.APISecret = exoscaleAPISecret

	provider, err := NewDNSProviderConfig(config)
	assert.NoError(t, err)

	err = provider.Present(exoscaleDomain, "", "123d==")
	assert.NoError(t, err)

	// Present Twice to handle create / update
	err = provider.Present(exoscaleDomain, "", "123d==")
	assert.NoError(t, err)
}

func TestLiveExoscaleCleanUp(t *testing.T) {
	if !exoscaleLiveTest {
		t.Skip("skipping live test")
	}

	time.Sleep(time.Second * 1)

	config := NewDefaultConfig()
	config.APIKey = exoscaleAPIKey
	config.APISecret = exoscaleAPISecret

	provider, err := NewDNSProviderConfig(config)
	assert.NoError(t, err)

	err = provider.CleanUp(exoscaleDomain, "", "123d==")
	assert.NoError(t, err)
}
