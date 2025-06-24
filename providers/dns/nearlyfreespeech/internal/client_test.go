package internal

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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

	client.signer.saltShaker = func() []byte { return []byte("0123456789ABCDEF") }
	client.signer.clock = func() time.Time { return time.Unix(1692475113, 0) }

	return client, mux
}

func testHandler(params map[string]string) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(rw, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		if req.Header.Get(authenticationHeader) == "" {
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

	mux.Handle("/dns/example.com/addRR", testHandler(params))

	record := Record{
		Name: "sub",
		Type: "TXT",
		Data: "txtTXTtxt",
		TTL:  30,
	}

	err := client.AddRecord(t.Context(), "example.com", record)
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

	err := client.AddRecord(t.Context(), "example.com", record)
	require.Error(t, err)
}

func TestClient_RemoveRecord(t *testing.T) {
	client, mux := setupTest(t)

	params := map[string]string{
		"data": "txtTXTtxt",
		"name": "sub",
		"type": "TXT",
	}

	mux.Handle("/dns/example.com/removeRR", testHandler(params))

	record := Record{
		Name: "sub",
		Type: "TXT",
		Data: "txtTXTtxt",
	}

	err := client.RemoveRecord(t.Context(), "example.com", record)
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

	err := client.RemoveRecord(t.Context(), "example.com", record)
	require.Error(t, err)
}

func TestSigner_Sign(t *testing.T) {
	testCases := []struct {
		desc     string
		path     string
		now      int64
		salt     string
		expected string
	}{
		{
			desc:     "basic",
			path:     "/path",
			now:      1692475113,
			salt:     "0123456789ABCDEF",
			expected: "user;1692475113;0123456789ABCDEF;417a9988c7ad7919b297884dd120b5808d8a1e6f",
		},
		{
			desc:     "another date",
			path:     "/path",
			now:      1692567766,
			salt:     "0123456789ABCDEF",
			expected: "user;1692567766;0123456789ABCDEF;b5c28286fd2e1a45a7c576dc2a6430116f721502",
		},
		{
			desc:     "another salt",
			path:     "/path",
			now:      1692475113,
			salt:     "FEDCBA9876543210",
			expected: "user;1692475113;FEDCBA9876543210;0f766822bda4fdc09829be4e1ea5e27ae3ae334e",
		},
		{
			desc:     "empty path",
			path:     "",
			now:      1692475113,
			salt:     "0123456789ABCDEF",
			expected: "user;1692475113;0123456789ABCDEF;c7c241a4d15d04d92805631d58d4d72ac1c339a1",
		},
		{
			desc:     "root path",
			path:     "/",
			now:      1692475113,
			salt:     "0123456789ABCDEF",
			expected: "user;1692475113;0123456789ABCDEF;c7c241a4d15d04d92805631d58d4d72ac1c339a1",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			signer := NewSigner()
			signer.saltShaker = func() []byte { return []byte(test.salt) }
			signer.clock = func() time.Time { return time.Unix(test.now, 0) }

			sign := signer.Sign(test.path, "data", "user", "secret")

			assert.Equal(t, test.expected, sign)
		})
	}
}
