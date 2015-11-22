package acme

import (
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/square/go-jose"
)

func TestTLSSNINonRootBind(t *testing.T) {
	privKey, _ := generatePrivateKey(rsakey, 128)
	jws := &jws{privKey: privKey.(*rsa.PrivateKey)}

	solver := &tlsSNIChallenge{jws: jws}
	clientChallenge := challenge{Type: "tls-sni-01", Status: "pending", URI: "localhost:4000", Token: "tls1"}

	// validate error on non-root bind to 443
	if err := solver.Solve(clientChallenge, "127.0.0.1"); err == nil {
		t.Error("BIND: Expected Solve to return an error but the error was nil.")
	} else {
		expectedError := "Could not start HTTPS server for challenge -> listen tcp :443: bind: permission denied"
		if err.Error() != expectedError {
			t.Errorf("Expected error \"%s\" but instead got \"%s\"", expectedError, err.Error())
		}
	}
}

func TestTLSSNI(t *testing.T) {
	privKey, _ := generatePrivateKey(rsakey, 512)
	optPort := "5001"

	ts := httptest.NewServer(nil)

	ts.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request challenge
		w.Header().Add("Replay-Nonce", "12345")

		if r.Method == "HEAD" {
			return
		}

		clientJws, _ := ioutil.ReadAll(r.Body)
		j, err := jose.ParseSigned(string(clientJws))
		if err != nil {
			t.Errorf("Client sent invalid JWS to the server.\n\t%v", err)
			return
		}
		output, err := j.Verify(&privKey.(*rsa.PrivateKey).PublicKey)
		if err != nil {
			t.Errorf("Unable to verify client data -> %v", err)
		}
		json.Unmarshal(output, &request)

		conn, err := tls.Dial("tcp", "localhost:"+optPort, &tls.Config{
			InsecureSkipVerify: true,
		})
		if err != nil {
			t.Errorf("Expected to connect to challenge server without an error. %s", err.Error())
		}

		// Expect the server to only return one certificate
		connState := conn.ConnectionState()
		if count := len(connState.PeerCertificates); count != 1 {
			t.Errorf("Expected the challenge server to return exactly one certificate but got %d", count)
		}

		remoteCert := connState.PeerCertificates[0]
		if count := len(remoteCert.DNSNames); count != 1 {
			t.Errorf("Expected the challenge certificate to have exactly one DNSNames entry but had %d", count)
		}

		zBytes := sha256.Sum256([]byte(request.KeyAuthorization))
		z := hex.EncodeToString(zBytes[:sha256.Size])
		domain := fmt.Sprintf("%s.%s.acme.invalid", z[:32], z[32:])

		if remoteCert.DNSNames[0] != domain {
			t.Errorf("Expected the challenge certificate DNSName to match %s but was %s", domain, remoteCert.DNSNames[0])
		}

		valid := challenge{Type: "tls-sni-01", Status: "valid", URI: ts.URL, Token: "tls1"}
		jsonBytes, _ := json.Marshal(&valid)
		w.Write(jsonBytes)
	})

	jws := &jws{privKey: privKey.(*rsa.PrivateKey), directoryURL: ts.URL}
	solver := &tlsSNIChallenge{jws: jws, optPort: optPort}
	clientChallenge := challenge{Type: "tls-sni-01", Status: "pending", URI: ts.URL, Token: "tls1"}

	if err := solver.Solve(clientChallenge, "127.0.0.1"); err != nil {
		t.Error("UNEXPECTED: Expected Solve to return no error but the error was %s.", err.Error())
	}
}
