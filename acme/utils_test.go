package acme

import (
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestWaitForTimeout(t *testing.T) {
	c := make(chan error)
	go func() {
		err := WaitFor(3*time.Second, 1*time.Second, func() (bool, error) {
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

func TestGetenvReadsEnvVars(t *testing.T) {
	os.Setenv("MY_SILLY_ENV_VAR", "bacon")
	readValue := GetenvOrFile("MY_SILLY_ENV_VAR")

	if readValue != "bacon" {
		t.Fatal("Expected bacon, got: ", readValue)
	}
}

func TestGetenvReadsFiles(t *testing.T) {
	os.Setenv("MY_SILLY_ENV_VAR_FILE", "/tmp/bacon.env.test")
	ioutil.WriteFile("/tmp/bacon.env.test", []byte("bacon"), 0644)

	readValue := GetenvOrFile("MY_SILLY_ENV_VAR")

	if readValue != "bacon" {
		t.Fatal("Expected bacon, got: ", readValue)
	}
}

func TestGetenvPrefersEnvVars(t *testing.T) {
	os.Setenv("MY_SILLY_ENV_VAR", "bacon1")
	os.Setenv("MY_SILLY_ENV_VAR_FILE", "/tmp/bacon.env.test")
	ioutil.WriteFile("/tmp/bacon.env.test", []byte("bacon2"), 0644)

	readValue := GetenvOrFile("MY_SILLY_ENV_VAR")

	if readValue != "bacon1" {
		t.Fatal("Expected bacon1, got: ", readValue)
	}
}
