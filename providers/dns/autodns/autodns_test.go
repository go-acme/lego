package autodns

import (
	"net/http"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/go-acme/lego/v3/platform/tester"
	"github.com/stretchr/testify/assert"
)

var envTest = tester.NewEnvTest(envAPIEndpoint, envAPIUser, envAPIPassword)

func TestNewDNSProvider(t *testing.T) {
	defaultEndpointURL, _ := url.Parse(defaultEndpoint)
	examplEndpointURL, _ := url.Parse(demoEndpoint)

	tests := []struct {
		name        string
		want        *DNSProvider
		wantErr     bool
		expectedErr string
		env         map[string]string
	}{
		{
			name: "complete, no errors",
			want: &DNSProvider{
				config: &Config{
					Endpoint:           defaultEndpointURL,
					Username:           "test",
					Password:           "1234",
					Context:            defaultEndpointContext,
					TTL:                defaultTTL,
					PropagationTimeout: 2 * time.Minute,
					PollingInterval:    2 * time.Second,
					HTTPClient: &http.Client{
						Timeout: 30 * time.Second,
					},
				},
			},
			env: map[string]string{
				envAPIUser:     "test",
				envAPIPassword: "1234",
			},
		},
		{
			name: "different endpoint url",
			want: &DNSProvider{
				config: &Config{
					Endpoint:           examplEndpointURL,
					Username:           "test",
					Password:           "1234",
					Context:            defaultEndpointContext,
					TTL:                defaultTTL,
					PropagationTimeout: 2 * time.Minute,
					PollingInterval:    2 * time.Second,
					HTTPClient: &http.Client{
						Timeout: 30 * time.Second,
					},
				},
			},
			env: map[string]string{
				envAPIUser:     "test",
				envAPIPassword: "1234",
				envAPIEndpoint: demoEndpoint,
			},
		},
		{
			name: "missing credentials",
			env: map[string]string{
				envAPIUser:     "",
				envAPIPassword: "",
			},
			wantErr:     true,
			expectedErr: "autodns: some credentials information are missing: AUTODNS_API_USER,AUTODNS_API_PASSWORD",
		},
		{
			name: "missing username",
			env: map[string]string{
				envAPIUser:     "",
				envAPIPassword: "1234",
			},
			wantErr:     true,
			expectedErr: "autodns: some credentials information are missing: AUTODNS_API_USER",
		},
		{
			name: "missing password",
			env: map[string]string{
				envAPIUser:     "user",
				envAPIPassword: "",
			},
			wantErr:     true,
			expectedErr: "autodns: some credentials information are missing: AUTODNS_API_PASSWORD",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer envTest.RestoreEnv()
			envTest.ClearEnv()
			envTest.Apply(tt.env)

			got, err := NewDNSProvider()
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDNSProvider() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (err != nil) && tt.wantErr {
				assert.EqualError(t, err, tt.expectedErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewDNSProvider() got = %v, want %v", got, tt.want)
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
	assert.NoError(t, err)

	err = provider.Present(envTest.GetDomain(), "", "123d==")
	assert.NoError(t, err)
}

func TestLiveCleanUp(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()
	provider, err := NewDNSProvider()
	assert.NoError(t, err)

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	assert.NoError(t, err)
}
