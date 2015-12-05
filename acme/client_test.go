package acme

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewClient(t *testing.T) {
	keyBits := 32 // small value keeps test fast
	key, err := rsa.GenerateKey(rand.Reader, keyBits)
	if err != nil {
		t.Fatal("Could not generate test key:", err)
	}
	user := mockUser{
		email:      "test@test.com",
		regres:     new(RegistrationResource),
		privatekey: key,
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, _ := json.Marshal(directory{NewAuthzURL: "http://test", NewCertURL: "http://test", NewRegURL: "http://test", RevokeCertURL: "http://test"})
		w.Write(data)
	}))

	client, err := NewClient(ts.URL, user, keyBits, nil)
	if err != nil {
		t.Fatalf("Could not create client: %v", err)
	}

	if client.jws == nil {
		t.Fatalf("Expected client.jws to not be nil")
	}
	if expected, actual := key, client.jws.privKey; actual != expected {
		t.Errorf("Expected jws.privKey to be %p but was %p", expected, actual)
	}

	if client.keyBits != keyBits {
		t.Errorf("Expected keyBits to be %d but was %d", keyBits, client.keyBits)
	}

	if expected, actual := 2, len(client.solvers); actual != expected {
		t.Fatalf("Expected %d solver(s), got %d", expected, actual)
	}
}

func TestNewClientOptPort(t *testing.T) {
	keyBits := 32 // small value keeps test fast
	key, err := rsa.GenerateKey(rand.Reader, keyBits)
	if err != nil {
		t.Fatal("Could not generate test key:", err)
	}
	user := mockUser{
		email:      "test@test.com",
		regres:     new(RegistrationResource),
		privatekey: key,
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, _ := json.Marshal(directory{NewAuthzURL: "http://test", NewCertURL: "http://test", NewRegURL: "http://test", RevokeCertURL: "http://test"})
		w.Write(data)
	}))

	optPort := "1234"
	client, err := NewClient(ts.URL, user, keyBits, []string{"http-01:" + optPort})
	if err != nil {
		t.Fatalf("Could not create client: %v", err)
	}

	httpSolver, ok := client.solvers["http-01"].(*httpChallenge)
	if !ok {
		t.Fatal("Expected http-01 solver to be httpChallenge type")
	}
	if httpSolver.jws != client.jws {
		t.Error("Expected http-01 to have same jws as client")
	}
	if httpSolver.optPort != optPort {
		t.Errorf("Expected http-01 to have optPort %s but was %s", optPort, httpSolver.optPort)
	}
}

func TestValidate(t *testing.T) {
	// Disable polling delay in validate for faster tests.
	pollInterval = 0

	var statuses []string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Minimal stub ACME server for validation.
		w.Header().Add("Replay-Nonce", "12345")
		switch r.Method {
		case "HEAD":
		case "POST":
			st := statuses[0]
			statuses = statuses[1:]
			writeJSONResponse(w, &challenge{Type: "http-01", Status: st, URI: "http://example.com/", Token: "token"})

		case "GET":
			st := statuses[0]
			statuses = statuses[1:]
			writeJSONResponse(w, &challenge{Type: "http-01", Status: st, URI: "http://example.com/", Token: "token"})

		default:
			http.Error(w, r.Method, http.StatusMethodNotAllowed)
		}
	}))
	defer ts.Close()

	privKey, _ := generatePrivateKey(rsakey, 512)
	j := &jws{privKey: privKey.(*rsa.PrivateKey), directoryURL: ts.URL}

	tsts := []struct {
		name     string
		statuses []string
		want     string
	}{
		{"POST-unexpected", []string{"weird"}, "unexpected"},
		{"POST-valid", []string{"valid"}, ""},
		{"POST-invalid", []string{"invalid"}, "not validate"},
		{"GET-unexpected", []string{"pending", "weird"}, "unexpected"},
		{"GET-valid", []string{"pending", "valid"}, ""},
		{"GET-invalid", []string{"pending", "invalid"}, "not validate"},
	}

	for _, tst := range tsts {
		statuses = tst.statuses
		if err := validate(j, ts.URL, challenge{Type: "http-01", Token: "token"}); err == nil && tst.want != "" {
			t.Errorf("[%s] validate: got error %v, want something with %q", tst.name, err, tst.want)
		} else if err != nil && !strings.Contains(err.Error(), tst.want) {
			t.Errorf("[%s] validate: got error %v, want something with %q", tst.name, err, tst.want)
		}
	}
}

// writeJSONResponse marshals the body as JSON and writes it to the response.
func writeJSONResponse(w http.ResponseWriter, body interface{}) {
	bs, err := json.Marshal(body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(bs); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// stubValidate is like validate, except it does nothing.
func stubValidate(j *jws, uri string, chlng challenge) error {
	return nil
}

type mockUser struct {
	email      string
	regres     *RegistrationResource
	privatekey *rsa.PrivateKey
}

func (u mockUser) GetEmail() string                       { return u.email }
func (u mockUser) GetRegistration() *RegistrationResource { return u.regres }
func (u mockUser) GetPrivateKey() *rsa.PrivateKey         { return u.privatekey }
