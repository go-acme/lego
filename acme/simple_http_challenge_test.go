package acme

import (
	"crypto/rsa"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/square/go-jose"
)

func TestSimpleHTTPCanSolve(t *testing.T) {
	challenge := &simpleHTTPChallenge{}

	// determine public ip
	resp, err := http.Get("https://icanhazip.com/")
	if err != nil {
		t.Errorf("Could not get public IP -> %v", err)
	}
	defer resp.Body.Close()

	ip, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Could not get public IP -> %v", err)
	}
	ipStr := string(ip)

	if expected, actual := false, challenge.CanSolve("google.com"); expected != actual {
		t.Errorf("Expected CanSolve to return %t for domain 'google.com' but was %t", expected, actual)
	}

	localResolv := strings.Replace(ipStr, "\n", "", -1) + ".xip.io"
	if expected, actual := true, challenge.CanSolve(localResolv); expected != actual {
		t.Errorf("Expected CanSolve to return %t for domain 'localhost' but was %t", expected, actual)
	}
}

func TestSimpleHTTP(t *testing.T) {
	privKey, err := generatePrivateKey(rsakey, 512)
	if err != nil {
		t.Errorf("Could not generate public key -> %v", err)
	}
	jws := &jws{privKey: privKey.(*rsa.PrivateKey)}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Replay-Nonce", "12345")
	}))

	solver := &simpleHTTPChallenge{jws: jws}
	clientChallenge := challenge{Type: "simpleHttp", Status: "pending", URI: ts.URL, Token: "123456789"}

	// validate error on non-root bind to 443
	if err = solver.Solve(clientChallenge, "test.domain"); err == nil {
		t.Error("BIND: Expected Solve to return an error but the error was nil.")
	}

	// Validate error on unexpected state
	solver.optPort = "23456"
	if err = solver.Solve(clientChallenge, "test.domain"); err == nil {
		t.Error("UNEXPECTED: Expected Solve to return an error but the error was nil.")
	}

	// Validate error on invalid status
	ts.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Replay-Nonce", "12345")
		failed := challenge{Type: "simpleHttp", Status: "invalid", URI: ts.URL, Token: "1234567810"}
		jsonBytes, _ := json.Marshal(&failed)
		w.Write(jsonBytes)
	})
	clientChallenge.Token = "1234567810"
	if err = solver.Solve(clientChallenge, "test.domain"); err == nil {
		t.Error("FAILED: Expected Solve to return an error but the error was nil.")
	}

	// Validate no error on valid response
	ts.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Replay-Nonce", "12345")
		valid := challenge{Type: "simpleHttp", Status: "valid", URI: ts.URL, Token: "1234567811"}
		jsonBytes, _ := json.Marshal(&valid)
		w.Write(jsonBytes)
	})
	clientChallenge.Token = "1234567811"
	if err = solver.Solve(clientChallenge, "test.domain"); err != nil {
		t.Errorf("VALID: Expected Solve to return no error but the error was -> %v", err)
	}

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
	if err = solver.Solve(clientChallenge, "test.domain"); err != nil {
		t.Errorf("VALID: Expected Solve to return no error but the error was -> %v", err)
	}
}
