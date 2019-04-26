// Package bindman implements a DNS provider for solving the DNS-01 challenge.
package bindman

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/go-acme/lego/platform/tester"
	bindmanClient "github.com/labbsr0x/bindman-dns-webhook/src/client"
	"github.com/stretchr/testify/require"
)

var envTest = tester.NewEnvTest(
	"BINDMAN_MANAGER_ADDRESS").
	WithDomain("BINDMAN_DOMAIN")

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				"BINDMAN_MANAGER_ADDRESS": "http://localhost",
			},
		},
		{
			desc: "missing bindman manager address",
			envVars: map[string]string{
				"BINDMAN_MANAGER_ADDRESS": "",
			},
			expected: "bindman: some credentials information are missing: BINDMAN_MANAGER_ADDRESS",
		},
		{
			desc: "empty bindman manager address",
			envVars: map[string]string{
				"BINDMAN_MANAGER_ADDRESS": "  ",
			},
			expected: "bindman: managerAddress parameter must be a non-empty string",
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
		desc     string
		config   *Config
		expected string
	}{
		{
			desc:   "success",
			config: &Config{BaseURL: "http://localhost"},
		},
		{
			desc:     "missing base URL",
			config:   &Config{BaseURL: ""},
			expected: "bindman: bindman manager address missing",
		},
		{
			desc:     "missing base URL",
			config:   &Config{BaseURL: "  "},
			expected: "bindman: managerAddress parameter must be a non-empty string",
		},
		{
			desc:     "missing config",
			expected: "bindman: the configuration of the DNS provider is nil",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			p, err := NewDNSProviderConfig(test.config)

			if len(test.expected) == 0 {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.config)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestDNSProvider_Present(t *testing.T) {
	type fields struct {
		config *Config
		client *bindmanClient.DNSWebhookClient
	}
	type args struct {
		domain  string
		token   string
		keyAuth string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "success when add record function return no error",
			fields: fields{
				client: &bindmanClient.DNSWebhookClient{
					ClientAPI: &MockHTTPClientAPI{Status: http.StatusNoContent},
				},
			},
			args: args{
				domain:  "hello.test.com",
				keyAuth: "szDTG4zmM0GsKG91QAGO2M4UYOJMwU8oFpWOP7eTjCw",
			},
			wantErr: false,
		},
		{
			name: "error when add record function return an error",
			fields: fields{
				client: &bindmanClient.DNSWebhookClient{
					ClientAPI: &MockHTTPClientAPI{Error: errors.New("error adding record")},
				},
			},
			args: args{
				domain:  "hello.test.com",
				keyAuth: "szDTG4zmM0GsKG91QAGO2M4UYOJMwU8oFpWOP7eTjCw",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DNSProvider{
				config: tt.fields.config,
				client: tt.fields.client,
			}
			if err := d.Present(tt.args.domain, tt.args.token, tt.args.keyAuth); (err != nil) != tt.wantErr {
				t.Errorf("DNSProvider.Present() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDNSProvider_CleanUp(t *testing.T) {
	type fields struct {
		config *Config
		client *bindmanClient.DNSWebhookClient
	}
	type args struct {
		domain  string
		token   string
		keyAuth string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "success when remove record function return no error",
			fields: fields{
				client: &bindmanClient.DNSWebhookClient{
					ClientAPI: &MockHTTPClientAPI{Status: http.StatusNoContent},
				},
			},
			args: args{
				domain:  "hello.test.com",
				keyAuth: "szDTG4zmM0GsKG91QAGO2M4UYOJMwU8oFpWOP7eTjCw",
			},
			wantErr: false,
		},
		{
			name: "error when remove record function return an error",
			fields: fields{
				client: &bindmanClient.DNSWebhookClient{
					ClientAPI: &MockHTTPClientAPI{Error: errors.New("error adding record")},
				},
			},
			args: args{
				domain:  "hello.test.com",
				keyAuth: "szDTG4zmM0GsKG91QAGO2M4UYOJMwU8oFpWOP7eTjCw",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DNSProvider{
				config: tt.fields.config,
				client: tt.fields.client,
			}
			if err := d.CleanUp(tt.args.domain, tt.args.token, tt.args.keyAuth); (err != nil) != tt.wantErr {
				t.Errorf("DNSProvider.CleanUp() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
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

type MockHTTPClientAPI struct {
	Data   []byte
	Status int
	Error  error
}

func (m *MockHTTPClientAPI) Put(url string, data []byte) (*http.Response, []byte, error) {
	return &http.Response{StatusCode: m.Status}, m.Data, m.Error
}

func (m *MockHTTPClientAPI) Post(url string, data []byte) (*http.Response, []byte, error) {
	return &http.Response{StatusCode: m.Status}, m.Data, m.Error
}

func (m *MockHTTPClientAPI) Get(url string) (*http.Response, []byte, error) {
	return &http.Response{StatusCode: m.Status}, m.Data, m.Error
}
func (m *MockHTTPClientAPI) Delete(url string) (*http.Response, []byte, error) {
	return &http.Response{StatusCode: m.Status}, m.Data, m.Error
}
