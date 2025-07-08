package internal

import (
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	noopBaseURL          = "https://api.reg.ru/api/regru2/nop"
	officialTestUser     = "test"
	officialTestPassword = "test"
)

func TestRemoveRecord(t *testing.T) {
	// TODO(ldez): remove skip when the reg.ru API will be fixed.
	t.Skip("there is a bug with the reg.ru API: INTERNAL_API_ERROR: Внутренняя ошибка, status code: 503")

	client := NewClient(officialTestUser, officialTestPassword)
	client.HTTPClient = &http.Client{Timeout: 30 * time.Second}

	err := client.RemoveTxtRecord(t.Context(), "test.ru", "_acme-challenge", "txttxttxt")
	require.NoError(t, err)
}

func TestRemoveRecord_errors(t *testing.T) {
	// TODO(ldez): remove skip when the reg.ru API will be fixed.
	if os.Getenv("CI") == "true" {
		t.Skip("there is a bug with the reg.ru and GitHub action: dial tcp 194.58.116.30:443: i/o timeout")
	}

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
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			client := NewClient(test.username, test.username)
			client.HTTPClient = &http.Client{Timeout: 30 * time.Second}
			client.baseURL, _ = url.Parse(test.baseURL)

			err := client.RemoveTxtRecord(t.Context(), test.domain, "_acme-challenge", "txttxttxt")
			require.EqualError(t, err, test.expected)
		})
	}
}

func TestAddTXTRecord(t *testing.T) {
	// TODO(ldez): remove skip when the reg.ru API will be fixed.
	t.Skip("there is a bug with the reg.ru API: INTERNAL_API_ERROR: Внутренняя ошибка, status code: 503")

	client := NewClient(officialTestUser, officialTestPassword)
	client.HTTPClient = &http.Client{Timeout: 30 * time.Second}

	err := client.AddTXTRecord(t.Context(), "test.ru", "_acme-challenge", "txttxttxt")
	require.NoError(t, err)
}

func TestAddTXTRecord_errors(t *testing.T) {
	// TODO(ldez): remove skip when the reg.ru API will be fixed.
	if os.Getenv("CI") == "true" {
		t.Skip("there is a bug with the reg.ru and GitHub action: dial tcp 194.58.116.30:443: i/o timeout")
	}

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
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			client := NewClient(test.username, test.username)
			client.HTTPClient = &http.Client{Timeout: 30 * time.Second}
			client.baseURL, _ = url.Parse(test.baseURL)

			err := client.AddTXTRecord(t.Context(), test.domain, "_acme-challenge", "txttxttxt")
			require.EqualError(t, err, test.expected)
		})
	}
}
