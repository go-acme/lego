// Package s3 implements a HTTP provider for solving the HTTP-01 challenge
// using AWS S3 in combination with AWS CloudFront.
package s3

import (
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/go-acme/lego/v4/challenge/http01"
	"github.com/stretchr/testify/assert"
)

const (
	domain  = "domain"
	token   = "token"
	keyAuth = "keyAuth"
)

func isLive() bool {
	s3Key := os.Getenv("AWS_ACCESS_KEY_ID")
	s3Secret := os.Getenv("AWS_SECRET_ACCESS_KEY")
	s3Region := os.Getenv("AWS_REGION")
	s3Bucket := os.Getenv("S3_BUCKET")
	if len(s3Key) > 0 && len(s3Secret) > 0 && len(s3Region) > 0 && len(s3Bucket) > 0 {
		return true
	}

	return false
}

func TestNewS3ProviderValid(t *testing.T) {
	s3Bucket := os.Getenv("S3_BUCKET")
	if !isLive() {
		t.Skip("skipping live test")
	}

	_, err := NewHTTPProvider(s3Bucket)
	assert.NoError(t, err)
}

func TestLiveS3ProviderPresent(t *testing.T) {
	s3Region := os.Getenv("AWS_REGION")
	s3Bucket := os.Getenv("S3_BUCKET")
	if !isLive() {
		t.Skip("skipping live test")
	}

	provider, err := NewHTTPProvider(s3Bucket)
	assert.NoError(t, err)

	err = provider.Present(domain, token, keyAuth)
	assert.NoError(t, err)
	// Need to wait a little bit before checking website
	time.Sleep(1 * time.Second)
	s3Host := "http://" + s3Bucket + ".s3-website." + s3Region + ".amazonaws.com"
	resp, err := http.Get(s3Host + http01.ChallengePath(token))
	assert.NoError(t, err)
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, data, []byte(keyAuth))
}
