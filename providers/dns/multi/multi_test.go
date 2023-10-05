package multi

import (
	"errors"
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const envDomain = "MULTI_DOMAIN"

const (
	EnvAccessKeyID     = "AWS_ACCESS_KEY_ID"
	EnvSecretAccessKey = "AWS_SECRET_ACCESS_KEY"
	EnvRegion          = "AWS_REGION"

	EnvServiceAccount     = "GCE_SERVICE_ACCOUNT"
	EnvServiceAccountFile = "GCE_SERVICE_ACCOUNT_FILE"
	EnvProject            = "GCE_PROJECT"

	SeqTime = 60 * time.Second
)

var envTest = tester.NewEnvTest(
	EnvAccessKeyID,
	EnvSecretAccessKey,
	EnvRegion,
	EnvServiceAccount,
	EnvProject).
	WithDomain(envDomain).
	WithLiveTestRequirements(EnvAccessKeyID, EnvSecretAccessKey, EnvRegion, EnvServiceAccount, EnvProject, envDomain)

type MockDNSProviderGood struct {
	PresentCalled bool
	CleanUpCalled bool
}

func (d *MockDNSProviderGood) Present(domain, token, keyAuth string) error {
	d.PresentCalled = true

	return nil
}

func (d *MockDNSProviderGood) CleanUp(domain, token, keyAuth string) error {
	d.CleanUpCalled = true

	return nil
}

type MockDNSProviderBad struct {
	PresentCalled bool
	CleanUpCalled bool
}

func (d *MockDNSProviderBad) Present(domain, token, keyAuth string) error {
	d.PresentCalled = true

	return errors.New("DNSProviderBad")
}

func (d *MockDNSProviderBad) CleanUp(domain, token, keyAuth string) error {
	d.CleanUpCalled = true

	return errors.New("DNSProviderBad")
}

type MockDNSProviderGoodSequential struct {
	MockDNSProviderGood
}

func (d *MockDNSProviderGoodSequential) Sequential() time.Duration {
	return SeqTime
}

func newDNSProviderGood() *MockDNSProviderGood {
	return &MockDNSProviderGood{
		PresentCalled: false,
		CleanUpCalled: false,
	}
}

func newDNSProviderBad() *MockDNSProviderBad {
	return &MockDNSProviderBad{
		PresentCalled: false,
		CleanUpCalled: false,
	}
}

func newDNSProviderGoodSequential() *MockDNSProviderGoodSequential {
	provider := newDNSProviderGood()

	return &MockDNSProviderGoodSequential{MockDNSProviderGood: *provider}
}

func Test_NewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvAccessKeyID:     "abc",
				EnvSecretAccessKey: "abc",
				EnvRegion:          "us",
				EnvProject:         "",
				EnvServiceAccount:  `{"project_id": "A","type": "service_account","client_email": "foo@bar.com","private_key_id": "pki","private_key": "pk","token_uri": "/token","client_secret": "secret","client_id": "C","refresh_token": "D"}`,
			},
			expected: "",
		},
		{
			desc: "missing gce project",
			envVars: map[string]string{
				EnvAccessKeyID:     "abc",
				EnvSecretAccessKey: "abc",
				EnvRegion:          "us",
				EnvProject:         "",
				EnvServiceAccount:  "",
			},
			expected: "multi: googlecloud: project name missing",
		},
		{
			desc: "success without route53 credentials",
			envVars: map[string]string{
				EnvProject:        "",
				EnvServiceAccount: `{"project_id": "A","type": "service_account","client_email": "foo@bar.com","private_key_id": "pki","private_key": "pk","token_uri": "/token","client_secret": "secret","client_id": "C","refresh_token": "D"}`,
			},
			expected: "",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer envTest.RestoreEnv()
			envTest.ClearEnv()
			envTest.Apply(test.envVars)
			p, err := NewDNSProvider()
			if test.expected == "" {
				require.NoError(t, err)
				require.NotNil(t, p)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.expected)
			}
		})
	}
}

func Test_NewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		config   *Config
		expected string
	}{
		{
			desc: "route53 and gcloud",
			envVars: map[string]string{
				EnvAccessKeyID:     "abc",
				EnvSecretAccessKey: "abc",
				EnvRegion:          "us",
				EnvProject:         "",
				EnvServiceAccount:  `{"project_id": "A","type": "service_account","client_email": "foo@bar.com","private_key_id": "pki","private_key": "pk","token_uri": "/token","client_secret": "secret","client_id": "C","refresh_token": "D"}`,
			},
			config: &Config{
				SubProviders: []string{"route53", "gcloud"},
			},
			expected: "",
		},
		{
			desc: "route53 only",
			envVars: map[string]string{
				EnvAccessKeyID:     "abc",
				EnvSecretAccessKey: "abc",
				EnvRegion:          "us",
			},
			config: &Config{
				SubProviders: []string{"route53"},
			},
			expected: "",
		},
		{
			desc:     "nil config",
			envVars:  map[string]string{},
			config:   nil,
			expected: "multi: the configuration of the DNS provider is nil",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer envTest.RestoreEnv()
			envTest.ClearEnv()
			envTest.Apply(test.envVars)
			p, err := NewDNSProviderConfig(test.config)
			if test.expected == "" {
				require.NoError(t, err)
				require.NotNil(t, p)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.expected)
			}
		})
	}
}

func Test_NewDNSProviderByNames(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		names    []string
		expected string
	}{
		{
			desc: "route53 and googledomains",
			envVars: map[string]string{
				EnvAccessKeyID:     "abc",
				EnvSecretAccessKey: "abc",
				EnvRegion:          "us",
				EnvProject:         "",
				EnvServiceAccount:  `{"project_id": "A","type": "service_account","client_email": "foo@bar.com","private_key_id": "pki","private_key": "pk","token_uri": "/token","client_secret": "secret","client_id": "C","refresh_token": "D"}`,
			},
			names:    []string{"route53", "gcloud"},
			expected: "",
		},
		{
			desc: "route53 only",
			envVars: map[string]string{
				EnvAccessKeyID:     "abc",
				EnvSecretAccessKey: "abc",
				EnvRegion:          "us",
			},
			names:    []string{"route53"},
			expected: "",
		},
		{
			desc:     "null string",
			envVars:  map[string]string{},
			names:    []string{""},
			expected: "multi: unrecognized DNS provider:",
		},
		{
			desc:     "unknown provider",
			envVars:  map[string]string{},
			names:    []string{"foobar"},
			expected: "multi: unrecognized DNS provider: foobar",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer envTest.RestoreEnv()
			envTest.ClearEnv()
			envTest.Apply(test.envVars)
			p, err := NewDNSProviderByNames(test.names...)
			if test.expected == "" {
				require.NoError(t, err)
				require.NotNil(t, p)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.expected)
			}
		})
	}
}

func Test_PresentCleanupSuccess(t *testing.T) {
	defer envTest.RestoreEnv()
	envTest.ClearEnv()

	domain := "example.com"
	keyAuth := "123456d=="
	sub1 := newDNSProviderGood()
	sub2 := newDNSProviderGood()

	p := NewDNSProviderFromOthers(sub1, sub2)

	err := p.Present(domain, "", keyAuth)
	require.NoError(t, err)
	assert.True(t, sub1.PresentCalled)
	assert.True(t, sub2.PresentCalled)

	err = p.CleanUp(domain, "", keyAuth)
	require.NoError(t, err)
	assert.True(t, sub1.CleanUpCalled)
	assert.True(t, sub2.CleanUpCalled)
}

func Test_PresentCleanupError(t *testing.T) {
	defer envTest.RestoreEnv()
	envTest.ClearEnv()

	domain := "example.com"
	keyAuth := "123456d=="
	sub1 := newDNSProviderGood()
	sub2 := newDNSProviderBad()

	p := NewDNSProviderFromOthers(sub1, sub2)

	err := p.Present(domain, "", keyAuth)
	require.Error(t, err)
	require.Contains(t, err.Error(), "DNSProviderBad")

	assert.True(t, sub1.PresentCalled)
	assert.True(t, sub2.PresentCalled)

	err = p.CleanUp(domain, "", keyAuth)
	require.Error(t, err)
	require.Contains(t, err.Error(), "DNSProviderBad")
	assert.True(t, sub1.CleanUpCalled)
	assert.True(t, sub2.CleanUpCalled)
}

func Test_ProviderNonSeqential(t *testing.T) {
	defer envTest.RestoreEnv()
	envTest.ClearEnv()

	sub1 := newDNSProviderGood()
	sub2 := newDNSProviderGood()

	p := NewDNSProviderFromOthers(sub1, sub2)

	_, ok := p.(sequential)
	assert.False(t, ok)
}

func Test_ProviderSeqential(t *testing.T) {
	defer envTest.RestoreEnv()
	envTest.ClearEnv()

	sub1 := newDNSProviderGood()
	sub2 := newDNSProviderGoodSequential()

	p := NewDNSProviderFromOthers(sub1, sub2)

	q, ok := p.(sequential)
	assert.True(t, ok)

	if ok {
		duration := q.Sequential()
		assert.Equal(t, duration, SeqTime)
	}
}
