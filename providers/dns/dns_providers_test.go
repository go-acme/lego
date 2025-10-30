package dns

import (
	"testing"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/go-acme/lego/v4/providers/dns/exec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var envTest = tester.NewEnvTest("EXEC_PATH")

func TestKnownDNSProviderSuccess(t *testing.T) {
	defer envTest.RestoreEnv()

	envTest.Apply(map[string]string{
		"EXEC_PATH": "abc",
	})

	provider, err := NewDNSChallengeProviderByName("exec")
	require.NoError(t, err)
	assert.NotNil(t, provider)

	assert.IsType(t, &exec.DNSProvider{}, provider, "The loaded DNS provider doesn't have the expected type.")
}

func TestKnownDNSProviderError(t *testing.T) {
	defer envTest.RestoreEnv()

	envTest.ClearEnv()

	provider, err := NewDNSChallengeProviderByName("exec")
	require.Error(t, err)
	assert.Nil(t, provider)
}

func TestUnknownDNSProvider(t *testing.T) {
	provider, err := NewDNSChallengeProviderByName("foobar")
	require.Error(t, err)
	assert.Nil(t, provider)
}
