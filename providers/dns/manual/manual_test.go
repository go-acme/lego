package manual

import (
	"io"
	"os"
	"testing"
	"time"

	"github.com/go-acme/lego/v5/internal/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(EnvPropagationTimeout, EnvPollingInterval).WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	defer envTest.RestoreEnv()

	envTest.ClearEnv()

	envTest.Apply(map[string]string{
		EnvPropagationTimeout: "12",
		EnvPollingInterval:    "34",
	})

	p, err := NewDNSProvider()

	require.NoError(t, err)
	require.NotNil(t, p)
	require.NotNil(t, p.config)

	timeout, interval := p.Timeout()

	assert.Equal(t, 12*time.Second, timeout)
	assert.Equal(t, 34*time.Second, interval)
}

func TestNewDNSProviderConfig(t *testing.T) {
	config := NewDefaultConfig()
	config.PropagationTimeout = 12 * time.Second
	config.PollingInterval = 34 * time.Second

	p, err := NewDNSProviderConfig(config)

	require.NoError(t, err)
	require.NotNil(t, p)
	require.NotNil(t, p.config)

	timeout, interval := p.Timeout()

	assert.Equal(t, 12*time.Second, timeout)
	assert.Equal(t, 34*time.Second, interval)
}

func TestDNSProvider_manual(t *testing.T) {
	backupStdin := os.Stdin

	defer func() { os.Stdin = backupStdin }()

	testCases := []struct {
		desc        string
		input       string
		expectError bool
	}{
		{
			desc:  "Press enter",
			input: "ok\n",
		},
		{
			desc:        "Missing enter",
			input:       "ok",
			expectError: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			file, err := os.CreateTemp(t.TempDir(), "lego_test")
			require.NoError(t, err)

			t.Cleanup(func() { _ = file.Close() })

			_, err = file.WriteString(test.input)
			require.NoError(t, err)

			_, err = file.Seek(0, io.SeekStart)
			require.NoError(t, err)

			os.Stdin = file

			manualProvider, err := NewDNSProvider()
			require.NoError(t, err)

			err = manualProvider.Present(t.Context(), "example.com", "", "")
			if test.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				err = manualProvider.CleanUp(t.Context(), "example.com", "", "")
				require.NoError(t, err)
			}
		})
	}
}
