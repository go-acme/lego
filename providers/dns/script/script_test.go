package script

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	liveTest bool
	script   string
	domain   string
)

func init() {
	script = os.Getenv("SCRIPT_PATH")
	domain = os.Getenv("SCRIPT_TEST_DOMAIN")
	liveTest = len(script) > 0 && len(domain) > 0
}

func restoreEnv() {
	os.Setenv("SCRIPT_PATH", script)
}

func TestNewDNSProviderValidEnv(t *testing.T) {
	os.Setenv("SCRIPT_PATH", "true")
	defer restoreEnv()
	_, err := NewDNSProvider()
	assert.NoError(t, err)
}

func TestNewDNSProviderMissingScriptErr(t *testing.T) {
	os.Setenv("SCRIPT_PATH", "")
	defer restoreEnv()
	_, err := NewDNSProvider()
	assert.EqualError(t, err, "exec: \"\": executable file not found in $PATH")
}

func TestNewDNSProviderNonexecScriptErr(t *testing.T) {
	os.Setenv("SCRIPT_PATH", "/dev/null")
	defer restoreEnv()
	_, err := NewDNSProvider()
	assert.EqualError(t, err, "exec: \"/dev/null\": permission denied")
}

func TestLivePresent(t *testing.T) {
	if !liveTest {
		t.Skip("skipping live test")
	}

	provider, err := NewDNSProvider()
	assert.NoError(t, err)

	err = provider.Present(domain, "", "123d==")
	assert.NoError(t, err)
}

func TestLiveCleanUp(t *testing.T) {
	if !liveTest {
		t.Skip("skipping live test")
	}

	time.Sleep(time.Second * 1)

	provider, err := NewDNSProvider()
	assert.NoError(t, err)

	err = provider.CleanUp(domain, "", "123d==")
	assert.NoError(t, err)
}
