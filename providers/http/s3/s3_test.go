// Package s3 implements a HTTP provider for solving the HTTP-01 challenge
// using AWS S3 in combination with AWS CloudFront.
package s3

import (
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/go-acme/lego/v4/challenge/http01"
	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	domain  = "example.com"
	token   = "foo"
	keyAuth = "bar"
)

var envTest = tester.NewEnvTest(
	"AWS_ACCESS_KEY_ID",
	"AWS_SECRET_ACCESS_KEY",
	"AWS_REGION",
	"S3_BUCKET")

func TestNewS3ProviderValid(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()

	_, err := NewHTTPProvider(envTest.GetValue("S3_BUCKET"))
	require.NoError(t, err)
}

func TestLiveS3ProviderPresent(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()

	s3Bucket := envTest.GetValue("S3_BUCKET")

	provider, err := NewHTTPProvider(s3Bucket)
	require.NoError(t, err)

	err = provider.Present(domain, token, keyAuth)
	require.NoError(t, err)

	// Need to wait a little bit before checking website
	time.Sleep(1 * time.Second)

	chlgPath := fmt.Sprintf("http://%s.s3.%s.amazonaws.com%s",
		s3Bucket, envTest.GetValue("AWS_REGION"), http01.ChallengePath(token))

	resp, err := http.Get(chlgPath)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(t, []byte(keyAuth), data)

	err = provider.CleanUp(domain, token, keyAuth)
	require.NoError(t, err)

	// Need to wait a little bit before checking website aghain
	time.Sleep(1 * time.Second)

	cleanupResp, err := http.Get(chlgPath)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, cleanupResp.StatusCode, 404)
}
