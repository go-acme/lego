package nonces

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-acme/lego/v4/acme"
	"github.com/go-acme/lego/v4/acme/api/internal/sender"
	"github.com/go-acme/lego/v4/platform/tester/servermock"
)

func TestNotHoldingLockWhileMakingHTTPRequests(t *testing.T) {
	manager, _ := servermock.NewBuilder(
		func(server *httptest.Server) (*Manager, error) {
			doer := sender.NewDoer(server.Client(), "lego-test")

			return NewManager(doer, server.URL), nil
		}).
		Route("HEAD /", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			time.Sleep(250 * time.Millisecond)

			rw.Header().Set("Replay-Nonce", "12345")
			rw.Header().Set("Retry-After", "0")

			servermock.JSONEncode(&acme.Challenge{Type: "http-01", Status: "Valid", URL: "https://example.com/", Token: "token"}).ServeHTTP(rw, req)
		})).
		BuildHTTPS(t)

	ch := make(chan bool)
	resultCh := make(chan bool)
	go func() {
		_, errN := manager.Nonce()
		if errN != nil {
			t.Log(errN)
		}
		ch <- true
	}()
	go func() {
		_, errN := manager.Nonce()
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
	case <-time.After(500 * time.Millisecond):
		t.Fatal("JWS is probably holding a lock while making HTTP request")
	}
}
