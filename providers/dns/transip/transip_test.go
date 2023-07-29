package transip

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/transip/gotransip/v6/domain"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(
	EnvAccountName,
	EnvPrivateKeyPath).
	WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvAccountName:    "johndoe",
				EnvPrivateKeyPath: "./fixtures/private.key",
			},
		},
		{
			desc: "missing all credentials",
			envVars: map[string]string{
				EnvAccountName:    "",
				EnvPrivateKeyPath: "",
			},
			expected: "transip: some credentials information are missing: TRANSIP_ACCOUNT_NAME,TRANSIP_PRIVATE_KEY_PATH",
		},
		{
			desc: "missing account name",
			envVars: map[string]string{
				EnvAccountName:    "",
				EnvPrivateKeyPath: "./fixtures/private.key",
			},
			expected: "transip: some credentials information are missing: TRANSIP_ACCOUNT_NAME",
		},
		{
			desc: "missing private key path",
			envVars: map[string]string{
				EnvAccountName:    "johndoe",
				EnvPrivateKeyPath: "",
			},
			expected: "transip: some credentials information are missing: TRANSIP_PRIVATE_KEY_PATH",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer envTest.RestoreEnv()
			envTest.ClearEnv()

			envTest.Apply(test.envVars)

			p, err := NewDNSProvider()

			if test.expected == "" {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.config)
				require.NotNil(t, p.repository)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}

	// The error message for a file not existing is different on Windows and Linux.
	// Therefore, we test if the error type is the same.
	t.Run("could not open private key path", func(t *testing.T) {
		defer envTest.RestoreEnv()
		envTest.ClearEnv()

		envTest.Apply(map[string]string{
			EnvAccountName:    "johndoe",
			EnvPrivateKeyPath: "./fixtures/non/existent/private.key",
		})

		_, err := NewDNSProvider()
		assert.ErrorIs(t, err, os.ErrNotExist)
	})
}

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc           string
		accountName    string
		privateKeyPath string
		expected       string
	}{
		{
			desc:           "success",
			accountName:    "johndoe",
			privateKeyPath: "./fixtures/private.key",
		},
		{
			desc:     "missing all credentials",
			expected: "transip: AccountName is required",
		},
		{
			desc:           "missing account name",
			privateKeyPath: "./fixtures/private.key",
			expected:       "transip: AccountName is required",
		},
		{
			desc:        "missing private key path",
			accountName: "johndoe",
			expected:    "transip: PrivateKeyReader, token or PrivateKeyReader is required",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.AccountName = test.accountName
			config.PrivateKeyPath = test.privateKeyPath

			p, err := NewDNSProviderConfig(config)

			if test.expected == "" {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.config)
				require.NotNil(t, p.repository)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}

	// The error message for a file not existing is different on Windows and Linux.
	// Therefore, we test if the error type is the same.
	t.Run("could not open private key path", func(t *testing.T) {
		config := NewDefaultConfig()
		config.AccountName = "johndoe"
		config.PrivateKeyPath = "./fixtures/non/existent/private.key"

		_, err := NewDNSProviderConfig(config)
		assert.ErrorIs(t, err, os.ErrNotExist)
	})
}

func TestDNSProvider_concurrentGetDNSEntries(t *testing.T) {
	client := &fakeClient{
		getInfoLatency:       50 * time.Millisecond,
		setDNSEntriesLatency: 500 * time.Millisecond,
		domainName:           "lego.wtf",
	}

	repo := domain.Repository{Client: client}

	p := &DNSProvider{
		config:     NewDefaultConfig(),
		repository: repo,
	}

	var wg sync.WaitGroup
	wg.Add(2)

	solve := func(domain1, suffix string, timeoutPresent, timeoutSolve, timeoutCleanup time.Duration) error {
		time.Sleep(timeoutPresent)

		err := p.Present(domain1, "", "")
		if err != nil {
			return err
		}

		time.Sleep(timeoutSolve)

		var found bool
		for _, entry := range client.dnsEntries {
			if strings.HasSuffix(entry.Name, suffix) {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("record %s not found: %v", suffix, client.dnsEntries)
		}

		time.Sleep(timeoutCleanup)

		return p.CleanUp(domain1, "", "")
	}

	go func() {
		defer wg.Done()
		err := solve("bar.lego.wtf", ".bar", 500*time.Millisecond, 100*time.Millisecond, 100*time.Millisecond)
		require.NoError(t, err)
	}()

	go func() {
		defer wg.Done()
		err := solve("foo.lego.wtf", ".foo", 500*time.Millisecond, 200*time.Millisecond, 100*time.Millisecond)
		require.NoError(t, err)
	}()

	wg.Wait()

	assert.Empty(t, client.dnsEntries)
}

func TestDNSProvider_concurrentAddDNSEntry(t *testing.T) {
	client := &fakeClient{
		domainName: "lego.wtf",
	}
	repo := domain.Repository{Client: client}

	p := &DNSProvider{
		config:     NewDefaultConfig(),
		repository: repo,
	}

	var wg sync.WaitGroup
	wg.Add(2)

	solve := func(domain1 string, timeoutPresent, timeoutCleanup time.Duration) error {
		time.Sleep(timeoutPresent)
		err := p.Present(domain1, "", "")
		if err != nil {
			return err
		}

		time.Sleep(timeoutCleanup)
		return p.CleanUp(domain1, "", "")
	}

	go func() {
		defer wg.Done()
		err := solve("bar.lego.wtf", 550*time.Millisecond, 500*time.Millisecond)
		require.NoError(t, err)
	}()

	go func() {
		defer wg.Done()
		err := solve("foo.lego.wtf", 500*time.Millisecond, 100*time.Millisecond)
		require.NoError(t, err)
	}()

	wg.Wait()

	assert.Empty(t, client.dnsEntries)
}

func TestLivePresent(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()
	provider, err := NewDNSProvider()
	require.NoError(t, err)

	err = provider.Present(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}

func TestLiveCleanUp(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()
	provider, err := NewDNSProvider()
	require.NoError(t, err)

	time.Sleep(1 * time.Second)

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}
