package servermock

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
)

// RequestBodyLink represents a handler utility to validate HTTP request bodies against a predefined byte slice.
type RequestBodyLink struct {
	body             []byte
	filename         string
	ignoreWhitespace bool
}

// CheckRequestBody creates a [RequestBodyLink] initialized with the provided request body string.
func CheckRequestBody(body string) *RequestBodyLink {
	return &RequestBodyLink{body: []byte(body)}
}

// CheckRequestBodyFromFile creates a [RequestBodyLink] initialized with the provided request body file.
func CheckRequestBodyFromFile(filename string) *RequestBodyLink {
	return &RequestBodyLink{filename: filename}
}

// CheckRequestBodyFromFixture creates a [RequestBodyLink] initialized with the provided request body file from the `fixtures` directory.
func CheckRequestBodyFromFixture(filename string) *RequestBodyLink {
	return CheckRequestBodyFromFile(filepath.Join("fixtures", filename))
}

// CheckRequestBodyFromInternal creates a [RequestBodyLink] initialized with the provided request body file from the `internal/fixtures directory.
func CheckRequestBodyFromInternal(filename string) *RequestBodyLink {
	return CheckRequestBodyFromFile(filepath.Join("internal", "fixtures", filename))
}

func (l *RequestBodyLink) Bind(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.ContentLength == 0 {
			http.Error(rw, fmt.Sprintf("%s: empty request body", req.URL.Path), http.StatusBadRequest)
			return
		}

		body, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		_ = req.Body.Close()

		expectedRaw := slices.Clone(l.body)

		if l.filename != "" {
			expectedRaw, err = os.ReadFile(l.filename)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		if len(expectedRaw) == 0 {
			http.Error(rw, fmt.Sprintf("%s: empty expected request body", req.URL.Path), http.StatusBadRequest)
			return
		}

		if l.ignoreWhitespace {
			body = trimLineSpace(body)
			expectedRaw = trimLineSpace(expectedRaw)
		}

		if !bytes.Equal(bytes.TrimSpace(expectedRaw), bytes.TrimSpace(body)) {
			msg := fmt.Sprintf("%s: request body differences: got: %s, want: %s", req.URL.Path,
				string(bytes.TrimSpace(body)), string(bytes.TrimSpace(expectedRaw)))
			http.Error(rw, msg, http.StatusBadRequest)
			return
		}

		next.ServeHTTP(rw, req)
	})
}

func (l *RequestBodyLink) IgnoreWhitespace() *RequestBodyLink {
	l.ignoreWhitespace = true

	return l
}

func trimLineSpace(body []byte) []byte {
	buf := bytes.NewBuffer(nil)
	for line := range bytes.Lines(body) {
		buf.Write(bytes.TrimSpace(line))
	}

	return buf.Bytes()
}
