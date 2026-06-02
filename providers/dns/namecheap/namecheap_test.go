package namecheap

import (
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v5/internal/tester/servermock"
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
			ch, _ := newPseudoRecord(t.Context(), test.domain, "")

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

			err := provider.Present(t.Context(), test.domain, "", "dummyKey")
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
			ch, _ := newPseudoRecord(t.Context(), test.domain, "")

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

			err := provider.CleanUp(t.Context(), test.domain, "", "dummyKey")
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
		domain   string
		valid    bool
		expected *pseudoRecord
	}{
		{
			domain: "a.b.c.test.co.uk",
			valid:  true,
			expected: &pseudoRecord{
				domain:   "a.b.c.test.co.uk",
				key:      "_acme-challenge.a.b.c",
				keyFqdn:  "_acme-challenge.a.b.c.test.co.uk.",
				keyValue: "47DEQpj8HBSa-_TImW-5JCeuQeRkm5NMpJWZG3hSuFU",
				tld:      "co.uk",
				sld:      "test",
				host:     "a.b.c",
			},
		},
		{
			domain: "test.co.uk",
			valid:  true,
			expected: &pseudoRecord{
				domain:   "test.co.uk",
				key:      "_acme-challenge",
				keyFqdn:  "_acme-challenge.test.co.uk.",
				keyValue: "47DEQpj8HBSa-_TImW-5JCeuQeRkm5NMpJWZG3hSuFU",
				tld:      "co.uk",
				sld:      "test",
				host:     "",
			},
		},
		{
			domain: "test.com",
			valid:  true,
			expected: &pseudoRecord{
				domain:   "test.com",
				key:      "_acme-challenge",
				keyFqdn:  "_acme-challenge.test.com.",
				keyValue: "47DEQpj8HBSa-_TImW-5JCeuQeRkm5NMpJWZG3hSuFU",
				tld:      "com",
				sld:      "test",
				host:     "",
			},
		},
		{
			domain: "test.co.com",
			valid:  true,
			expected: &pseudoRecord{
				domain:   "test.co.com",
				key:      "_acme-challenge",
				keyFqdn:  "_acme-challenge.test.co.com.",
				keyValue: "47DEQpj8HBSa-_TImW-5JCeuQeRkm5NMpJWZG3hSuFU",
				tld:      "co.com",
				sld:      "test",
				host:     "",
			},
		},
		{
			domain: "www.test.com.au",
			valid:  true,
			expected: &pseudoRecord{
				domain:   "www.test.com.au",
				key:      "_acme-challenge.www",
				keyFqdn:  "_acme-challenge.www.test.com.au.",
				keyValue: "47DEQpj8HBSa-_TImW-5JCeuQeRkm5NMpJWZG3hSuFU",
				tld:      "com.au",
				sld:      "test",
				host:     "www",
			},
		},
		{
			domain: "www.za.com",
			valid:  true,
			expected: &pseudoRecord{
				domain:   "www.za.com",
				key:      "_acme-challenge",
				keyFqdn:  "_acme-challenge.www.za.com.",
				keyValue: "47DEQpj8HBSa-_TImW-5JCeuQeRkm5NMpJWZG3hSuFU",
				tld:      "za.com",
				sld:      "www",
			},
		},
		{
			domain: "my.test.tf",
			valid:  true,
			expected: &pseudoRecord{
				domain:   "my.test.tf",
				key:      "_acme-challenge.my",
				keyFqdn:  "_acme-challenge.my.test.tf.",
				keyValue: "47DEQpj8HBSa-_TImW-5JCeuQeRkm5NMpJWZG3hSuFU",
				tld:      "tf",
				sld:      "test",
				host:     "my",
			},
		},
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

			pr, err := newPseudoRecord(t.Context(), test.domain, "")
			if err != nil {
				valid = false
			}

			if test.valid && !valid {
				t.Errorf("Expected '%s' to split", test.domain)
			} else if !test.valid && valid {
				t.Errorf("Expected '%s' to produce error", test.domain)
			}

			if test.valid && valid {
				require.NotNil(t, pr)

				assert.Equal(t, test.expected, pr)
			}
		})
	}
}

func mockBuilder() *servermock.Builder[*DNSProvider] {
	return servermock.NewBuilder(func(server *httptest.Server) (*DNSProvider, error) {
		config := NewDefaultConfig()
		config.HTTPClient = server.Client()
		config.BaseURL = server.URL
		config.APIUser = envTestUser
		config.APIKey = envTestKey
		config.ClientIP = envTestClientIP

		return NewDNSProviderConfig(config)
	})
}
