package yandex

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	yandexLiveTest bool
	yandexEmail    string
	yandexAPIKey   string
	yandexDomain   string
)

func init() {
	yandexEmail = os.Getenv("YANDEX_EMAIL")
	yandexAPIKey = os.Getenv("YANDEX_API_KEY")
	yandexDomain = os.Getenv("YANDEX_DOMAIN")
	if len(yandexEmail) > 0 && len(yandexAPIKey) > 0 && len(yandexDomain) > 0 {
		yandexLiveTest = true
	}
}

func restoreYandexEnv() {
	os.Setenv("YANDEX_EMAIL", yandexEmail)
	os.Setenv("YANDEX_API_KEY", yandexAPIKey)
}

func TestNewDNSProviderValid(t *testing.T) {
	os.Setenv("YANDEX_EMAIL", "")
	os.Setenv("YANDEX_API_KEY", "")
	_, err := NewDNSProviderCredentials("123", "123")
	assert.NoError(t, err)
	restoreYandexEnv()
}

func TestNewDNSProviderValidEnv(t *testing.T) {
	os.Setenv("YANDEX_EMAIL", "test@example.com")
	os.Setenv("YANDEX_API_KEY", "123")
	_, err := NewDNSProvider()
	assert.NoError(t, err)
	restoreYandexEnv()
}

func TestNewDNSProviderMissingCredErr(t *testing.T) {
	os.Setenv("YANDEX_EMAIL", "")
	os.Setenv("YANDEX_API_KEY", "")
	_, err := NewDNSProvider()
	assert.EqualError(t, err, "Yandex credentials missing")
	restoreYandexEnv()
}

func TestYandexPresent(t *testing.T) {
	if !yandexLiveTest {
		t.Skip("skipping live test")
	}

	provider, err := NewDNSProviderCredentials(yandexEmail, yandexAPIKey)
	assert.NoError(t, err)

	err = provider.Present(yandexDomain, "", "123d==")
	assert.NoError(t, err)
}

func TestYandexCleanUp(t *testing.T) {
	if !yandexLiveTest {
		t.Skip("skipping live test")
	}

	time.Sleep(time.Second * 2)

	provider, err := NewDNSProviderCredentials(yandexEmail, yandexAPIKey)
	assert.NoError(t, err)

	err = provider.CleanUp(yandexDomain, "", "123d==")
	assert.NoError(t, err)
}
