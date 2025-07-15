package namecheap

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	envTestUser     = "foo"
	envTestKey      = "bar"
	envTestClientIP = "10.0.0.1"
)

type testCase struct {
	name             string
	domain           string
	errString        string
	getHostsResponse string
	setHostsResponse string
}

var testCases = []testCase{
	{
		name:             "Test:Success:1",
		domain:           "test.example.com",
		getHostsResponse: "getHosts_success1.xml",
		setHostsResponse: "setHosts_success1.xml",
	},
	{
		name:             "Test:Success:2",
		domain:           "example.com",
		getHostsResponse: "getHosts_success2.xml",
		setHostsResponse: "setHosts_success2.xml",
	},
	{
		name:             "Test:Error:BadApiKey:1",
		domain:           "test.example.com",
		errString:        "API Key is invalid or API access has not been enabled [1011102]",
		getHostsResponse: "getHosts_errorBadAPIKey1.xml",
	},
}

func TestDNSProvider_Present(t *testing.T) {
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			ch, _ := newPseudoRecord(test.domain, "")

			provider := mockBuilder().
				Route("GET /",
					servermock.ResponseFromInternal(test.getHostsResponse),
					servermock.CheckForm().Strict().
						With("ClientIp", "10.0.0.1").
						With("Command", "namecheap.domains.dns.getHosts").
						With("SLD", ch.sld).
						With("TLD", ch.tld).
						With("UserName", "foo").
						With("ApiKey", "bar").
						With("ApiUser", "foo"),
				).
				Route("POST /",
					servermock.ResponseFromInternal(test.setHostsResponse),
					servermock.CheckForm().
						With("ClientIp", "10.0.0.1").
						With("Command", "namecheap.domains.dns.setHosts").
						With("SLD", ch.sld).
						With("TLD", ch.tld).
						With("UserName", "foo").
						With("ApiKey", "bar").
						With("ApiUser", "foo"),
				).
				Build(t)

			err := provider.Present(test.domain, "", "dummyKey")
			if test.errString != "" {
				assert.EqualError(t, err, "namecheap: "+test.errString)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDNSProvider_CleanUp(t *testing.T) {
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			ch, _ := newPseudoRecord(test.domain, "")

			provider := mockBuilder().
				Route("GET /",
					servermock.ResponseFromInternal(test.getHostsResponse),
					servermock.CheckForm().Strict().
						With("ClientIp", "10.0.0.1").
						With("Command", "namecheap.domains.dns.getHosts").
						With("SLD", ch.sld).
						With("TLD", ch.tld).
						With("UserName", "foo").
						With("ApiKey", "bar").
						With("ApiUser", "foo"),
				).
				Route("POST /",
					servermock.ResponseFromInternal(test.setHostsResponse),
					servermock.CheckForm().
						With("ClientIp", "10.0.0.1").
						With("Command", "namecheap.domains.dns.setHosts").
						With("SLD", ch.sld).
						With("TLD", ch.tld).
						With("UserName", "foo").
						With("ApiKey", "bar").
						With("ApiUser", "foo"),
				).
				Build(t)

			err := provider.CleanUp(test.domain, "", "dummyKey")
			if test.errString != "" {
				assert.EqualError(t, err, "namecheap: "+test.errString)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func Test_newPseudoRecord_domainSplit(t *testing.T) {
	tests := []struct {
		domain string
		valid  bool
		tld    string
		sld    string
		host   string
	}{
		{domain: "a.b.c.test.co.uk", valid: true, tld: "co.uk", sld: "test", host: "a.b.c"},
		{domain: "test.co.uk", valid: true, tld: "co.uk", sld: "test"},
		{domain: "test.com", valid: true, tld: "com", sld: "test"},
		{domain: "test.co.com", valid: true, tld: "co.com", sld: "test"},
		{domain: "www.test.com.au", valid: true, tld: "com.au", sld: "test", host: "www"},
		{domain: "www.za.com", valid: true, tld: "za.com", sld: "www"},
		{domain: "my.test.tf", valid: true, tld: "tf", sld: "test", host: "my"},
		{},
		{domain: "a"},
		{domain: "com"},
		{domain: "com.au"},
		{domain: "co.com"},
		{domain: "co.uk"},
		{domain: "tf"},
		{domain: "za.com"},
	}

	for _, test := range tests {
		t.Run(test.domain, func(t *testing.T) {
			valid := true
			ch, err := newPseudoRecord(test.domain, "")
			if err != nil {
				valid = false
			}

			if test.valid && !valid {
				t.Errorf("Expected '%s' to split", test.domain)
			} else if !test.valid && valid {
				t.Errorf("Expected '%s' to produce error", test.domain)
			}

			if test.valid && valid {
				require.NotNil(t, ch)
				assert.Equal(t, test.domain, ch.domain, "domain")
				assert.Equal(t, test.tld, ch.tld, "tld")
				assert.Equal(t, test.sld, ch.sld, "sld")
				assert.Equal(t, test.host, ch.host, "host")
			}
		})
	}
}

func mockBuilder() *servermock.Builder[*DNSProvider] {
	return servermock.NewBuilder(func(server *httptest.Server) (*DNSProvider, error) {
		config := NewDefaultConfig()
		config.BaseURL = server.URL
		config.APIUser = envTestUser
		config.APIKey = envTestKey
		config.ClientIP = envTestClientIP
		config.HTTPClient = &http.Client{Timeout: 60 * time.Second}

		return NewDNSProviderConfig(config)
	})
}
