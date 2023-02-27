package nfqueue

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNonLinux(t *testing.T) {
	// this test doesn't apply for linux
	if runtime.GOOS == "linux" {
		return
	}
	serv, _ := NewHttpDpiProvider("3331")
	err := serv.Present("exemple.org", "somerandomstring", "otherrandomstring")
	// just test if error mentions linux here
	assert.Contains(t, err.Error(), "linux")
}
