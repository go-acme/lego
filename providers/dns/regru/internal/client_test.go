package internal

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	noopBaseURL          = "https://api.reg.ru/api/regru2/nop"
	officialTestUser     = "test"
	officialTestPassword = "test"
)

func TestRemoveRecord(t *testing.T) {
	client := NewClient(officialTestUser, officialTestPassword)

	err := client.RemoveTxtRecord("test.ru", "_acme-challenge", "txttxttxt")
	require.NoError(t, err)
}

func TestRemoveRecord_errors(t *testing.T) {
	testCases := []struct {
		desc     string
		domain   string
		username string
		password string
		baseURL  string
		expected string
	}{
		{
			desc:     "authentication failed",
			domain:   "test.ru",
			username: "",
			password: "",
			baseURL:  noopBaseURL,
			expected: "API error: NO_AUTH: No authorization mechanism selected",
		},
		{
			desc:     "domain error",
			domain:   "",
			username: officialTestUser,
			password: officialTestPassword,
			baseURL:  defaultBaseURL,
			expected: "API error: NO_DOMAIN: domain_name not given or empty",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			client := NewClient(test.username, test.username)

			err := client.RemoveTxtRecord(test.domain, "_acme-challenge", "txttxttxt")
			require.EqualError(t, err, test.expected)
		})
	}
}

func TestAddTXTRecord(t *testing.T) {
	client := NewClient(officialTestUser, officialTestPassword)

	err := client.AddTXTRecord("test.ru", "_acme-challenge", "txttxttxt")
	require.NoError(t, err)
}

func TestAddTXTRecord_errors(t *testing.T) {
	testCases := []struct {
		desc     string
		domain   string
		username string
		password string
		baseURL  string
		expected string
	}{
		{
			desc:     "authentication failed",
			domain:   "test.ru",
			username: "",
			password: "",
			baseURL:  noopBaseURL,
			expected: "API error: NO_AUTH: No authorization mechanism selected",
		},
		{
			desc:     "domain error",
			domain:   "",
			username: officialTestUser,
			password: officialTestPassword,
			baseURL:  defaultBaseURL,
			expected: "API error: NO_DOMAIN: domain_name not given or empty",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			client := NewClient(test.username, test.username)

			err := client.AddTXTRecord(test.domain, "_acme-challenge", "txttxttxt")
			require.EqualError(t, err, test.expected)
		})
	}
}
