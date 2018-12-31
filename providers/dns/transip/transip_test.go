package transip

import (
	"encoding/xml"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/transip/gotransip"
	"github.com/transip/gotransip/domain"
	"github.com/xenolf/lego/log"
	"github.com/xenolf/lego/platform/tester"
)

type argDNSEntries struct {
	Items domain.DNSEntries `xml:"item"`
}

type argDomainName struct {
	DomainName string `xml:",chardata"`
}

type fakeClient struct {
	dnsEntries           []domain.DNSEntry
	setDNSEntriesLatency time.Duration
	getInfoLatency       time.Duration
}

func (f *fakeClient) Call(r gotransip.SoapRequest, b interface{}) error {
	switch r.Method {
	case "getInfo":
		d := b.(*domain.Domain)
		cp := f.dnsEntries

		if f.getInfoLatency != 0 {
			time.Sleep(f.getInfoLatency)
		}
		d.DNSEntries = cp

		log.Printf("getInfo: %+v\n", d.DNSEntries)
		return nil
	case "setDnsEntries":
		var domainName argDomainName
		var dnsEntries argDNSEntries

		args := readArgs(r)
		for _, arg := range args {
			if strings.HasPrefix(arg, "<domainName") {
				err := xml.Unmarshal([]byte(arg), &domainName)
				if err != nil {
					panic(err)
				}
			} else if strings.HasPrefix(arg, "<dnsEntries") {
				err := xml.Unmarshal([]byte(arg), &dnsEntries)
				if err != nil {
					panic(err)
				}
			}
		}

		log.Printf("setDnsEntries domainName: %+v\n", domainName)
		log.Printf("setDnsEntries dnsEntries: %+v\n", dnsEntries)

		if f.setDNSEntriesLatency != 0 {
			time.Sleep(f.setDNSEntriesLatency)
		}

		f.dnsEntries = dnsEntries.Items
		return nil
	default:
		return nil
	}
}

var envTest = tester.NewEnvTest(
	"TRANSIP_ACCOUNT_NAME",
	"TRANSIP_PRIVATE_KEY_PATH").
	WithDomain("TRANSIP_DOMAIN")

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				"TRANSIP_ACCOUNT_NAME":     "johndoe",
				"TRANSIP_PRIVATE_KEY_PATH": "./fixtures/private.key",
			},
		},
		{
			desc: "missing all credentials",
			envVars: map[string]string{
				"TRANSIP_ACCOUNT_NAME":     "",
				"TRANSIP_PRIVATE_KEY_PATH": "",
			},
			expected: "transip: some credentials information are missing: TRANSIP_ACCOUNT_NAME,TRANSIP_PRIVATE_KEY_PATH",
		},
		{
			desc: "missing account name",
			envVars: map[string]string{
				"TRANSIP_ACCOUNT_NAME":     "",
				"TRANSIP_PRIVATE_KEY_PATH": "./fixtures/private.key",
			},
			expected: "transip: some credentials information are missing: TRANSIP_ACCOUNT_NAME",
		},
		{
			desc: "missing private key path",
			envVars: map[string]string{
				"TRANSIP_ACCOUNT_NAME":     "johndoe",
				"TRANSIP_PRIVATE_KEY_PATH": "",
			},
			expected: "transip: some credentials information are missing: TRANSIP_PRIVATE_KEY_PATH",
		},
		{
			desc: "could not open private key path",
			envVars: map[string]string{
				"TRANSIP_ACCOUNT_NAME":     "johndoe",
				"TRANSIP_PRIVATE_KEY_PATH": "./fixtures/non/existent/private.key",
			},
			expected: "transip: could not open private key: stat ./fixtures/non/existent/private.key: no such file or directory",
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
				require.NotNil(t, p.client)
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
			expected:    "transip: PrivateKeyPath or PrivateKeyBody is required",
		},
		{
			desc:           "could not open private key path",
			accountName:    "johndoe",
			privateKeyPath: "./fixtures/non/existent/private.key",
			expected:       "transip: could not open private key: stat ./fixtures/non/existent/private.key: no such file or directory",
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
				require.NotNil(t, p.client)
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
	}

	p := &DNSProvider{
		config: NewDefaultConfig(),
		client: client,
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()

		domain1 := "bar.lego.wtf"

		time.Sleep(500 * time.Millisecond)
		err := p.Present(domain1, "", "")
		require.NoError(t, err)

		time.Sleep(100 * time.Millisecond)
		var found bool
		for _, entry := range client.dnsEntries {
			if strings.HasSuffix(entry.Name, ".bar") {
				found = true
			}
		}
		assert.True(t, found, "record bar not found: %v", client.dnsEntries)

		time.Sleep(100 * time.Millisecond)
		err = p.CleanUp(domain1, "", "")
		require.NoError(t, err)
	}()

	go func() {
		defer wg.Done()

		domain2 := "foo.lego.wtf"

		time.Sleep(500 * time.Millisecond)
		err := p.Present(domain2, "", "")
		require.NoError(t, err)

		time.Sleep(200 * time.Millisecond)
		var found bool
		for _, entry := range client.dnsEntries {
			if strings.HasSuffix(entry.Name, ".foo") {
				found = true
			}
		}
		assert.True(t, found, "record foo not found: %v", client.dnsEntries)

		time.Sleep(100 * time.Millisecond)
		err = p.CleanUp(domain2, "", "")
		require.NoError(t, err)
	}()

	wg.Wait()

	assert.Empty(t, client.dnsEntries)
}

func TestDNSProvider_concurrentSetDNSEntries(t *testing.T) {
	client := &fakeClient{}

	p := &DNSProvider{
		config: NewDefaultConfig(),
		client: client,
	}

	token := ""
	keyAuth := ""

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		time.Sleep(550 * time.Millisecond)
		err := p.Present("bar.lego.wtf", token, keyAuth)
		require.NoError(t, err)

		time.Sleep(500 * time.Millisecond)
		err = p.CleanUp("bar.lego.wtf", token, keyAuth)
		require.NoError(t, err)
	}()

	go func() {
		defer wg.Done()
		time.Sleep(500 * time.Millisecond)
		err := p.Present("foo.lego.wtf", token, keyAuth)
		require.NoError(t, err)

		time.Sleep(100 * time.Millisecond)
		err = p.CleanUp("foo.lego.wtf", token, keyAuth)
		require.NoError(t, err)
	}()

	wg.Wait()

	assert.Empty(t, client.dnsEntries)
}

func readArgs(req gotransip.SoapRequest) []string {
	v := reflect.ValueOf(req)
	f := v.FieldByName("args")

	var args []string
	for i := 0; i < f.Len(); i++ {
		args = append(args, f.Slice(0, f.Len()).Index(i).String())
	}

	return args
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
