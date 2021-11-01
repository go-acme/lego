package namecheap

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	envTestUser     = "foo"
	envTestKey      = "bar"
	envTestClientIP = "10.0.0.1"
)

func TestDNSProvider_getHosts(t *testing.T) {
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			p := setupTest(t, &test)

			ch, err := newChallenge(test.domain, "")
			require.NoError(t, err)

			hosts, err := p.getHosts(ch.sld, ch.tld)
			if test.errString != "" {
				assert.EqualError(t, err, test.errString)
			} else {
				assert.NoError(t, err)
			}

		next1:
			for _, h := range hosts {
				for _, th := range test.hosts {
					if h == th {
						continue next1
					}
				}
				t.Errorf("getHosts case %s unexpected record [%s:%s:%s]", test.name, h.Type, h.Name, h.Address)
			}

		next2:
			for _, th := range test.hosts {
				for _, h := range hosts {
					if h == th {
						continue next2
					}
				}
				t.Errorf("getHosts case %s missing record [%s:%s:%s]", test.name, th.Type, th.Name, th.Address)
			}
		})
	}
}

func TestDNSProvider_setHosts(t *testing.T) {
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			p := setupTest(t, &test)

			ch, err := newChallenge(test.domain, "")
			require.NoError(t, err)

			hosts, err := p.getHosts(ch.sld, ch.tld)
			if test.errString != "" {
				assert.EqualError(t, err, test.errString)
			} else {
				require.NoError(t, err)
			}
			if err != nil {
				return
			}

			err = p.setHosts(ch.sld, ch.tld, hosts)
			require.NoError(t, err)
		})
	}
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
				fmt.Fprint(w, tc.getHostsResponse)
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
				fmt.Fprint(w, tc.setHostsResponse)
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

type testCase struct {
	name             string
	domain           string
	hosts            []Record
	errString        string
	getHostsResponse string
	setHostsResponse string
}

var testCases = []testCase{
	{
		name:   "Test:Success:1",
		domain: "test.example.com",
		hosts: []Record{
			{Type: "A", Name: "home", Address: "10.0.0.1", MXPref: "10", TTL: "1799"},
			{Type: "A", Name: "www", Address: "10.0.0.2", MXPref: "10", TTL: "1200"},
			{Type: "AAAA", Name: "a", Address: "::0", MXPref: "10", TTL: "1799"},
			{Type: "CNAME", Name: "*", Address: "example.com.", MXPref: "10", TTL: "1799"},
			{Type: "MXE", Name: "example.com", Address: "10.0.0.5", MXPref: "10", TTL: "1800"},
			{Type: "URL", Name: "xyz", Address: "https://google.com", MXPref: "10", TTL: "1799"},
		},
		getHostsResponse: responseGetHostsSuccess1,
		setHostsResponse: responseSetHostsSuccess1,
	},
	{
		name:   "Test:Success:2",
		domain: "example.com",
		hosts: []Record{
			{Type: "A", Name: "@", Address: "10.0.0.2", MXPref: "10", TTL: "1200"},
			{Type: "A", Name: "www", Address: "10.0.0.3", MXPref: "10", TTL: "60"},
		},
		getHostsResponse: responseGetHostsSuccess2,
		setHostsResponse: responseSetHostsSuccess2,
	},
	{
		name:             "Test:Error:BadApiKey:1",
		domain:           "test.example.com",
		errString:        "API Key is invalid or API access has not been enabled [1011102]",
		getHostsResponse: responseGetHostsErrorBadAPIKey1,
	},
}

const responseGetHostsSuccess1 = `<?xml version="1.0" encoding="utf-8"?>
<ApiResponse Status="OK" xmlns="http://api.namecheap.com/xml.response">
  <Errors />
  <Warnings />
  <RequestedCommand>namecheap.domains.dns.getHosts</RequestedCommand>
  <CommandResponse Type="namecheap.domains.dns.getHosts">
    <DomainDNSGetHostsResult Domain="example.com" EmailType="MXE" IsUsingOurDNS="true">
      <host HostId="217076" Name="www" Type="A" Address="10.0.0.2" MXPref="10" TTL="1200" AssociatedAppTitle="" FriendlyName="" IsActive="true" IsDDNSEnabled="false" />
      <host HostId="217069" Name="home" Type="A" Address="10.0.0.1" MXPref="10" TTL="1799" AssociatedAppTitle="" FriendlyName="" IsActive="true" IsDDNSEnabled="false" />
      <host HostId="217071" Name="a" Type="AAAA" Address="::0" MXPref="10" TTL="1799" AssociatedAppTitle="" FriendlyName="" IsActive="true" IsDDNSEnabled="false" />
      <host HostId="217075" Name="*" Type="CNAME" Address="example.com." MXPref="10" TTL="1799" AssociatedAppTitle="" FriendlyName="" IsActive="true" IsDDNSEnabled="false" />
      <host HostId="217073" Name="example.com" Type="MXE" Address="10.0.0.5" MXPref="10" TTL="1800" AssociatedAppTitle="MXE" FriendlyName="MXE1" IsActive="true" IsDDNSEnabled="false" />
      <host HostId="217077" Name="xyz" Type="URL" Address="https://google.com" MXPref="10" TTL="1799" AssociatedAppTitle="" FriendlyName="" IsActive="true" IsDDNSEnabled="false" />
    </DomainDNSGetHostsResult>
  </CommandResponse>
  <Server>PHX01SBAPI01</Server>
  <GMTTimeDifference>--5:00</GMTTimeDifference>
  <ExecutionTime>3.338</ExecutionTime>
</ApiResponse>`

const responseSetHostsSuccess1 = `<?xml version="1.0" encoding="utf-8"?>
<ApiResponse Status="OK" xmlns="http://api.namecheap.com/xml.response">
  <Errors />
  <Warnings />
  <RequestedCommand>namecheap.domains.dns.setHosts</RequestedCommand>
  <CommandResponse Type="namecheap.domains.dns.setHosts">
    <DomainDNSSetHostsResult Domain="example.com" IsSuccess="true">
      <Warnings />
    </DomainDNSSetHostsResult>
  </CommandResponse>
  <Server>PHX01SBAPI01</Server>
  <GMTTimeDifference>--5:00</GMTTimeDifference>
  <ExecutionTime>2.347</ExecutionTime>
</ApiResponse>`

const responseGetHostsSuccess2 = `<?xml version="1.0" encoding="utf-8"?>
<ApiResponse Status="OK" xmlns="http://api.namecheap.com/xml.response">
  <Errors />
  <Warnings />
  <RequestedCommand>namecheap.domains.dns.getHosts</RequestedCommand>
  <CommandResponse Type="namecheap.domains.dns.getHosts">
    <DomainDNSGetHostsResult Domain="example.com" EmailType="MXE" IsUsingOurDNS="true">
      <host HostId="217076" Name="@" Type="A" Address="10.0.0.2" MXPref="10" TTL="1200" AssociatedAppTitle="" FriendlyName="" IsActive="true" IsDDNSEnabled="false" />
      <host HostId="217069" Name="www" Type="A" Address="10.0.0.3" MXPref="10" TTL="60" AssociatedAppTitle="" FriendlyName="" IsActive="true" IsDDNSEnabled="false" />
    </DomainDNSGetHostsResult>
  </CommandResponse>
  <Server>PHX01SBAPI01</Server>
  <GMTTimeDifference>--5:00</GMTTimeDifference>
  <ExecutionTime>3.338</ExecutionTime>
</ApiResponse>`

const responseSetHostsSuccess2 = `<?xml version="1.0" encoding="utf-8"?>
<ApiResponse Status="OK" xmlns="http://api.namecheap.com/xml.response">
  <Errors />
  <Warnings />
  <RequestedCommand>namecheap.domains.dns.setHosts</RequestedCommand>
  <CommandResponse Type="namecheap.domains.dns.setHosts">
    <DomainDNSSetHostsResult Domain="example.com" IsSuccess="true">
      <Warnings />
    </DomainDNSSetHostsResult>
  </CommandResponse>
  <Server>PHX01SBAPI01</Server>
  <GMTTimeDifference>--5:00</GMTTimeDifference>
  <ExecutionTime>2.347</ExecutionTime>
</ApiResponse>`

const responseGetHostsErrorBadAPIKey1 = `<?xml version="1.0" encoding="utf-8"?>
<ApiResponse Status="ERROR" xmlns="http://api.namecheap.com/xml.response">
  <Errors>
    <Error Number="1011102">API Key is invalid or API access has not been enabled</Error>
  </Errors>
  <Warnings />
  <RequestedCommand />
  <Server>PHX01SBAPI01</Server>
  <GMTTimeDifference>--5:00</GMTTimeDifference>
  <ExecutionTime>0</ExecutionTime>
</ApiResponse>`
