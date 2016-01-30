package acme

import (
	"bufio"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestDNSValidServerResponse(t *testing.T) {
	preCheckDNS = func(domain, fqdn string) bool {
		return true
	}
	privKey, _ := generatePrivateKey(rsakey, 512)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Replay-Nonce", "12345")
		w.Write([]byte("{\"type\":\"dns01\",\"status\":\"valid\",\"uri\":\"http://some.url\",\"token\":\"http8\"}"))
	}))

	manualProvider, _ := NewDNSProviderManual()
	jws := &jws{privKey: privKey.(*rsa.PrivateKey), directoryURL: ts.URL}
	solver := &dnsChallenge{jws: jws, validate: validate, provider: manualProvider}
	clientChallenge := challenge{Type: "dns01", Status: "pending", URI: ts.URL, Token: "http8"}

	go func() {
		time.Sleep(time.Second * 2)
		f := bufio.NewWriter(os.Stdout)
		defer f.Flush()
		f.WriteString("\n")
	}()

	if err := solver.Solve(clientChallenge, "example.com"); err != nil {
		t.Errorf("VALID: Expected Solve to return no error but the error was -> %v", err)
	}
}

func TestPreCheckDNS(t *testing.T) {
	if !preCheckDNS("api.letsencrypt.org", "acme-staging.api.letsencrypt.org") {
		t.Errorf("preCheckDNS failed for acme-staging.api.letsencrypt.org")
	}
}
