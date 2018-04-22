package acme

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/xenolf/lego/log"
)

// WaitFor polls the given function 'f', once every 'interval', up to 'timeout'.
func WaitFor(timeout, interval time.Duration, f func() (bool, error)) error {
	log.Infof("Wait [timeout: %s, interval: %s]", timeout, interval)

	var lastErr string
	timeup := time.After(timeout)
	for {
		select {
		case <-timeup:
			return fmt.Errorf("time limit exceeded: last error: %s", lastErr)
		default:
		}

		stop, err := f()
		if stop {
			return nil
		}
		if err != nil {
			lastErr = err.Error()
		}

		time.Sleep(interval)
	}
}

// Attempts to resolve 'key' as an environment variable. Failing that, it will
// check to see if '$key_FILE' exists. If so, it will attempt to read from the
// referenced file to populate a value.
func GetenvOrFile(envVar string) string {
	envVarValue := os.Getenv(envVar)

	if envVarValue != "" {
		return envVarValue
	}

	fileVar := envVar + "_FILE"
	fileVarValue := os.Getenv(fileVar)

	if fileVarValue == "" {
		return envVarValue
	}

	fileContents, err := ioutil.ReadFile(fileVarValue)

	if err != nil {
		fmt.Printf("Error reading the file %s (defined by env var %s): %s\n", fileVarValue, fileVar, err)
		return ""
	}

	return string(fileContents)
}
