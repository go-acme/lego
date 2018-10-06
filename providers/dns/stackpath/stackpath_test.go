package stackpath

import (
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	stackpathLiveTest     bool
	stackpathClientID     string
	stackpathClientSecret string
	stackpathStackID      string
	stackpathDomain       string
)

func init() {
	stackpathClientID = os.Getenv("STACKPATH_CLIENT_ID")
	stackpathClientSecret = os.Getenv("STACKPATH_CLIENT_SECRET")
	stackpathStackID = os.Getenv("STACKPATH_STACK_ID")
	stackpathDomain = os.Getenv("STACKPATH_DOMAIN")

	if len(stackpathClientID) > 0 &&
		len(stackpathClientSecret) > 0 &&
		len(stackpathStackID) > 0 &&
		len(stackpathDomain) > 0 {
		stackpathLiveTest = true
	}
}

func TestLiveStackpathPresent(t *testing.T) {
	if !stackpathLiveTest {
		t.Skip("skipping live test")
	}

	config := NewDefaultConfig()
	config.ClientID = stackpathClientID
	config.ClientSecret = stackpathClientSecret
	config.StackID = stackpathStackID

	provider, err := NewDNSProviderConfig(config)
	require.NoError(t, err)

	err = provider.Present(stackpathDomain, "", "123d==")
	require.NoError(t, err)
}

func TestLiveStackpathCleanUp(t *testing.T) {
	if !stackpathLiveTest {
		t.Skip("skipping live test")
	}

	time.Sleep(time.Second * 1)

	config := NewDefaultConfig()
	config.ClientID = stackpathClientID
	config.ClientSecret = stackpathClientSecret
	config.StackID = stackpathStackID

	provider, err := NewDNSProviderConfig(config)
	require.NoError(t, err)

	err = provider.CleanUp(stackpathDomain, "", "123d==")
	require.NoError(t, err)
}

func TestNewDNSProviderConfig(t *testing.T) {
	tests := map[string]struct {
		config      *Config
		want        *DNSProvider
		expectedErr string
	}{
		"no_config": {
			config:      nil,
			expectedErr: "the configuration of the DNS provider is nil",
		},
		"no_client_id": {
			config: &Config{
				ClientSecret: "secret",
				StackID:      "stackID",
			},
			expectedErr: "credentials missing",
		},
		"no_client_secret": {
			config: &Config{
				ClientID: "clientID",
				StackID:  "stackID",
			},
			expectedErr: "credentials missing",
		},
		"no_stack_id": {
			config: &Config{
				ClientID:     "clientID",
				ClientSecret: "secret",
			},
			expectedErr: "stack id missing",
		},
	}
	for ttName, tt := range tests {
		t.Run(ttName, func(t *testing.T) {
			got, err := NewDNSProviderConfig(tt.config)
			require.EqualError(t, err, tt.expectedErr)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewDNSProviderConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
