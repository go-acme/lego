package autodns

import (
	"reflect"
	"testing"

	"github.com/go-acme/lego/v3/platform/tester"
	"github.com/stretchr/testify/assert"
)

var envTest = tester.NewEnvTest(envApiEndpoint)

func TestNewDNSProvider(t *testing.T) {
	tests := []struct {
		name    string
		want    *DNSProvider
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
