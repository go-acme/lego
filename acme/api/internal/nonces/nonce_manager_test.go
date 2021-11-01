package nonces

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-acme/lego/v4/acme"
	"github.com/go-acme/lego/v4/acme/api/internal/sender"
	"github.com/go-acme/lego/v4/platform/tester"
)

func TestNotHoldingLockWhileMakingHTTPRequests(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(250 * time.Millisecond)
		w.Header().Set("Replay-Nonce", "12345")
		w.Header().Set("Retry-After", "0")
		err := tester.WriteJSONResponse(w, &acme.Challenge{Type: "http-01", Status: "Valid", URL: "http://example.com/", Token: "token"})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}))
	t.Cleanup(server.Close)

	doer := sender.NewDoer(http.DefaultClient, "lego-test")
	j := NewManager(doer, server.URL)
	ch := make(chan bool)
	resultCh := make(chan bool)
	go func() {
		_, errN := j.Nonce()
		if errN != nil {
			t.Log(errN)
		}
		ch <- true
	}()
	go func() {
		_, errN := j.Nonce()
		if errN != nil {
			t.Log(errN)
		}
		ch <- true
	}()
	go func() {
		<-ch
		<-ch
		resultCh <- true
	}()
	select {
	case <-resultCh:
	case <-time.After(400 * time.Millisecond):
		t.Fatal("JWS is probably holding a lock while making HTTP request")
	}
}
