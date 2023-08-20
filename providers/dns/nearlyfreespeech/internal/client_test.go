package internal

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T) (*Client, *http.ServeMux) {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := NewClient("user", "secret")
	client.HTTPClient = server.Client()
	client.baseURL, _ = url.Parse(server.URL)
	// Make everything deterministic for golden tests of signatures.
	client.genSalt = func() []byte { return []byte("0123456789ABCDEF") }
	client.timeNow = func() time.Time { return time.Unix(1692475113, 0) }

	return client, mux
}

func testHandler(params map[string]string, authHeader string) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(rw, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		if req.Header.Get(authenticationHeader) != authHeader {
			// As an aid in fixing the test if the exact formatting of the body is changed.
			fmt.Printf("authHeader: got %q wanted %q, returning 403\n", req.Header.Get(authenticationHeader), authHeader)
			http.Error(rw, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		err := req.ParseForm()
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		for k, v := range params {
			if req.PostForm.Get(k) != v {
				http.Error(rw, fmt.Sprintf("data: got %s want %s", k, v), http.StatusBadRequest)
				return
			}
		}
	}
}

func testErrorHandler() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(rw, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		file, err := os.Open("./fixtures/error.json")
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		rw.WriteHeader(http.StatusUnauthorized)

		_, _ = io.Copy(rw, file)
	}
}

func TestClient_AddRecord(t *testing.T) {
	client, mux := setupTest(t)

	params := map[string]string{
		"data": "txtTXTtxt",
		"name": "sub",
		"type": "TXT",
		"ttl":  "30",
	}

	// The reason we're testing that we're getting this exact authHeader is
	// so that we can detect if some change that is not expected to modify
	// neither the request nor the authHeader does so. If you're changing
	// the request, please ensure your change actually works against the
	// real nfsn API before modifying this golden value.
	mux.Handle("/dns/example.com/addRR", testHandler(params, "user;1692475113;0123456789ABCDEF;24a32faf74c7bd0525f560ff12a1c1fb6545bafc"))

	record := Record{
		Name: "sub",
		Type: "TXT",
		Data: "txtTXTtxt",
		TTL:  30,
	}

	err := client.AddRecord(context.Background(), "example.com", record)
	require.NoError(t, err)
}

func TestClient_AddRecord_error(t *testing.T) {
	client, mux := setupTest(t)

	mux.Handle("/dns/example.com/addRR", testErrorHandler())

	record := Record{
		Name: "sub",
		Type: "TXT",
		Data: "txtTXTtxt",
		TTL:  30,
	}

	err := client.AddRecord(context.Background(), "example.com", record)
	require.Error(t, err)
}

func TestClient_RemoveRecord(t *testing.T) {
	client, mux := setupTest(t)

	params := map[string]string{
		"data": "txtTXTtxt",
		"name": "sub",
		"type": "TXT",
	}

	// The reason we're testing that we're getting this exact authHeader is
	// so that we can detect if some change that is not expected to modify
	// neither the request nor the authHeader does so. If you're changing
	// the request, please ensure your change actually works against the
	// real nfsn API before modifying this golden value.
	mux.Handle("/dns/example.com/removeRR", testHandler(params, "user;1692475113;0123456789ABCDEF;699f01f077ca487bd66ac370d6dfc5b122c65522"))

	record := Record{
		Name: "sub",
		Type: "TXT",
		Data: "txtTXTtxt",
	}

	err := client.RemoveRecord(context.Background(), "example.com", record)
	require.NoError(t, err)
}

func TestClient_RemoveRecord_error(t *testing.T) {
	client, mux := setupTest(t)

	mux.Handle("/dns/example.com/removeRR", testErrorHandler())

	record := Record{
		Name: "sub",
		Type: "TXT",
		Data: "txtTXTtxt",
	}

	err := client.RemoveRecord(context.Background(), "example.com", record)
	require.Error(t, err)
}
