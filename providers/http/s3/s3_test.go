package s3

import (
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xenolf/lego/acme"
)

var (
	s3Key      string
	s3Secret   string
	s3Region   string
	s3Bucket   string
	s3LiveTest bool
)

const (
	domain  = "domain"
	token   = "token"
	keyAuth = "keyAuth"
)

func init() {
	s3Key = os.Getenv("AWS_ACCESS_KEY_ID")
	s3Secret = os.Getenv("AWS_SECRET_ACCESS_KEY")
	s3Region = os.Getenv("AWS_REGION")
	s3Bucket = os.Getenv("S3_BUCKET")
	if len(s3Key) > 0 && len(s3Secret) > 0 && len(s3Region) > 0 && len(s3Bucket) > 0 {
		s3LiveTest = true
	}
}

func TestNewS3ProviderValid(t *testing.T) {
	if !s3LiveTest {
		t.Skip("skipping live test")
	}

	_, err := NewHTTPProvider(s3Bucket, s3Region)
	assert.NoError(t, err)
}

func TestLiveS3ProviderPresent(t *testing.T) {
	if !s3LiveTest {
		t.Skip("skipping live test")
	}

	provider, err := NewHTTPProvider(s3Bucket, s3Region)
	assert.NoError(t, err)

	err = provider.Present(domain, token, keyAuth)
	assert.NoError(t, err)

	s3Host := "https://" + s3Bucket + ".s3.amazonaws.com"
	resp, err := http.Get(s3Host + acme.HTTP01ChallengePath(token))
	assert.NoError(t, err)
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, data, []byte(keyAuth))
}

func TestLiveS3ProviderCleanUp(t *testing.T) {
	if !s3LiveTest {
		t.Skip("skipping live test")
	}

	provider, err := NewHTTPProvider(s3Bucket, s3Region)
	assert.NoError(t, err)

	err = provider.CleanUp(domain, token, keyAuth)
	assert.NoError(t, err)
}
