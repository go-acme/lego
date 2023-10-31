package s3

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

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

func TestLiveNewHTTPProvider_Valid(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()

	_, err := NewHTTPProvider(envTest.GetValue("S3_BUCKET"))
	require.NoError(t, err)
}

func TestLiveNewHTTPProvider(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()

	s3Bucket := os.Getenv("S3_BUCKET")

	provider, err := NewHTTPProvider(s3Bucket)
	require.NoError(t, err)

	// Present

	err = provider.Present(domain, token, keyAuth)
	require.NoError(t, err)

	chlgPath := fmt.Sprintf("http://%s.s3.%s.amazonaws.com%s",
		s3Bucket, envTest.GetValue("AWS_REGION"), http01.ChallengePath(token))

	resp, err := http.Get(chlgPath)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(t, []byte(keyAuth), data)

	// CleanUp

	err = provider.CleanUp(domain, token, keyAuth)
	require.NoError(t, err)

	cleanupResp, err := http.Get(chlgPath)
	require.NoError(t, err)

	assert.Equal(t, 403, cleanupResp.StatusCode)
}
