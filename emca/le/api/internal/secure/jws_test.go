package secure

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/xenolf/lego/emca/le"
	"github.com/xenolf/lego/emca/le/api/internal/sender"
)

func TestNotHoldingLockWhileMakingHTTPRequests(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(250 * time.Millisecond)
		w.Header().Add("Replay-Nonce", "12345")
		w.Header().Add("Retry-After", "0")
		err := writeJSONResponse(w, &le.Challenge{Type: "http-01", Status: "Valid", URL: "http://example.com/", Token: "token"})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}))
	defer ts.Close()

	do := sender.NewDo(http.DefaultClient, "lego-test")
	j := NewNonceManager(do, ts.URL)
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

// writeJSONResponse marshals the body as JSON and writes it to the response.
func writeJSONResponse(w http.ResponseWriter, body interface{}) error {
	bs, err := json.Marshal(body)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(bs); err != nil {
		return err
	}

	return nil
}
