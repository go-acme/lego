package internal

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T, method, pattern string, status int, file string) *Client {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	mux.HandleFunc(pattern, func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != method {
			http.Error(rw, fmt.Sprintf("unsupported method: %s", req.Method), http.StatusBadRequest)
			return
		}

		apiToken := req.Header.Get("Authorization")
		if apiToken != "Bearer secret" {
			http.Error(rw, fmt.Sprintf("invalid credentials: %s", apiToken), http.StatusBadRequest)
			return
		}

		if file == "" {
			rw.WriteHeader(status)
			return
		}

		open, err := os.Open(filepath.Join("fixtures", file))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		defer func() { _ = open.Close() }()

		rw.WriteHeader(status)
		_, err = io.Copy(rw, open)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	serverURL, _ := url.Parse(server.URL)

	client := NewClient(serverURL, "secret")
	client.HTTPClient = server.Client()

	return client
}

func TestClient_joinPath(t *testing.T) {
	testCases := []struct {
		desc     string
		baseURL  string
		uri      string
		expected string
	}{
		{
			desc:     "host with path",
			baseURL:  "https://example.com/test",
			uri:      "/foo",
			expected: "https://example.com/test/foo",
		},
		{
			desc:     "host with path + trailing slash",
			baseURL:  "https://example.com/test/",
			uri:      "/foo",
			expected: "https://example.com/test/foo",
		},
		{
			desc:     "no URI",
			baseURL:  "https://example.com/test",
			uri:      "",
			expected: "https://example.com/test",
		},
		{
			desc:     "host without path",
			baseURL:  "https://example.com",
			uri:      "/foo",
			expected: "https://example.com/foo",
		},
		{
			desc:     "api",
			baseURL:  "https://example.com",
			uri:      "/api",
			expected: "https://example.com/api",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			host, err := url.Parse(test.baseURL)
			require.NoError(t, err)

			client := NewClient(host, "secret")

			endpoint := client.joinPath(test.uri)

			assert.Equal(t, test.expected, endpoint.String())
		})
	}
}

func TestClient_UpdateRecords_error(t *testing.T) {
	client := setupTest(t, http.MethodPatch, "/zones/example.org/rrsets", http.StatusUnprocessableEntity, "error.json")

	rrSet := []UpdateRRSet{
		{
			Name:       "acme.example.org.",
			ChangeType: "add",
			Type:       "TXT",
			Records: []Record{{
				Content: "\"my-acme-challenge\"",
			}},
		},
	}

	err := client.UpdateRecords(context.Background(), "example.org", rrSet)
	require.ErrorAs(t, err, &apiResponse{})
}

func TestClient_UpdateRecords(t *testing.T) {
	client := setupTest(t, http.MethodPatch, "/zones/example.org/rrsets", http.StatusOK, "rrsets-response.json")

	rrSet := []UpdateRRSet{
		{
			Name:       "acme.example.org.",
			ChangeType: "add",
			Type:       "TXT",
			Records: []Record{{
				Content: "\"my-acme-challenge\"",
			}},
		},
	}

	err := client.UpdateRecords(context.Background(), "example.org", rrSet)
	require.NoError(t, err)
}
