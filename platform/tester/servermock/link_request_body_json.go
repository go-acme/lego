package servermock

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"

	"github.com/google/go-cmp/cmp"
)

// RequestBodyJSONLink validates JSON request bodies.
type RequestBodyJSONLink struct {
	body     []byte
	filename string
	data     any
}

// CheckRequestJSONBody creates a [RequestBodyJSONLink] initialized with a string.
func CheckRequestJSONBody(body string) *RequestBodyJSONLink {
	return &RequestBodyJSONLink{body: []byte(body)}
}

// CheckRequestJSONBodyFromStruct creates a [RequestBodyJSONLink] initialized with a struct.
func CheckRequestJSONBodyFromStruct(data any) *RequestBodyJSONLink {
	return &RequestBodyJSONLink{data: data}
}

// CheckRequestJSONBodyFromFile creates a [RequestBodyJSONLink] initialized with the provided request body file.
func CheckRequestJSONBodyFromFile(filename string) *RequestBodyJSONLink {
	return &RequestBodyJSONLink{filename: filename}
}

func (l *RequestBodyJSONLink) Bind(next http.Handler) http.Handler {
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

		var expected, actual any

		expectedRaw := slices.Clone(l.body)

		switch {
		case l.filename != "":
			expectedRaw, err = os.ReadFile(filepath.Join("fixtures", l.filename))
			if err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				return
			}

		case l.data != nil:
			expectedRaw, err = json.Marshal(l.data)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		if len(expectedRaw) == 0 {
			http.Error(rw, fmt.Sprintf("%s: empty expected request body", req.URL.Path), http.StatusBadRequest)
			return
		}

		err = json.Unmarshal(expectedRaw, &expected)
		if err != nil {
			msg := fmt.Sprintf("%s: the expected request body is not valid JSON: %v", req.URL.Path, err)
			http.Error(rw, msg, http.StatusBadRequest)
			return
		}

		err = json.Unmarshal(body, &actual)
		if err != nil {
			msg := fmt.Sprintf("%s: request body is not valid JSON: %v", req.URL.Path, err)
			http.Error(rw, msg, http.StatusBadRequest)
			return
		}

		if !cmp.Equal(actual, expected) {
			msg := fmt.Sprintf("%s: request body differences: %s", req.URL.Path, cmp.Diff(actual, expected))
			http.Error(rw, msg, http.StatusBadRequest)
			return
		}

		next.ServeHTTP(rw, req)
	})
}
