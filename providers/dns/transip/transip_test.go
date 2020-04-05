package transip

import (
	"encoding/json"
	"fmt"
	"github.com/transip/gotransip/v6/rest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-acme/lego/v3/platform/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/transip/gotransip/v6/domain"
)

type fakeClient struct {
	dnsEntries           []domain.DNSEntry
	setDNSEntriesLatency time.Duration
	getInfoLatency       time.Duration
	domainName           string
}

type dnsEntryWrapper struct {
	DNSEntry domain.DNSEntry `json:"dnsEntry"`
}

type dnsEntriesWrapper struct {
	DNSEntries []domain.DNSEntry `json:"dnsEntries"`
}

func (f *fakeClient) Get(request rest.Request, dest interface{}) error {
	if f.getInfoLatency != 0 {
		time.Sleep(f.getInfoLatency)
	}

	switch request.Endpoint {
	case fmt.Sprintf("/domains/%s/dns", f.domainName):
		entries := dnsEntriesWrapper{DNSEntries: f.dnsEntries}
		body, err := json.Marshal(entries)

		if err != nil {
			return fmt.Errorf("can't encode json: %w", err)
		}

		err = json.Unmarshal(body, dest)

		if err != nil {
			return fmt.Errorf("can't decode json: %w", err)
		}
	default:
		return fmt.Errorf("function GET for endpoint %s not implemented", request.Endpoint)
	}

	return nil
}

func (f *fakeClient) Put(request rest.Request) error {
	if f.getInfoLatency != 0 {
		time.Sleep(f.getInfoLatency)
	}

	return fmt.Errorf("function PUT for endpoint %s not implemented", request.Endpoint)
}

func (f *fakeClient) Post(request rest.Request) error {
	if f.getInfoLatency != 0 {
		time.Sleep(f.getInfoLatency)
	}
	switch request.Endpoint {
	case fmt.Sprintf("/domains/%s/dns", f.domainName):
		body, err := request.GetJSONBody()
		if err != nil {
			return fmt.Errorf("unable get request body")
		}

		var entry dnsEntryWrapper
		if err := json.Unmarshal(body, &entry); err != nil {
			return fmt.Errorf("unable to decode request body")
		}

		f.dnsEntries = append(f.dnsEntries, entry.DNSEntry)
	default:
		return fmt.Errorf("function POST for endpoint %s not implemented", request.Endpoint)
	}

	return nil
}

func (f *fakeClient) Delete(request rest.Request) error {
	if f.getInfoLatency != 0 {
		time.Sleep(f.getInfoLatency)
	}

	switch request.Endpoint {
	case fmt.Sprintf("/domains/%s/dns", f.domainName):
		fmt.Println("removing dns entry")
		body, err := request.GetJSONBody()
		if err != nil {
			return fmt.Errorf("unable get request body")
		}

		var entry dnsEntryWrapper
		if err := json.Unmarshal(body, &entry); err != nil {
			return fmt.Errorf("unable to decode request body")
		}

		cp := f.dnsEntries

		for i, e := range f.dnsEntries {
			if e.Name == entry.DNSEntry.Name {
				fmt.Printf("found match %s\n", e.Name)
				cp = append(f.dnsEntries[:i], f.dnsEntries[i+1:]...)
			}
		}

		f.dnsEntries = cp
	default:
		return fmt.Errorf("function DELETE for endpoint %s not implemented", request.Endpoint)
	}

	return nil
}

func (f *fakeClient) Patch(request rest.Request) error {
	if f.getInfoLatency != 0 {
		time.Sleep(f.getInfoLatency)
	}

	return fmt.Errorf("function PATCH for endpoint %s not implemented", request.Endpoint)
}

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
		{
			desc: "could not open private key path",
			envVars: map[string]string{
				EnvAccountName:    "johndoe",
				EnvPrivateKeyPath: "./fixtures/non/existent/private.key",
			},
			expected: "transip: error while opening private key file: open ./fixtures/non/existent/private.key: The system cannot find the path specified.",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer envTest.RestoreEnv()
			envTest.ClearEnv()

			envTest.Apply(test.envVars)

			p, err := NewDNSProvider()

			if len(test.expected) == 0 {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.config)
				require.NotNil(t, p.repository)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
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
		{
			desc:           "could not open private key path",
			accountName:    "johndoe",
			privateKeyPath: "./fixtures/non/existent/private.key",
			expected:       "transip: error while opening private key file: open ./fixtures/non/existent/private.key: The system cannot find the path specified.",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.AccountName = test.accountName
			config.PrivateKeyPath = test.privateKeyPath

			p, err := NewDNSProviderConfig(config)

			if len(test.expected) == 0 {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.config)
				require.NotNil(t, p.repository)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestDNSProvider_concurrentGetInfo(t *testing.T) {
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

	solve := func(domain1 string, suffix string, timeoutPresent time.Duration, timeoutSolve time.Duration, timeoutCleanup time.Duration) error {
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

func TestDNSProvider_concurrentSetDNSEntries(t *testing.T) {
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

	solve := func(domain1 string, timeoutPresent time.Duration, timeoutCleanup time.Duration) error {
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
