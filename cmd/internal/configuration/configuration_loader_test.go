package configuration

import (
	"testing"

	"github.com/go-acme/lego/v5/lego"
	"github.com/stretchr/testify/assert"
)

func TestGetServerConfig(t *testing.T) {
	testCases := []struct {
		desc     string
		cfg      *Configuration
		expected *Server
	}{
		{
			desc: "No server on account",
			cfg: &Configuration{
				Accounts: map[string]*Account{
					"acc01": {},
				},
			},
			expected: &Server{
				URL:                 lego.DirectoryURLLetsEncrypt,
				OverallRequestLimit: 18,
			},
		},
		{
			desc: "server shortcode on account",
			cfg: &Configuration{
				Accounts: map[string]*Account{
					"acc01": {
						Server: lego.CodeLetsEncryptStaging,
					},
				},
			},
			expected: &Server{
				URL:                 lego.DirectoryURLLetsEncryptStaging,
				OverallRequestLimit: 18,
			},
		},
		{
			desc: "server URL on account",
			cfg: &Configuration{
				Accounts: map[string]*Account{
					"acc01": {
						Server: "https://example.com/acme/directory",
					},
				},
			},
			expected: &Server{
				URL:                 "https://example.com/acme/directory",
				OverallRequestLimit: 18,
			},
		},
		{
			desc: "server configuration reference",
			cfg: &Configuration{
				Servers: map[string]*Server{
					"foo": {
						URL:                 "https://example.com/acme/directory",
						OverallRequestLimit: 6,
					},
				},
				Accounts: map[string]*Account{
					"acc01": {
						Server: "foo",
					},
				},
			},
			expected: &Server{
				URL:                 "https://example.com/acme/directory",
				OverallRequestLimit: 6,
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			srv := GetServerConfig(test.cfg, "acc01")

			assert.Equal(t, test.expected, srv)
		})
	}
}
