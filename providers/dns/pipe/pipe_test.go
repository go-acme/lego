package pipe

import (
	"io/ioutil"
	"os"
	"path"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	liveTest bool
	pipePath string
	domain   string
)

func init() {
	pipePath = os.Getenv("PIPE_PATH")
	domain = os.Getenv("PIPE_TEST_DOMAIN")
	liveTest = len(pipePath) > 0 && len(domain) > 0
}

func restoreEnv() {
	os.Setenv("PIPE_PATH", pipePath)
}

func TestNewDNSProviderValidEnv(t *testing.T) {
	tmpdir, err := ioutil.TempDir("/tmp", "legotest")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpdir)
	fullPath := path.Join(tmpdir, "pipe")
	err = syscall.Mkfifo(fullPath, 0600)
	assert.NoError(t, err)
	os.Setenv("PIPE_PATH", fullPath)
	defer restoreEnv()
	_, err = NewDNSProvider()
	assert.NoError(t, err)
}

func TestNewDNSProviderMissingPipeErr(t *testing.T) {
	os.Setenv("PIPE_PATH", "")
	defer restoreEnv()
	_, err := NewDNSProvider()
	assert.EqualError(t, err, "open : no such file or directory")
}

func TestNewDNSProviderNonpipeErr(t *testing.T) {
	os.Setenv("EXEC_PATH", "/dev/null")
	defer restoreEnv()
	_, err := NewDNSProvider()
	assert.EqualError(t, err, "open : no such file or directory")
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
