package wait

import (
	"testing"
	"time"
)

func TestForTimeout(t *testing.T) {
	c := make(chan error)
	go func() {
		c <- For("", 3*time.Second, 1*time.Second, func() (bool, error) {
			return false, nil
		})
	}()

	timeout := time.After(6 * time.Second)
	select {
	case <-timeout:
		t.Fatal("timeout exceeded")
	case err := <-c:
		if err == nil {
			t.Errorf("expected timeout error; got %v", err)
		}
		t.Logf("%v", err)
	}
}
