package gcore

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/require"
)

var envTest = tester.NewEnvTest(envPermanentToken).WithDomain(envNamespace + "DOMAIN")

type mockClient struct{}

func (m mockClient) AddTXTRecord(ctx context.Context, fqdn, value string, ttl int) error {
	if strings.Contains(fqdn, "err") {
		return fmt.Errorf("err")
	}
	return nil
}

func (m mockClient) RemoveTXTRecord(ctx context.Context, fqdn, value string) error {
	if strings.Contains(fqdn, "err") {
		return fmt.Errorf("err")
	}
	return nil
}

func TestNewDefaultConfig(t *testing.T) {
	tests := []struct {
		name string
		exec func()
		want Config
	}{
		{
			name: "default",
			exec: func() {},
			want: Config{
				PropagationTimeout: defaultPropagationTimeout,
				PollingInterval:    defaultPollingInterval,
				TTL:                dns01.DefaultTTL,
				HTTPTimeout:        defaultPropagationTimeout,
			},
		},
		{
			name: "custom",
			exec: func() {
				_ = os.Setenv(envTTL, fmt.Sprintf("%d", 10))
				_ = os.Setenv(envHTTPTimeout, fmt.Sprintf("%d", 1))
				_ = os.Setenv(envPollingInterval, fmt.Sprintf("%d", 4))
				_ = os.Setenv(envPropagationTimeout, fmt.Sprintf("%d", 6))
			},
			want: Config{
				PropagationTimeout: 6 * time.Second,
				PollingInterval:    4 * time.Second,
				TTL:                10,
				HTTPTimeout:        time.Second,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.exec()
			if got := NewDefaultConfig(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewDefaultConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDNSProvider_Present(t *testing.T) {
	type args struct {
		domain  string
		token   string
		keyAuth string
	}
	tests := []struct {
		name     string
		provider *DNSProvider
		args     args
		wantErr  bool
	}{
		{
			name: "success",
			provider: &DNSProvider{
				Config: NewDefaultConfig(),
				Client: mockClient{},
			},
			args: args{
				domain:  "any",
				token:   "",
				keyAuth: "",
			},
			wantErr: false,
		},
		{
			name: "error",
			provider: &DNSProvider{
				Config: NewDefaultConfig(),
				Client: mockClient{},
			},
			args: args{
				domain:  "err",
				token:   "",
				keyAuth: "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := tt.provider
			if err := d.Present(tt.args.domain, tt.args.token, tt.args.keyAuth); (err != nil) != tt.wantErr {
				t.Errorf("Present() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDNSProvider_CleanUp(t *testing.T) {
	type args struct {
		domain  string
		token   string
		keyAuth string
	}
	tests := []struct {
		name     string
		provider *DNSProvider
		args     args
		wantErr  bool
	}{
		{
			name: "success",
			provider: &DNSProvider{
				Config: NewDefaultConfig(),
				Client: mockClient{},
			},
			args: args{
				domain:  "any",
				token:   "",
				keyAuth: "",
			},
			wantErr: false,
		},
		{
			name: "error",
			provider: &DNSProvider{
				Config: NewDefaultConfig(),
				Client: mockClient{},
			},
			args: args{
				domain:  "err",
				token:   "",
				keyAuth: "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := tt.provider
			if err := d.CleanUp(tt.args.domain, tt.args.token, tt.args.keyAuth); (err != nil) != tt.wantErr {
				t.Errorf("CleanUp() error = %v, wantErr %v", err, tt.wantErr)
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

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}
