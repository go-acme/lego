package namecheap

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/namecheap/internal"
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
	hosts            []internal.Record
	errString        string
	getHostsResponse string
	setHostsResponse string
}

var testCases = []testCase{
	{
		name:   "Test:Success:1",
		domain: "test.example.com",
		hosts: []internal.Record{
			{Type: "A", Name: "home", Address: "10.0.0.1", MXPref: "10", TTL: "1799"},
			{Type: "A", Name: "www", Address: "10.0.0.2", MXPref: "10", TTL: "1200"},
			{Type: "AAAA", Name: "a", Address: "::0", MXPref: "10", TTL: "1799"},
			{Type: "CNAME", Name: "*", Address: "example.com.", MXPref: "10", TTL: "1799"},
			{Type: "MXE", Name: "example.com", Address: "10.0.0.5", MXPref: "10", TTL: "1800"},
			{Type: "URL", Name: "xyz", Address: "https://google.com", MXPref: "10", TTL: "1799"},
		},
		getHostsResponse: "getHosts_success1.xml",
		setHostsResponse: "setHosts_success1.xml",
	},
	{
		name:   "Test:Success:2",
		domain: "example.com",
		hosts: []internal.Record{
			{Type: "A", Name: "@", Address: "10.0.0.2", MXPref: "10", TTL: "1200"},
			{Type: "A", Name: "www", Address: "10.0.0.3", MXPref: "10", TTL: "60"},
		},
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

func setupTest(t *testing.T, tc *testCase) *DNSProvider {
	t.Helper()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			values := r.URL.Query()
			cmd := values.Get("Command")
			switch cmd {
			case "namecheap.domains.dns.getHosts":
				assertHdr(t, tc, &values)
				w.WriteHeader(http.StatusOK)
				writeFixture(w, tc.getHostsResponse)
			default:
				t.Errorf("Unexpected GET command: %s", cmd)
			}

		case http.MethodPost:
			err := r.ParseForm()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			values := r.Form
			cmd := values.Get("Command")
			switch cmd {
			case "namecheap.domains.dns.setHosts":
				assertHdr(t, tc, &values)
				w.WriteHeader(http.StatusOK)
				writeFixture(w, tc.setHostsResponse)
			default:
				t.Errorf("Unexpected POST command: %s", cmd)
			}

		default:
			t.Errorf("Unexpected http method: %s", r.Method)
		}
	})

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	return mockDNSProvider(t, server.URL)
}

func mockDNSProvider(t *testing.T, baseURL string) *DNSProvider {
	t.Helper()

	config := NewDefaultConfig()
	config.BaseURL = baseURL
	config.APIUser = envTestUser
	config.APIKey = envTestKey
	config.ClientIP = envTestClientIP
	config.HTTPClient = &http.Client{Timeout: 60 * time.Second}

	provider, err := NewDNSProviderConfig(config)
	require.NoError(t, err)

	return provider
}

func assertHdr(t *testing.T, tc *testCase, values *url.Values) {
	t.Helper()

	ch, _ := newChallenge(tc.domain, "")
	assert.Equal(t, envTestUser, values.Get("ApiUser"), "ApiUser")
	assert.Equal(t, envTestKey, values.Get("ApiKey"), "ApiKey")
	assert.Equal(t, envTestUser, values.Get("UserName"), "UserName")
	assert.Equal(t, envTestClientIP, values.Get("ClientIp"), "ClientIp")
	assert.Equal(t, ch.sld, values.Get("SLD"), "SLD")
	assert.Equal(t, ch.tld, values.Get("TLD"), "TLD")
}

func writeFixture(rw http.ResponseWriter, filename string) {
	file, err := os.Open(filepath.Join("internal", "fixtures", filename))
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	defer func() { _ = file.Close() }()

	_, _ = io.Copy(rw, file)
}

func TestDNSProvider_Present(t *testing.T) {
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			p := setupTest(t, &test)

			err := p.Present(test.domain, "", "dummyKey")
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
			p := setupTest(t, &test)

			err := p.CleanUp(test.domain, "", "dummyKey")
			if test.errString != "" {
				assert.EqualError(t, err, "namecheap: "+test.errString)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDomainSplit(t *testing.T) {
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
		test := test
		t.Run(test.domain, func(t *testing.T) {
			valid := true
			ch, err := newChallenge(test.domain, "")
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
