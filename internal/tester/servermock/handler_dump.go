package servermock

import (
	"fmt"
	"net/http"
	"net/http/httputil"
)

// DumpRequest logs the full HTTP request to the console, including the body if present.
func DumpRequest() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		dump, err := httputil.DumpRequest(req, true)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Println(string(dump))
	}
}
