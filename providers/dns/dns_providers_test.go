package dns

import (
	"runtime/debug"
	"slices"
	"strings"
	"testing"

	"github.com/go-acme/lego/v5/internal/tester"
	"github.com/go-acme/lego/v5/providers/dns/exec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	buildTagDNSLight  = "lego_dnslight"
	buildTagDNSPrefix = "lego_dns_"
	buildTagDNSExec   = "lego_dns_exec"
)

var envTest = tester.NewEnvTest("EXEC_PATH")

func TestKnownDNSProviderSuccess(t *testing.T) {
	if !isExecProviderAvailable() {
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
	if !isExecProviderAvailable() {
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

func isExecProviderAvailable() bool {
	tags := getBuildTags()

	return !slices.Contains(tags, buildTagDNSLight) || slices.Contains(tags, buildTagDNSExec)
}

func getBuildTags() []string {
	buildInfo, available := debug.ReadBuildInfo()
	if !available {
		return nil
	}

	for _, setting := range buildInfo.Settings {
		if setting.Key == "-tags" {
			return strings.Split(setting.Value, ",")
		}
	}

	return nil
}
