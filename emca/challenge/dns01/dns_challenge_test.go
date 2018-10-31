package dns01

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
	"github.com/xenolf/lego/emca/internal/secure"
	"github.com/xenolf/lego/emca/le"
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

	clientChallenge := le.Challenge{Type: "dns01", Status: "pending", URL: ts.URL, Token: "http8"}

	solver := &Challenge{
		jws:      secure.NewJWS(nil, privKey, ts.URL),
		validate: stubValidate,
		provider: manualProvider,
	}

	err = solver.Solve(clientChallenge, "example.com")
	require.NoError(t, err)
}

// FIXME remove?
// stubValidate is like validate, except it does nothing.
func stubValidate(_ *secure.JWS, _, _ string, _ le.Challenge) error {
	return nil
}
