package vinyldns

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T) (*http.ServeMux, *DNSProvider) {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	config := NewDefaultConfig()
	config.AccessKey = "foo"
	config.SecretKey = "bar"
	config.Host = server.URL

	p, err := NewDNSProviderConfig(config)
	require.NoError(t, err)

	return mux, p
}

type mockRouter struct {
	debug bool

	mu     sync.Mutex
	routes map[string]map[string]http.HandlerFunc
}

func newMockRouter() *mockRouter {
	routes := map[string]map[string]http.HandlerFunc{
		http.MethodGet:    {},
		http.MethodPost:   {},
		http.MethodPut:    {},
		http.MethodDelete: {},
	}

	return &mockRouter{
		routes: routes,
	}
}

func (h *mockRouter) Debug() *mockRouter {
	h.debug = true

	return h
}

func (h *mockRouter) Get(path string, statusCode int, filename string) *mockRouter {
	h.add(http.MethodGet, path, statusCode, filename)
	return h
}

func (h *mockRouter) Post(path string, statusCode int, filename string) *mockRouter {
	h.add(http.MethodPost, path, statusCode, filename)
	return h
}

func (h *mockRouter) Put(path string, statusCode int, filename string) *mockRouter {
	h.add(http.MethodPut, path, statusCode, filename)
	return h
}

func (h *mockRouter) Delete(path string, statusCode int, filename string) *mockRouter {
	h.add(http.MethodDelete, path, statusCode, filename)
	return h
}

func (h *mockRouter) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.debug {
		fmt.Println(req)
	}

	rt := h.routes[req.Method]
	if rt == nil {
		http.NotFound(rw, req)
		return
	}

	hdl := rt[req.URL.Path]
	if hdl == nil {
		http.NotFound(rw, req)
		return
	}

	hdl(rw, req)
}

func (h *mockRouter) add(method, path string, statusCode int, filename string) {
	h.routes[method][path] = func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(statusCode)

		data, err := os.ReadFile(fmt.Sprintf("./fixtures/%s.json", filename))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		rw.Header().Set("Content-Type", "application/json")
		_, _ = rw.Write(data)
	}
}
