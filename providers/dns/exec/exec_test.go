package exec

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/xenolf/lego/challenge/dns01"
	"github.com/xenolf/lego/log"
)

func TestDNSProvider_Present(t *testing.T) {
	backupLogger := log.Logger
	defer func() {
		log.Logger = backupLogger
	}()

	logRecorder := &LogRecorder{}
	log.Logger = logRecorder

	type expected struct {
		args  string
		error bool
	}

	testCases := []struct {
		desc     string
		config   *Config
		expected expected
	}{
		{
			desc: "Standard mode",
			config: &Config{
				Program: "echo",
				Mode:    "",
			},
			expected: expected{
				args: "present _acme-challenge.domain. pW9ZKG0xz_PCriK-nCMOjADy9eJcgGWIzkkj2fN4uZM\n",
			},
		},
		{
			desc: "program error",
			config: &Config{
				Program: "ogellego",
				Mode:    "",
			},
			expected: expected{error: true},
		},
		{
			desc: "Raw mode",
			config: &Config{
				Program: "echo",
				Mode:    "RAW",
			},
			expected: expected{
				args: "present -- domain token keyAuth\n",
			},
		},
	}

	var message string
	logRecorder.On("Println", mock.Anything).Run(func(args mock.Arguments) {
		message = args.String(0)
		fmt.Fprintln(os.Stdout, "XXX", message)
	})

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			message = ""

			provider, err := NewDNSProviderConfig(test.config)
			require.NoError(t, err)

			err = provider.Present("domain", "token", "keyAuth")
			if test.expected.error {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expected.args, message)
			}
		})
	}
}

func TestDNSProvider_CleanUp(t *testing.T) {
	backupLogger := log.Logger
	defer func() {
		log.Logger = backupLogger
	}()

	logRecorder := &LogRecorder{}
	log.Logger = logRecorder

	type expected struct {
		args  string
		error bool
	}

	testCases := []struct {
		desc     string
		config   *Config
		expected expected
	}{
		{
			desc: "Standard mode",
			config: &Config{
				Program: "echo",
				Mode:    "",
			},
			expected: expected{
				args: "cleanup _acme-challenge.domain. pW9ZKG0xz_PCriK-nCMOjADy9eJcgGWIzkkj2fN4uZM\n",
			},
		},
		{
			desc: "program error",
			config: &Config{
				Program: "ogellego",
				Mode:    "",
			},
			expected: expected{error: true},
		},
		{
			desc: "Raw mode",
			config: &Config{
				Program: "echo",
				Mode:    "RAW",
			},
			expected: expected{
				args: "cleanup -- domain token keyAuth\n",
			},
		},
	}

	var message string
	logRecorder.On("Println", mock.Anything).Run(func(args mock.Arguments) {
		message = args.String(0)
		fmt.Fprintln(os.Stdout, "XXX", message)
	})

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			message = ""

			provider, err := NewDNSProviderConfig(test.config)
			require.NoError(t, err)

			err = provider.CleanUp("domain", "token", "keyAuth")
			if test.expected.error {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expected.args, message)
			}
		})
	}
}

func TestDNSProvider_Timeout(t *testing.T) {
	type expected struct {
		Timeout, Interval time.Duration
	}

	testCases := []struct {
		desc     string
		config   *Config
		expected expected
	}{
		{
			desc: "Standard mode",
			config: &Config{
				Program: "./fixtures/timeout.sh",
				Mode:    "",
			},
			expected: expected{
				Timeout:  30 * time.Second,
				Interval: 5 * time.Second,
			},
		},
		{
			desc: "Raw mode",
			config: &Config{
				Program: "./fixtures/timeout.sh",
				Mode:    "RAW",
			},
			expected: expected{
				Timeout:  30 * time.Second,
				Interval: 5 * time.Second,
			},
		},
		{
			desc: "program error",
			config: &Config{
				Program: "ogellego",
				Mode:    "",
			},
			expected: expected{
				Timeout:  dns01.DefaultPropagationTimeout,
				Interval: dns01.DefaultPollingInterval,
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			// message = ""

			provider, err := NewDNSProviderConfig(test.config)
			require.NoError(t, err)

			timeout, interval := provider.Timeout()

			assert.Equal(t, test.expected.Timeout, timeout)
			assert.Equal(t, test.expected.Interval, interval)
		})
	}
}
