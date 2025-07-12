package servermock

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
)

// ResponseFromFileHandler handles HTTP responses using the content of a file.
type ResponseFromFileHandler struct {
	statusCode int
	headers    http.Header
	filename   string
}

func ResponseFromFile(filename string) *ResponseFromFileHandler {
	return &ResponseFromFileHandler{
		statusCode: http.StatusOK,
		headers:    http.Header{},
		filename:   filename,
	}
}

func ResponseFromFixture(filename string) *ResponseFromFileHandler {
	return ResponseFromFile(filepath.Join("fixtures", filename))
}

func (h *ResponseFromFileHandler) ServeHTTP(rw http.ResponseWriter, _ *http.Request) {
	for k, values := range h.headers {
		for _, v := range values {
			rw.Header().Add(k, v)
		}
	}

	if h.filename == "" {
		rw.WriteHeader(h.statusCode)
		return
	}

	if filepath.Ext(h.filename) == ".json" {
		rw.Header().Set(contentTypeHeader, applicationJSONMimeType)
	}

	file, err := os.Open(h.filename)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	defer func() { _ = file.Close() }()

	rw.WriteHeader(h.statusCode)

	_, err = io.Copy(rw, file)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *ResponseFromFileHandler) WithStatusCode(status int) *ResponseFromFileHandler {
	if h.statusCode >= http.StatusContinue {
		h.statusCode = status
	}

	return h
}

func (h *ResponseFromFileHandler) WithHeader(name, value string, values ...string) *ResponseFromFileHandler {
	for _, v := range slices.Concat([]string{value}, values) {
		h.headers.Add(name, v)
	}

	return h
}
