package autodns

import (
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"github.com/go-acme/lego/v3/platform/tester"
	"github.com/stretchr/testify/assert"
)

var envTest = tester.NewEnvTest(envApiEndpoint, envApiUser, envApiPassword)

func TestNewDNSProvider(t *testing.T) {
	defaultEndpointURL, _ := url.Parse(defaultEndpoint)
	examplEndpointUrl, _ := url.Parse(demoEndpoint)

	tests := []struct {
		name    string
		want    *DNSProvider
		wantErr bool
		env     map[string]string
	}{
		{
			name: "complete, no errors",
			want: &DNSProvider{
				config: &Config{
					Endpoint:   defaultEndpointURL,
					Username:   "test",
					Password:   "1234",
					Context:    4,
					HTTPClient: &http.Client{},
				},
			},
			env: map[string]string{
				envApiUser:     "test",
				envApiPassword: "1234",
			},
		},
		{
			name: "different endpoint url",
			want: &DNSProvider{
				config: &Config{
					Endpoint:   examplEndpointUrl,
					Username:   "test",
					Password:   "1234",
					Context:    4,
					HTTPClient: &http.Client{},
				},
			},
			env: map[string]string{
				envApiUser:     "test",
				envApiPassword: "1234",
				envApiEndpoint: demoEndpoint,
			},
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
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewDNSProvider() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDNSProvider_Present(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()
	provider, err := NewDNSProvider()
	assert.NoError(t, err)

	err = provider.Present(envTest.GetDomain(), "", "123d==")
	assert.NoError(t, err)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()
	provider, err := NewDNSProvider()
	assert.NoError(t, err)

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	assert.NoError(t, err)
}
