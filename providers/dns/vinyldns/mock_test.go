package vinyldns

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// MockResponse represents a predefined response used by a mock server.
type MockResponse struct {
	StatusCode int
	Body       string
}

// MockResponseMap maps request paths to responses.
type MockResponseMap map[string]map[string]MockResponse

func newMockServer(t *testing.T, responses MockResponseMap) string {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.RequestURI
		method := r.Method
		resp, ok := responses[method][path]
		if !ok {
			resp, ok = responses[method][r.URL.Path]
			if !ok {
				msg := fmt.Sprintf("Requested path not found in response map: %s using method: %s", path, method)
				require.FailNow(t, msg)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		_, err := w.Write([]byte(resp.Body))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}))

	t.Cleanup(server.Close)

	time.Sleep(100 * time.Millisecond)

	return server.URL
}
