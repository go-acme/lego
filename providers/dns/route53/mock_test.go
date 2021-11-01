package route53

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
type MockResponseMap map[string]MockResponse

func setupTest(t *testing.T, responses MockResponseMap) string {
	t.Helper()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		resp, ok := responses[path]
		if !ok {
			resp, ok = responses[r.RequestURI]
			if !ok {
				msg := fmt.Sprintf("Requested path not found in response map: %s", path)
				require.FailNow(t, msg)
			}
		}

		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(resp.StatusCode)
		_, err := w.Write([]byte(resp.Body))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	time.Sleep(100 * time.Millisecond)

	return server.URL
}
