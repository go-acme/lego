package wait

import (
	"testing"
	"time"
)

func TestForTimeout(t *testing.T) {
	c := make(chan error)
	go func() {
		err := For("", 3*time.Second, 1*time.Second, func() (bool, error) {
			return false, nil
		})
		c <- err
	}()

	timeout := time.After(4 * time.Second)
	select {
	case <-timeout:
		t.Fatal("timeout exceeded")
	case err := <-c:
		if err == nil {
			t.Errorf("expected timeout error; got %v", err)
		}
	}
}
