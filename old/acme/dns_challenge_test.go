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
		jws:      newJWS(privKey, ts.URL),
		validate: validate,
		provider: manualProvider,
	}

	err = solver.Solve(clientChallenge, "example.com")
	require.NoError(t, err)
}
