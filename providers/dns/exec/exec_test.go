package exec

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	liveTest bool
	command  string
	domain   string
)

func init() {
	command = os.Getenv("EXEC_PATH")
	domain = os.Getenv("EXEC_TEST_DOMAIN")
	liveTest = len(command) > 0 && len(domain) > 0
}

func restoreEnv() {
	os.Setenv("EXEC_PATH", command)
}

func TestNewDNSProviderValidEnv(t *testing.T) {
	os.Setenv("EXEC_PATH", "true")
	defer restoreEnv()
	_, err := NewDNSProvider()
	assert.NoError(t, err)
}

func TestNewDNSProviderMissingCmdErr(t *testing.T) {
	os.Setenv("EXEC_PATH", "")
	defer restoreEnv()
	_, err := NewDNSProvider()
	assert.EqualError(t, err, "exec: \"\": executable file not found in $PATH")
}

func TestNewDNSProviderNonexecCmdErr(t *testing.T) {
	os.Setenv("EXEC_PATH", "/dev/null")
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
