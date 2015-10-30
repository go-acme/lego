package acme

import (
	"crypto/rsa"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/square/go-jose"
)

func TestSimpleHTTPNonRootBind(t *testing.T) {
	privKey, _ := generatePrivateKey(rsakey, 128)
	jws := &jws{privKey: privKey.(*rsa.PrivateKey)}

	solver := &simpleHTTPChallenge{jws: jws}
	clientChallenge := challenge{Type: "simpleHttp", Status: "pending", URI: "localhost:4000", Token: "1"}

	// validate error on non-root bind to 443
	if err := solver.Solve(clientChallenge, "test.domain"); err == nil {
		t.Error("BIND: Expected Solve to return an error but the error was nil.")
	} else {
		expectedError := "Could not start HTTPS server for challenge -> listen tcp :443: bind: permission denied"
		if err.Error() != expectedError {
			t.Errorf("Expected error %s but instead got %s", expectedError, err.Error())
		}
	}
}

func TestSimpleHTTPShortRSA(t *testing.T) {
	privKey, _ := generatePrivateKey(rsakey, 128)
	jws := &jws{privKey: privKey.(*rsa.PrivateKey), nonces: []string{"test1", "test2"}}

	solver := &simpleHTTPChallenge{jws: jws, optPort: "23456"}
	clientChallenge := challenge{Type: "simpleHttp", Status: "pending", URI: "http://localhost:4000", Token: "2"}

	if err := solver.Solve(clientChallenge, "test.domain"); err == nil {
		t.Error("UNEXPECTED: Expected Solve to return an error but the error was nil.")
	} else {
		expectedError := "Could not start HTTPS server for challenge -> startHTTPSServer: Failed to sign message. crypto/rsa: message too long for RSA public key size"
		if err.Error() != expectedError {
			t.Errorf("Expected error %s but instead got %s", expectedError, err.Error())
		}
	}
}

func TestSimpleHTTPConnectionRefusal(t *testing.T) {
	privKey, _ := generatePrivateKey(rsakey, 512)
	jws := &jws{privKey: privKey.(*rsa.PrivateKey), nonces: []string{"test1", "test2"}}

	solver := &simpleHTTPChallenge{jws: jws, optPort: "23456"}
	clientChallenge := challenge{Type: "simpleHttp", Status: "pending", URI: "http://localhost:4000", Token: "3"}

	if err := solver.Solve(clientChallenge, "test.domain"); err == nil {
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

func TestSimpleHTTPUnexpectedServerState(t *testing.T) {
	privKey, _ := generatePrivateKey(rsakey, 512)
	jws := &jws{privKey: privKey.(*rsa.PrivateKey)}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Replay-Nonce", "12345")
		w.Write([]byte("{\"type\":\"simpleHttp\",\"status\":\"what\",\"uri\":\"http://some.url\",\"token\":\"4\"}"))
	}))

	solver := &simpleHTTPChallenge{jws: jws, optPort: "23456"}
	clientChallenge := challenge{Type: "simpleHttp", Status: "pending", URI: ts.URL, Token: "4"}

	if err := solver.Solve(clientChallenge, "test.domain"); err == nil {
		t.Error("UNEXPECTED: Expected Solve to return an error but the error was nil.")
	} else {
		expectedError := "The server returned an unexpected state."
		if err.Error() != expectedError {
			t.Errorf("Expected error %s but instead got %s", expectedError, err.Error())
		}
	}
}

func TestSimpleHTTPChallengeServerUnexpectedDomain(t *testing.T) {
	privKey, _ := generatePrivateKey(rsakey, 512)
	jws := &jws{privKey: privKey.(*rsa.PrivateKey)}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			tr := &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}
			client := &http.Client{Transport: tr}
			req, _ := client.Get("https://localhost:23456/.well-known/acme-challenge/" + "5")
			reqBytes, _ := ioutil.ReadAll(req.Body)
			if string(reqBytes) != "TEST" {
				t.Error("Expected simpleHTTP server to return string TEST on unexpected domain.")
			}
		}

		w.Header().Add("Replay-Nonce", "12345")
		w.Write([]byte("{\"type\":\"simpleHttp\",\"status\":\"invalid\",\"uri\":\"http://some.url\",\"token\":\"5\"}"))
	}))

	solver := &simpleHTTPChallenge{jws: jws, optPort: "23456"}
	clientChallenge := challenge{Type: "simpleHttp", Status: "pending", URI: ts.URL, Token: "5"}

	if err := solver.Solve(clientChallenge, "test.domain"); err == nil {
		t.Error("UNEXPECTED: Expected Solve to return an error but the error was nil.")
	}
}

func TestSimpleHTTPServerError(t *testing.T) {
	privKey, _ := generatePrivateKey(rsakey, 512)
	jws := &jws{privKey: privKey.(*rsa.PrivateKey)}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "HEAD" {
			w.Header().Add("Replay-Nonce", "12345")
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Add("Replay-Nonce", "12345")
			w.Write([]byte("{\"type\":\"urn:acme:error:unauthorized\",\"detail\":\"Error creating new authz :: Syntax error\"}"))
		}
	}))

	solver := &simpleHTTPChallenge{jws: jws, optPort: "23456"}
	clientChallenge := challenge{Type: "simpleHttp", Status: "pending", URI: ts.URL, Token: "6"}

	if err := solver.Solve(clientChallenge, "test.domain"); err == nil {
		t.Error("UNEXPECTED: Expected Solve to return an error but the error was nil.")
	} else {
		expectedError := "[500] Type: urn:acme:error:unauthorized Detail: Error creating new authz :: Syntax error"
		if err.Error() != expectedError {
			t.Errorf("Expected error |%s| but instead got |%s|", expectedError, err.Error())
		}
	}
}

func TestSimpleHTTPInvalidServerState(t *testing.T) {
	privKey, _ := generatePrivateKey(rsakey, 512)
	jws := &jws{privKey: privKey.(*rsa.PrivateKey)}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Replay-Nonce", "12345")
		w.Write([]byte("{\"type\":\"simpleHttp\",\"status\":\"invalid\",\"uri\":\"http://some.url\",\"token\":\"7\"}"))
	}))

	solver := &simpleHTTPChallenge{jws: jws, optPort: "23456"}
	clientChallenge := challenge{Type: "simpleHttp", Status: "pending", URI: ts.URL, Token: "7"}

	if err := solver.Solve(clientChallenge, "test.domain"); err == nil {
		t.Error("UNEXPECTED: Expected Solve to return an error but the error was nil.")
	} else {
		expectedError := "The server could not validate our request."
		if err.Error() != expectedError {
			t.Errorf("Expected error |%s| but instead got |%s|", expectedError, err.Error())
		}
	}
}

func TestSimpleHTTPValidServerResponse(t *testing.T) {
	privKey, _ := generatePrivateKey(rsakey, 512)
	jws := &jws{privKey: privKey.(*rsa.PrivateKey)}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Replay-Nonce", "12345")
		w.Write([]byte("{\"type\":\"simpleHttp\",\"status\":\"valid\",\"uri\":\"http://some.url\",\"token\":\"8\"}"))
	}))

	solver := &simpleHTTPChallenge{jws: jws, optPort: "23456"}
	clientChallenge := challenge{Type: "simpleHttp", Status: "pending", URI: ts.URL, Token: "8"}

	if err := solver.Solve(clientChallenge, "test.domain"); err != nil {
		t.Errorf("VALID: Expected Solve to return no error but the error was -> %v", err)
	}
}

func TestSimpleHTTPValidFull(t *testing.T) {
	privKey, _ := generatePrivateKey(rsakey, 512)
	jws := &jws{privKey: privKey.(*rsa.PrivateKey)}

	ts := httptest.NewServer(nil)

	solver := &simpleHTTPChallenge{jws: jws, optPort: "23456"}
	clientChallenge := challenge{Type: "simpleHttp", Status: "pending", URI: ts.URL, Token: "9"}

	// Validate server on port 23456 which responds appropriately
	clientChallenge.Token = "1234567812"
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

		reqURL := "https://localhost:23456/.well-known/acme-challenge/" + clientChallenge.Token
		t.Logf("Request URL is: %s", reqURL)
		req, _ := http.NewRequest("GET", reqURL, nil)
		req.Host = "test.domain"
		resp, err := client.Do(req)
		if err != nil {
			t.Errorf("Expected the solver to listen on port 23456 -> %v", err)
		}
		defer resp.Body.Close()

		body, _ := ioutil.ReadAll(resp.Body)
		bodyStr := string(body)
		clientResponse, err := jose.ParseSigned(bodyStr)
		if err != nil {
			t.Errorf("Client answered with invalid JWS.\n\t%v", err)
			return
		}
		_, err = clientResponse.Verify(&privKey.(*rsa.PrivateKey).PublicKey)
		if err != nil {
			t.Errorf("Unable to verify client data -> %v", err)
		}

		valid := challenge{Type: "simpleHttp", Status: "valid", URI: ts.URL, Token: "1234567812"}
		jsonBytes, _ := json.Marshal(&valid)
		w.Write(jsonBytes)
	})

	if err := solver.Solve(clientChallenge, "test.domain"); err != nil {
		t.Errorf("VALID: Expected Solve to return no error but the error was -> %v", err)
	}
}
