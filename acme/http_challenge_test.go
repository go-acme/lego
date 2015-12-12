package acme

import (
	"crypto/rsa"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/square/go-jose"
)

func TestHTTPNonRootBind(t *testing.T) {
	privKey, _ := generatePrivateKey(rsakey, 128)
	jws := &jws{privKey: privKey.(*rsa.PrivateKey)}

	solver := &httpChallenge{jws: jws}
	clientChallenge := challenge{Type: "http01", Status: "pending", URI: "localhost:4000", Token: "http1"}

	// validate error on non-root bind to 80
	if err := solver.Solve(clientChallenge, "127.0.0.1"); err == nil {
		t.Error("BIND: Expected Solve to return an error but the error was nil.")
	} else {
		expectedError := "Could not start HTTP server for challenge -> listen tcp :80: bind: permission denied"
		if err.Error() != expectedError {
			t.Errorf("Expected error \"%s\" but instead got \"%s\"", expectedError, err.Error())
		}
	}
}

func TestHTTPShortRSA(t *testing.T) {
	privKey, _ := generatePrivateKey(rsakey, 128)
	jws := &jws{privKey: privKey.(*rsa.PrivateKey), nonces: []string{"test1", "test2"}}

	solver := &httpChallenge{jws: jws, optPort: "23456"}
	clientChallenge := challenge{Type: "http01", Status: "pending", URI: "http://localhost:4000", Token: "http2"}

	if err := solver.Solve(clientChallenge, "127.0.0.1"); err == nil {
		t.Error("UNEXPECTED: Expected Solve to return an error but the error was nil.")
	} else {
		expectedError := "Failed to post JWS message. -> crypto/rsa: message too long for RSA public key size"
		if err.Error() != expectedError {
			t.Errorf("Expected error %s but instead got %s", expectedError, err.Error())
		}
	}
}

func TestHTTPConnectionRefusal(t *testing.T) {
	privKey, _ := generatePrivateKey(rsakey, 512)
	jws := &jws{privKey: privKey.(*rsa.PrivateKey), nonces: []string{"test1", "test2"}}

	solver := &httpChallenge{jws: jws, optPort: "23456"}
	clientChallenge := challenge{Type: "http01", Status: "pending", URI: "http://localhost:4000", Token: "http3"}

	if err := solver.Solve(clientChallenge, "127.0.0.1"); err == nil {
		t.Error("UNEXPECTED: Expected Solve to return an error but the error was nil.")
	} else {
		reg := "/Failed to post JWS message\\. -> Post http:\\/\\/localhost:4000: dial tcp 127\\.0\\.0\\.1:4000: (getsockopt: )?connection refused/g"
		test2 := "Failed to post JWS message. -> Post http://localhost:4000: dial tcp 127.0.0.1:4000: connection refused"
		r, _ := regexp.Compile(reg)
		if r.MatchString(err.Error()) && r.MatchString(test2) {
			t.Errorf("Expected \"%s\" to match %s", err.Error(), reg)
		}
	}
}

func TestHTTPUnexpectedServerState(t *testing.T) {
	privKey, _ := generatePrivateKey(rsakey, 512)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Replay-Nonce", "12345")
		w.Write([]byte("{\"type\":\"http01\",\"status\":\"what\",\"uri\":\"http://some.url\",\"token\":\"http4\"}"))
	}))

	jws := &jws{privKey: privKey.(*rsa.PrivateKey), directoryURL: ts.URL}
	solver := &httpChallenge{jws: jws, optPort: "23456"}
	clientChallenge := challenge{Type: "http01", Status: "pending", URI: ts.URL, Token: "http4"}

	if err := solver.Solve(clientChallenge, "127.0.0.1"); err == nil {
		t.Error("UNEXPECTED: Expected Solve to return an error but the error was nil.")
	} else {
		expectedError := "The server returned an unexpected state."
		if err.Error() != expectedError {
			t.Errorf("Expected error %s but instead got %s", expectedError, err.Error())
		}
	}
}

func TestHTTPChallengeServerUnexpectedDomain(t *testing.T) {
	privKey, _ := generatePrivateKey(rsakey, 512)
	jws := &jws{privKey: privKey.(*rsa.PrivateKey)}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			tr := &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}
			client := &http.Client{Transport: tr}
			req, _ := client.Get("https://localhost:23456/.well-known/acme-challenge/" + "htto5")
			reqBytes, _ := ioutil.ReadAll(req.Body)
			if string(reqBytes) != "TEST" {
				t.Error("Expected http01 server to return string TEST on unexpected domain.")
			}
		}

		w.Header().Add("Replay-Nonce", "12345")
		w.Write([]byte("{\"type\":\"http01\",\"status\":\"invalid\",\"uri\":\"http://some.url\",\"token\":\"http5\"}"))
	}))

	solver := &httpChallenge{jws: jws, optPort: "23456"}
	clientChallenge := challenge{Type: "http01", Status: "pending", URI: ts.URL, Token: "http5"}

	if err := solver.Solve(clientChallenge, "127.0.0.1"); err == nil {
		t.Error("UNEXPECTED: Expected Solve to return an error but the error was nil.")
	}
}

func TestHTTPServerError(t *testing.T) {
	privKey, _ := generatePrivateKey(rsakey, 512)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "HEAD" {
			w.Header().Add("Replay-Nonce", "12345")
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Add("Replay-Nonce", "12345")
			w.Write([]byte("{\"type\":\"urn:acme:error:unauthorized\",\"detail\":\"Error creating new authz :: Syntax error\"}"))
		}
	}))

	jws := &jws{privKey: privKey.(*rsa.PrivateKey), directoryURL: ts.URL}
	solver := &httpChallenge{jws: jws, optPort: "23456"}
	clientChallenge := challenge{Type: "http01", Status: "pending", URI: ts.URL, Token: "http6"}

	if err := solver.Solve(clientChallenge, "127.0.0.1"); err == nil {
		t.Error("UNEXPECTED: Expected Solve to return an error but the error was nil.")
	} else {
		expectedError := "acme: Error 500 - urn:acme:error:unauthorized - Error creating new authz :: Syntax error"
		if err.Error() != expectedError {
			t.Errorf("Expected error |%s| but instead got |%s|", expectedError, err.Error())
		}
	}
}

func TestHTTPInvalidServerState(t *testing.T) {
	privKey, _ := generatePrivateKey(rsakey, 512)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Replay-Nonce", "12345")
		w.Write([]byte("{\"type\":\"http01\",\"status\":\"invalid\",\"uri\":\"http://some.url\",\"token\":\"http7\"}"))
	}))

	jws := &jws{privKey: privKey.(*rsa.PrivateKey), directoryURL: ts.URL}
	solver := &httpChallenge{jws: jws, optPort: "23456"}
	clientChallenge := challenge{Type: "http01", Status: "pending", URI: ts.URL, Token: "http7"}

	if err := solver.Solve(clientChallenge, "127.0.0.1"); err == nil {
		t.Error("UNEXPECTED: Expected Solve to return an error but the error was nil.")
	} else {
		expectedError := "acme: Error 0 -  - \nError Detail:\n"
		if err.Error() != expectedError {
			t.Errorf("Expected error |%s| but instead got |%s|", expectedError, err.Error())
		}
	}
}

func TestHTTPValidServerResponse(t *testing.T) {
	privKey, _ := generatePrivateKey(rsakey, 512)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Replay-Nonce", "12345")
		w.Write([]byte("{\"type\":\"http01\",\"status\":\"valid\",\"uri\":\"http://some.url\",\"token\":\"http8\"}"))
	}))

	jws := &jws{privKey: privKey.(*rsa.PrivateKey), directoryURL: ts.URL}
	solver := &httpChallenge{jws: jws, optPort: "23456"}
	clientChallenge := challenge{Type: "http01", Status: "pending", URI: ts.URL, Token: "http8"}

	if err := solver.Solve(clientChallenge, "127.0.0.1"); err != nil {
		t.Errorf("VALID: Expected Solve to return no error but the error was -> %v", err)
	}
}

func TestHTTPValidFull(t *testing.T) {
	privKey, _ := generatePrivateKey(rsakey, 512)

	ts := httptest.NewServer(nil)

	jws := &jws{privKey: privKey.(*rsa.PrivateKey), directoryURL: ts.URL}
	solver := &httpChallenge{jws: jws, optPort: "23457"}
	clientChallenge := challenge{Type: "http01", Status: "pending", URI: ts.URL, Token: "http9"}

	// Validate server on port 23456 which responds appropriately
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

		transport := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
		client := &http.Client{Transport: transport}

		reqURL := "http://localhost:23457/.well-known/acme-challenge/" + clientChallenge.Token
		t.Logf("Request URL is: %s", reqURL)
		req, err := http.NewRequest("GET", reqURL, nil)
		if err != nil {
			t.Error(err)
		}
		req.Host = "127.0.0.1"
		resp, err := client.Do(req)
		if err != nil {
			t.Errorf("Expected the solver to listen on port 23457 -> %v", err)
		}
		defer resp.Body.Close()

		body, _ := ioutil.ReadAll(resp.Body)
		bodyStr := string(body)

		if resp.Header.Get("Content-Type") != "text/plain" {
			t.Errorf("Expected server to respond with content type text/plain.")
		}

		tokenRegex := regexp.MustCompile("^[\\w-]{43}$")
		parts := strings.Split(bodyStr, ".")

		if len(parts) != 2 {
			t.Errorf("Expected server token to be a composite of two strings, seperated by a dot")
		}

		if parts[0] != clientChallenge.Token {
			t.Errorf("Expected the first part of the server token to be the challenge token.")
		}

		if !tokenRegex.MatchString(parts[1]) {
			t.Errorf("Expected the second part of the server token to be a properly formatted key authorization")
		}

		valid := challenge{Type: "http01", Status: "valid", URI: ts.URL, Token: "1234567812"}
		jsonBytes, _ := json.Marshal(&valid)
		w.Write(jsonBytes)
	})

	if err := solver.Solve(clientChallenge, "127.0.0.1"); err != nil {
		t.Errorf("VALID: Expected Solve to return no error but the error was -> %v", err)
	}
}
