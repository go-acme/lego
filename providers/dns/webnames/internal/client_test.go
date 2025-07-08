package internal

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T, filename string, expectedParams url.Values) *Client {
	t.Helper()

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(rw, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		if req.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
			http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		err := req.ParseForm()
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		for k, v := range expectedParams {
			val := req.PostForm.Get(k)
			if len(v) == 0 {
				http.Error(rw, fmt.Sprintf("%s: no value", k), http.StatusBadRequest)
				return
			}

			if val != v[0] {
				http.Error(rw, fmt.Sprintf("%s: invalid value: %s != %s", k, val, v[0]), http.StatusBadRequest)
				return
			}
		}

		file, err := os.Open(path.Join("fixtures", filename))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		defer func() { _ = file.Close() }()

		_, err = io.Copy(rw, file)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	server := httptest.NewServer(mux)

	client := NewClient("secret")
	client.baseURL = server.URL
	client.HTTPClient = server.Client()

	return client
}

func TestClient_AddTXTRecord(t *testing.T) {
	testCases := []struct {
		desc     string
		filename string
		require  require.ErrorAssertionFunc
	}{
		{
			desc:     "ok",
			filename: "ok.json",
			require:  require.NoError,
		},
		{
			desc:     "error",
			filename: "error.json",
			require:  require.Error,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			data := url.Values{}
			data.Set("domain", "example.com")
			data.Set("type", "TXT")
			data.Set("record", "foo:txtTXTtxt")
			data.Set("action", "add")

			client := setupTest(t, test.filename, data)

			domain := "example.com"
			subDomain := "foo"
			content := "txtTXTtxt"

			err := client.AddTXTRecord(t.Context(), domain, subDomain, content)
			test.require(t, err)
		})
	}
}

func TestClient_RemoveTxtRecord(t *testing.T) {
	testCases := []struct {
		desc     string
		filename string
		require  require.ErrorAssertionFunc
	}{
		{
			desc:     "ok",
			filename: "ok.json",
			require:  require.NoError,
		},
		{
			desc:     "error",
			filename: "error.json",
			require:  require.Error,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			data := url.Values{}
			data.Set("domain", "example.com")
			data.Set("type", "TXT")
			data.Set("record", "foo:txtTXTtxt")
			data.Set("action", "delete")

			client := setupTest(t, test.filename, data)

			domain := "example.com"
			subDomain := "foo"
			content := "txtTXTtxt"

			err := client.RemoveTXTRecord(t.Context(), domain, subDomain, content)
			test.require(t, err)
		})
	}
}
