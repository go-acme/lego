package servermock

import (
	"encoding/json"
	"net/http"
)

// JSONEncodeHandler is a handler that encodes data into JSON and writes it to an HTTP response.
type JSONEncodeHandler struct {
	data       any
	statusCode int
}

func JSONEncode(data any) *JSONEncodeHandler {
	return &JSONEncodeHandler{
		data:       data,
		statusCode: http.StatusOK,
	}
}

func (h *JSONEncodeHandler) ServeHTTP(rw http.ResponseWriter, _ *http.Request) {
	rw.Header().Set(contentTypeHeader, applicationJSONMimeType)

	rw.WriteHeader(h.statusCode)

	err := json.NewEncoder(rw).Encode(h.data)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *JSONEncodeHandler) WithStatusCode(status int) *JSONEncodeHandler {
	if h.statusCode >= http.StatusContinue {
		h.statusCode = status
	}

	return h
}
