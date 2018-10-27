package acme

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDNSValidServerResponse(t *testing.T) {
	backupPreCheckDNS := PreCheckDNS
	defer func() {
		PreCheckDNS = backupPreCheckDNS
	}()

	PreCheckDNS = func(fqdn, value string) (bool, error) {
		return true, nil
	}

	privKey, err := rsa.GenerateKey(rand.Reader, 512)
	require.NoError(t, err)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Replay-Nonce", "12345")

		_, err = w.Write([]byte("{\"type\":\"dns01\",\"status\":\"valid\",\"uri\":\"http://some.url\",\"token\":\"http8\"}"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}))

	go func() {
		time.Sleep(time.Second * 2)
		f := bufio.NewWriter(os.Stdout)
		defer f.Flush()
		_, _ = f.WriteString("\n")
	}()

	manualProvider, err := NewDNSProviderManual()
	require.NoError(t, err)

	clientChallenge := challenge{Type: "dns01", Status: "pending", URL: ts.URL, Token: "http8"}

	solver := &dnsChallenge{
		jws:      &jws{privKey: privKey, getNonceURL: ts.URL},
		validate: validate,
		provider: manualProvider,
	}

	err = solver.Solve(clientChallenge, "example.com")
	require.NoError(t, err)
}

func TestToFqdn(t *testing.T) {
	testCases := []struct {
		desc     string
		domain   string
		expected string
	}{
		{
			desc:     "simple",
			domain:   "foo.bar.com",
			expected: "foo.bar.com.",
		},
		{
			desc:     "already FQDN",
			domain:   "foo.bar.com.",
			expected: "foo.bar.com.",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			fqdn := ToFqdn(test.domain)
			assert.Equal(t, test.expected, fqdn)
		})
	}
}

func TestUnFqdn(t *testing.T) {
	testCases := []struct {
		desc     string
		fqdn     string
		expected string
	}{
		{
			desc:     "simple",
			fqdn:     "foo.bar.com.",
			expected: "foo.bar.com",
		},
		{
			desc:     "already domain",
			fqdn:     "foo.bar.com",
			expected: "foo.bar.com",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			domain := UnFqdn(test.fqdn)

			assert.Equal(t, test.expected, domain)
		})
	}
}
