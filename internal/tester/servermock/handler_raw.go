package servermock

import (
	"net/http"
	"slices"
)

// RawResponseHandler is a custom HTTP handler that serves raw response data.
type RawResponseHandler struct {
	statusCode int
	headers    http.Header
	data       []byte
}

func RawResponse(data []byte) *RawResponseHandler {
	return &RawResponseHandler{
		statusCode: http.StatusOK,
		headers:    http.Header{},
		data:       data,
	}
}

func RawStringResponse(data string) *RawResponseHandler {
	return RawResponse([]byte(data))
}

func (h *RawResponseHandler) ServeHTTP(rw http.ResponseWriter, _ *http.Request) {
	for k, values := range h.headers {
		for _, v := range values {
			rw.Header().Add(k, v)
		}
	}

	rw.WriteHeader(h.statusCode)

	if len(h.data) == 0 {
		return
	}

	_, err := rw.Write(h.data)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *RawResponseHandler) WithStatusCode(status int) *RawResponseHandler {
	if h.statusCode >= http.StatusContinue {
		h.statusCode = status
	}

	return h
}

func (h *RawResponseHandler) WithHeader(name, value string, values ...string) *RawResponseHandler {
	for _, v := range slices.Concat([]string{value}, values) {
		h.headers.Add(name, v)
	}

	return h
}
