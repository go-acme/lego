package dreamhost

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	dreamHostLiveTest           bool
	dreamHostTestDomain         string
	dreamHostAPIKey             string
	fakeDreamHostAPIKey         = "asdf1234"
	fakeDreamHostChallengeToken = "foobar"
	fakeDreamHostKeyAuth        = "w6uP8Tcg6K2QR905Rms8iXTlksL6OD1KOWBxTK7wxPI"
)

func init() {
	dreamHostAPIKey = os.Getenv("DREAMHOST_API_KEY")
	dreamHostTestDomain = os.Getenv("DREAMHOST_TEST_DOMAIN")
	if len(dreamHostAPIKey) > 0 && len(dreamHostTestDomain) > 0 {
		dreamHostLiveTest = true
	}
}

func restoreEnv() {
	os.Setenv("DREAMHOST_API_KEY", dreamHostAPIKey)
}

func TestNewDNSProviderValidEnv(t *testing.T) {
	defer restoreEnv()
	os.Setenv("DREAMHOST_API_KEY", "somekey")

	_, err := NewDNSProvider()
	require.NoError(t, err)
}

func TestNewDNSProviderMissingCredErr(t *testing.T) {
	defer restoreEnv()
	os.Setenv("DREAMHOST_API_KEY", "")

	_, err := NewDNSProvider()
	assert.EqualError(t, err, "dreamhost: some credentials information are missing: DREAMHOST_API_KEY")
}
func TestDreamHostPresent(t *testing.T) {
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "method")

		q := r.URL.Query()

		assert.Equal(t, q.Get("key"), fakeDreamHostAPIKey, "key mismatch")
		assert.Equal(t, q.Get("cmd"), "dns-add_record", "cmd mismatch")
		assert.Equal(t, q.Get("record"), "_acme-challenge.example.com")
		assert.Equal(t, q.Get("value"), fakeDreamHostKeyAuth, "value mismatch")

		_, err := fmt.Fprintf(w, `{"data":"record_added","result":"success"}`)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}))
	defer mock.Close()

	config := NewDefaultConfig()
	config.APIKey = fakeDreamHostAPIKey
	config.BaseURL = mock.URL

	provider, err := NewDNSProviderConfig(config)
	require.NoError(t, err)
	require.NotNil(t, provider)

	err = provider.Present("example.com", "", fakeDreamHostChallengeToken)
	require.NoError(t, err, "fail to create TXT record")
}

func TestDreamHostPresentFailed(t *testing.T) {
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "method")

		_, err := fmt.Fprintf(w, `{"data":"record_already_exists_remove_first","result":"error"}`)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}))
	defer mock.Close()

	config := NewDefaultConfig()
	config.APIKey = fakeDreamHostAPIKey
	config.BaseURL = mock.URL

	provider, err := NewDNSProviderConfig(config)
	require.NoError(t, err)
	require.NotNil(t, provider)

	err = provider.Present("example.com", "", fakeDreamHostChallengeToken)
	require.Error(t, err, "API error not detected")
}

func TestDreamHostCleanup(t *testing.T) {
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "method")

		q := r.URL.Query()

		assert.Equal(t, q.Get("key"), fakeDreamHostAPIKey, "key mismatch")
		assert.Equal(t, q.Get("cmd"), "dns-remove_record", "cmd mismatch")
		assert.Equal(t, q.Get("record"), "_acme-challenge.example.com")
		assert.Equal(t, q.Get("value"), fakeDreamHostKeyAuth, "value mismatch")

		_, err := fmt.Fprintf(w, `{"data":"record_removed","result":"success"}`)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}))
	defer mock.Close()

	config := NewDefaultConfig()
	config.APIKey = fakeDreamHostAPIKey
	config.BaseURL = mock.URL

	provider, err := NewDNSProviderConfig(config)
	require.NoError(t, err)
	require.NotNil(t, provider)

	err = provider.CleanUp("example.com", "", fakeDreamHostChallengeToken)
	require.NoError(t, err, "failed to remove TXT record")
}

func TestLiveDreamHostPresentAndCleanUp(t *testing.T) {
	if !dreamHostLiveTest {
		t.Skip("skipping live test")
	}

	time.Sleep(time.Second * 1)

	config := NewDefaultConfig()
	config.APIKey = dreamHostAPIKey

	provider, err := NewDNSProviderConfig(config)
	require.NoError(t, err)
	err = provider.Present(dreamHostTestDomain, "", "123d==")
	require.NoError(t, err)

	time.Sleep(time.Second * 1)

	err = provider.CleanUp(dreamHostTestDomain, "", "123d==")
	require.NoError(t, err)
}
