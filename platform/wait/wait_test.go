package wait

import (
	"errors"
	"testing"
	"time"
)

func TestFor_timeout(t *testing.T) {
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

func TestFor_stop(t *testing.T) {
	c := make(chan error)
	go func() {
		c <- For("", 3*time.Second, 1*time.Second, func() (bool, error) {
			return true, nil
		})
	}()

	timeout := time.After(6 * time.Second)
	select {
	case <-timeout:
		t.Fatal("timeout exceeded")
	case err := <-c:
		if err != nil {
			t.Errorf("expected no timeout error; got %v", err)
		}
	}
}

func TestFor_stop_error(t *testing.T) {
	c := make(chan error)
	go func() {
		c <- For("", 3*time.Second, 1*time.Second, func() (bool, error) {
			return true, errors.New("oops")
		})
	}()

	timeout := time.After(6 * time.Second)
	select {
	case <-timeout:
		t.Fatal("timeout exceeded")
	case err := <-c:
		if err == nil {
			t.Errorf("expected error; got %v", err)
		}
		t.Logf("%v", err)
	}
}
