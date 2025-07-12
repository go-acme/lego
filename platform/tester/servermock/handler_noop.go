package servermock

import (
	"net/http"
	"slices"
)

// NoopHandler is a simple HTTP handler that responds without processing requests.
type NoopHandler struct {
	statusCode int
	headers    http.Header
}

func Noop() *NoopHandler {
	return &NoopHandler{
		statusCode: http.StatusOK,
		headers:    http.Header{},
	}
}

func (h *NoopHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	for k, values := range h.headers {
		for _, v := range values {
			rw.Header().Add(k, v)
		}
	}

	rw.WriteHeader(h.statusCode)
}

func (h *NoopHandler) WithStatusCode(status int) *NoopHandler {
	if h.statusCode >= http.StatusContinue {
		h.statusCode = status
	}

	return h
}

func (h *NoopHandler) WithHeader(name, value string, values ...string) *NoopHandler {
	for _, v := range slices.Concat([]string{value}, values) {
		h.headers.Add(name, v)
	}

	return h
}
