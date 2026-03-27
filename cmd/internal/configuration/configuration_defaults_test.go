package configuration

import (
	"testing"

	"github.com/go-acme/lego/v5/certcrypto"
	"github.com/go-acme/lego/v5/certificate"
	"github.com/go-acme/lego/v5/lego"
	"github.com/stretchr/testify/assert"
)

func TestApplyDefaults(t *testing.T) {
	cfg := &Configuration{}

	ApplyDefaults(cfg)

	expected := &Configuration{
		Storage: defaultLegoDirectory,
		Servers: map[string]*Server{},
		Accounts: map[string]*Account{
			DefaultAccountID: {
				Server:  lego.DirectoryURLLetsEncrypt,
				KeyType: certcrypto.EC256,
			},
		},
		Challenges:   map[string]*Challenge{},
		Certificates: map[string]*Certificate{},
	}

	assert.Equal(t, expected, cfg)
}

func Test_applyServersDefaults(t *testing.T) {
	testCases := []struct {
		desc     string
		cfg      *Configuration
		expected *Configuration
	}{
		{
			desc:     "no servers",
			cfg:      &Configuration{},
			expected: &Configuration{},
		},
		{
			desc: "empty server",
			cfg: &Configuration{
				Servers: map[string]*Server{
					"a": {},
				},
			},
			expected: &Configuration{
				Servers: map[string]*Server{
					"a": {OverallRequestLimit: certificate.DefaultOverallRequestLimit},
				},
			},
		},
		{
			desc: "overall request limit already set",
			cfg: &Configuration{
				Servers: map[string]*Server{
					"a": {OverallRequestLimit: 6},
				},
			},
			expected: &Configuration{
				Servers: map[string]*Server{
					"a": {OverallRequestLimit: 6},
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			applyServersDefaults(test.cfg)

			assert.Equal(t, test.expected, test.cfg)
		})
	}
}

func Test_applyAccountsDefaults(t *testing.T) {
	testCases := []struct {
		desc     string
		cfg      *Configuration
		expected *Configuration
	}{
		{
			desc:     "no accounts",
			cfg:      &Configuration{},
			expected: &Configuration{},
		},
		{
			desc: "empty account",
			cfg: &Configuration{
				Accounts: map[string]*Account{
					"a": {},
				},
			},
			expected: &Configuration{
				Accounts: map[string]*Account{
					"a": {Server: lego.DirectoryURLLetsEncrypt, KeyType: certcrypto.EC256},
				},
			},
		},
		{
			desc: "no server",
			cfg: &Configuration{
				Accounts: map[string]*Account{
					"a": {KeyType: certcrypto.RSA2048},
				},
			},
			expected: &Configuration{
				Accounts: map[string]*Account{
					"a": {Server: lego.DirectoryURLLetsEncrypt, KeyType: certcrypto.RSA2048},
				},
			},
		},
		{
			desc: "no key type",
			cfg: &Configuration{
				Accounts: map[string]*Account{
					"c": {Server: "https://localhost:14000/dir"},
				},
			},
			expected: &Configuration{
				Accounts: map[string]*Account{
					"c": {Server: "https://localhost:14000/dir", KeyType: certcrypto.EC256},
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			applyAccountsDefaults(test.cfg)

			assert.Equal(t, test.expected, test.cfg)
		})
	}
}

func Test_applyChallengesDefaults(t *testing.T) {
	testCases := []struct {
		desc     string
		cfg      *Configuration
		expected *Configuration
	}{
		{
			desc:     "no challenges",
			cfg:      &Configuration{},
			expected: &Configuration{},
		},
		{
			desc: "empty TLS challenge",
			cfg: &Configuration{
				Challenges: map[string]*Challenge{
					"a": {
						TLS: &TLSChallenge{},
					},
				},
			},
			expected: &Configuration{
				Challenges: map[string]*Challenge{
					"a": {
						TLS: &TLSChallenge{Address: defaultTLSAddress},
					},
				},
			},
		},
		{
			desc: "empty HTTP challenge",
			cfg: &Configuration{
				Challenges: map[string]*Challenge{
					"a": {
						HTTP: &HTTPChallenge{},
					},
				},
			},
			expected: &Configuration{
				Challenges: map[string]*Challenge{
					"a": {
						HTTP: &HTTPChallenge{Address: defaultHTTPAddress},
					},
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			applyChallengesDefaults(test.cfg)

			assert.Equal(t, test.expected, test.cfg)
		})
	}
}

func Test_applyCertificatesDefaults(t *testing.T) {
	testCases := []struct {
		desc     string
		cfg      *Configuration
		expected *Configuration
	}{
		{
			desc:     "no certificates",
			cfg:      &Configuration{},
			expected: &Configuration{},
		},
		{
			desc: "empty certificate",
			cfg: &Configuration{
				Certificates: map[string]*Certificate{
					"a": {},
				},
			},
			expected: &Configuration{
				Certificates: map[string]*Certificate{
					"a": {
						Account: DefaultAccountID,
						KeyType: certcrypto.EC256,
						Renew: &RenewConfiguration{
							ARI: &ARIConfiguration{},
						},
					},
				},
			},
		},
		{
			desc: "empty certificate, with one account",
			cfg: &Configuration{
				Accounts: map[string]*Account{
					"acc": {},
				},
				Certificates: map[string]*Certificate{
					"a": {},
				},
			},
			expected: &Configuration{
				Accounts: map[string]*Account{
					"acc": {},
				},
				Certificates: map[string]*Certificate{
					"a": {
						Account: "acc",
						Renew: &RenewConfiguration{
							ARI: &ARIConfiguration{},
						},
					},
				},
			},
		},
		{
			desc: "empty certificate, with multiple accounts",
			cfg: &Configuration{
				Accounts: map[string]*Account{
					"accA": {},
					"accB": {},
				},
				Certificates: map[string]*Certificate{
					"a": {},
				},
			},
			expected: &Configuration{
				Accounts: map[string]*Account{
					"accA": {},
					"accB": {},
				},
				Certificates: map[string]*Certificate{
					"a": {
						Account: "",
						KeyType: certcrypto.EC256,
						Renew: &RenewConfiguration{
							ARI: &ARIConfiguration{},
						},
					},
				},
			},
		},
		{
			desc: "explicit account",
			cfg: &Configuration{
				Certificates: map[string]*Certificate{
					"a": {Account: "acc"},
				},
			},
			expected: &Configuration{
				Certificates: map[string]*Certificate{
					"a": {
						Account: "acc",
						KeyType: certcrypto.EC256,
						Renew: &RenewConfiguration{
							ARI: &ARIConfiguration{},
						},
					},
				},
			},
		},
		{
			desc: "explicit renew",
			cfg: &Configuration{
				Certificates: map[string]*Certificate{
					"a": {
						Renew: &RenewConfiguration{
							ReuseKey: true,
						},
					},
				},
			},
			expected: &Configuration{
				Certificates: map[string]*Certificate{
					"a": {
						Account: DefaultAccountID,
						KeyType: certcrypto.EC256,
						Renew: &RenewConfiguration{
							ReuseKey: true,
							ARI:      &ARIConfiguration{},
						},
					},
				},
			},
		},
		{
			desc: "explicit ari",
			cfg: &Configuration{
				Certificates: map[string]*Certificate{
					"a": {
						Renew: &RenewConfiguration{
							ARI: &ARIConfiguration{
								Disable: true,
							},
						},
					},
				},
			},
			expected: &Configuration{
				Certificates: map[string]*Certificate{
					"a": {
						Account: DefaultAccountID,
						KeyType: certcrypto.EC256,
						Renew: &RenewConfiguration{
							ARI: &ARIConfiguration{
								Disable: true,
							},
						},
					},
				},
			},
		},
		{
			desc: "default HTTP challenge",
			cfg: &Configuration{
				Challenges: map[string]*Challenge{},
				Certificates: map[string]*Certificate{
					"a": {
						Challenge: defaultHTTP01,
					},
				},
			},
			expected: &Configuration{
				Challenges: map[string]*Challenge{
					defaultHTTP01: {HTTP: &HTTPChallenge{
						Address: defaultHTTPAddress,
					}},
				},
				Certificates: map[string]*Certificate{
					"a": {
						Account:   DefaultAccountID,
						Challenge: defaultHTTP01,
						KeyType:   certcrypto.EC256,
						Renew: &RenewConfiguration{
							ARI: &ARIConfiguration{},
						},
					},
				},
			},
		},
		{
			desc: "default TLS challenge",
			cfg: &Configuration{
				Challenges: map[string]*Challenge{},
				Certificates: map[string]*Certificate{
					"a": {
						Challenge: defaultTLSALPN01,
					},
				},
			},
			expected: &Configuration{
				Challenges: map[string]*Challenge{
					defaultTLSALPN01: {TLS: &TLSChallenge{
						Address: defaultTLSAddress,
					}},
				},
				Certificates: map[string]*Certificate{
					"a": {
						Account:   DefaultAccountID,
						Challenge: defaultTLSALPN01,
						KeyType:   certcrypto.EC256,
						Renew: &RenewConfiguration{
							ARI: &ARIConfiguration{},
						},
					},
				},
			},
		},
		{
			desc: "default if only one challenge",
			cfg: &Configuration{
				Challenges: map[string]*Challenge{
					"chlgA": {DNS: &DNSChallenge{Provider: "foo"}},
				},
				Certificates: map[string]*Certificate{
					"a": {},
				},
			},
			expected: &Configuration{
				Challenges: map[string]*Challenge{
					"chlgA": {DNS: &DNSChallenge{Provider: "foo"}},
				},
				Certificates: map[string]*Certificate{
					"a": {
						Account:   DefaultAccountID,
						KeyType:   certcrypto.EC256,
						Challenge: "chlgA",
						Renew: &RenewConfiguration{
							ARI: &ARIConfiguration{},
						},
					},
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			applyCertificatesDefaults(test.cfg)

			assert.Equal(t, test.expected, test.cfg)
		})
	}
}
