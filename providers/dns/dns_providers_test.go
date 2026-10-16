package dns

import (
	"runtime/debug"
	"strings"
	"testing"

	"github.com/go-acme/lego/v5/internal/tester"
	"github.com/go-acme/lego/v5/providers/dns/exec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var envTest = tester.NewEnvTest("EXEC_PATH")

func TestKnownDNSProviderSuccess(t *testing.T) {
	if !hasExecProvider() {
		t.Skip("the exec provider is not available.")
	}

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
	if !hasExecProvider() {
		t.Skip("the exec provider is not available.")
	}

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

func hasExecProvider() bool {
	buildInfo, available := debug.ReadBuildInfo()
	if !available {
		return true
	}

	for _, setting := range buildInfo.Settings {
		if setting.Key == "-tags" {
			if strings.Contains(setting.Value, "lego_dnslight") {
				return strings.Contains(setting.Value, "lego_dns_exec")
			}

			return true
		}
	}

	return true
}
