package configuration

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidate(t *testing.T) {
	cfg, err := ReadConfiguration(filepath.FromSlash("./testdata/simple.yml"))
	require.NoError(t, err)

	ApplyDefaults(cfg)

	err = Validate(cfg)
	require.NoError(t, err)
}

func Test_validateLog(t *testing.T) {
	cfg := &Configuration{
		Log: &Log{Format: "foo"},
	}

	err := validateLog(cfg)

	require.EqualError(t, err, "invalid log format 'foo'")
}

func Test_validateChallenges(t *testing.T) {
	testCases := []struct {
		desc     string
		cfg      *Configuration
		expected string
	}{
		{
			desc:     "no challenge configurations",
			cfg:      &Configuration{},
			expected: "no challenge configurations found",
		},
		{
			desc: "empty challenge name",
			cfg: &Configuration{
				Challenges: map[string]*Challenge{
					"": {TLS: &TLSChallenge{}},
				},
			},
			expected: "the challenge name cannot be empty",
		},
		{
			desc: "no challenges",
			cfg: &Configuration{
				Challenges: map[string]*Challenge{
					"a": {},
				},
			},
			expected: "a: at least one challenge type must be defined",
		},
		{
			desc: "DNS challenge without a provider",
			cfg: &Configuration{
				Challenges: map[string]*Challenge{
					"a": {
						DNS: &DNSChallenge{},
					},
				},
			},
			expected: "a: a provider is required",
		},
		{
			desc: "DNS challenge propagation: wait and DisableAuthoritativeNameservers",
			cfg: &Configuration{
				Challenges: map[string]*Challenge{
					"a": {
						DNS: &DNSChallenge{
							Provider: "foo",
							Propagation: &Propagation{
								DisableAuthoritativeNameservers: true,
								Wait:                            1,
							},
						},
					},
				},
			},
			expected: "a: 'wait' and 'disableAuthoritativeNameservers' are mutually exclusive",
		},
		{
			desc: "DNS challenge propagation: wait and DisableRecursiveNameservers",
			cfg: &Configuration{
				Challenges: map[string]*Challenge{
					"a": {
						DNS: &DNSChallenge{
							Provider: "foo",
							Propagation: &Propagation{
								DisableRecursiveNameservers: true,
								Wait:                        1,
							},
						},
					},
				},
			},
			expected: "a: 'wait' and 'disableRecursiveNameservers' are mutually exclusive",
		},
		{
			desc: "DNS persist challenge propagation: wait and DisableAuthoritativeNameservers",
			cfg: &Configuration{
				Challenges: map[string]*Challenge{
					"a": {
						DNSPersist: &DNSPersistChallenge{
							Propagation: &Propagation{
								DisableAuthoritativeNameservers: true,
								Wait:                            1,
							},
						},
					},
				},
			},
			expected: "a: 'wait' and 'disableAuthoritativeNameservers' are mutually exclusive",
		},
		{
			desc: "DNS persist challenge propagation: wait and DisableRecursiveNameservers",
			cfg: &Configuration{
				Challenges: map[string]*Challenge{
					"a": {
						DNSPersist: &DNSPersistChallenge{
							Propagation: &Propagation{
								DisableRecursiveNameservers: true,
								Wait:                        1,
							},
						},
					},
				},
			},
			expected: "a: 'wait' and 'disableRecursiveNameservers' are mutually exclusive",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			err := validateChallenges(test.cfg)

			require.EqualError(t, err, test.expected)
		})
	}
}

func Test_validateCertificates(t *testing.T) {
	testCases := []struct {
		desc     string
		cfg      *Configuration
		expected string
	}{
		{
			desc:     "no certificates",
			cfg:      &Configuration{},
			expected: "no certificate configurations found",
		},
		{
			desc: "empty name",
			cfg: &Configuration{
				Certificates: map[string]*Certificate{
					"": {},
				},
			},
			expected: "the certificate name cannot be empty",
		},
		{
			desc: "domain or CSR missing",
			cfg: &Configuration{
				Certificates: map[string]*Certificate{
					"a": {},
				},
			},
			expected: "a: at least one domain or CSR must be provided",
		},
		{
			desc: "domain and CSR",
			cfg: &Configuration{
				Certificates: map[string]*Certificate{
					"a": {
						Domains: []string{"example.com"},
						CSR:     "foo",
					},
				},
			},
			expected: "a: domains and CSR are mutually exclusive",
		},
		{
			desc: "missing account",
			cfg: &Configuration{
				Certificates: map[string]*Certificate{
					"a": {Domains: []string{"example.com"}},
				},
			},
			expected: "a: an account is required",
		},
		{
			desc: "missing challenge",
			cfg: &Configuration{
				Certificates: map[string]*Certificate{
					"a": {
						Account: "acc",
						Domains: []string{"example.com"},
					},
				},
			},
			expected: "a: a challenge is required",
		},
		{
			desc: "not existing account",
			cfg: &Configuration{
				Certificates: map[string]*Certificate{
					"a": {
						Account:   "acc",
						Challenge: "chlg",
						Domains:   []string{"example.com"},
					},
				},
			},
			expected: "a: account: 'acc' not found",
		},
		{
			desc: "not existing challenge",
			cfg: &Configuration{
				Accounts: map[string]*Account{
					"acc": {},
				},
				Certificates: map[string]*Certificate{
					"a": {
						Account:   "acc",
						Challenge: "chlg",
						Domains:   []string{"example.com"},
					},
				},
			},
			expected: "a: challenge: 'chlg' not found",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			err := validateCertificates(test.cfg)

			require.EqualError(t, err, test.expected)
		})
	}
}

func Test_validateServers(t *testing.T) {
	cfg := &Configuration{Servers: map[string]*Server{"": {}}}

	err := validateServers(cfg)

	require.EqualError(t, err, "the server name cannot be empty")
}

func Test_validateAccounts(t *testing.T) {
	testCases := []struct {
		desc     string
		cfg      *Configuration
		expected string
	}{
		{
			desc:     "no accounts",
			cfg:      &Configuration{},
			expected: "no account configurations found",
		},
		{
			desc: "empty name",
			cfg: &Configuration{
				Accounts: map[string]*Account{
					"": {},
				},
			},
			expected: "account '': the account name cannot be empty",
		},
		{
			desc: "missing KID and HMAC key",
			cfg: &Configuration{
				Accounts: map[string]*Account{
					"a": {
						ExternalAccountBinding: &ExternalAccountBinding{},
					},
				},
			},
			expected: "account 'a': KID and HMAC key must be provided for External Account Binding",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			err := validateAccounts(test.cfg)

			require.EqualError(t, err, test.expected)
		})
	}
}
