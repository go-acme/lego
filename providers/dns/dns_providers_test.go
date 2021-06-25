package dns

import (
	"encoding/json"
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
	assert.Error(t, err)
	assert.Nil(t, provider)
}

func TestGetSupportedProvider(t *testing.T) {
	assert.Equal(t, len(_str2provider), len(GetSupportedProvider()))
}

func TestMarshal(t *testing.T) {
	var p SupportedProvider

	e := json.Unmarshal([]byte(`"foobar"`), &p)
	require.Error(t, e)
	assert.ErrorAs(t, e, &ErrUnsupportedProvider{}, "unsupported provider")

	require.Error(t, json.Unmarshal([]byte(`"foo`), &p), "invalid json format")

	e = json.Unmarshal([]byte(`"exec"`), &p)
	require.NoError(t, e)
	assert.Equal(t, ProviderExec, p)

	out, e := json.Marshal(p)
	require.NoError(t, e)
	assert.Equal(t, `"exec"`, string(out))
}

func TestIsProviderSupported(t *testing.T) {
	assert.True(t, IsProviderSupporter("exec"))
	assert.False(t, IsProviderSupporter("foobar"))
}

func TestUnknownDNSProvider(t *testing.T) {
	provider, err := NewDNSChallengeProviderByName("foobar")
	assert.Error(t, err)
	assert.Nil(t, provider)
}
