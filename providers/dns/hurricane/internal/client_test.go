package internal

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_UpdateTxtRecord(t *testing.T) {
	testCases := []struct {
		code     string
		expected assert.ErrorAssertionFunc
	}{
		{
			code:     codeGood,
			expected: assert.NoError,
		},
		{
			code:     codeNoChg + ` "0123456789abcdef"`,
			expected: assert.NoError,
		},
		{
			code:     codeAbuse,
			expected: assert.Error,
		},
		{
			code:     codeBadAgent,
			expected: assert.Error,
		},
		{
			code:     codeBadAuth,
			expected: assert.Error,
		},
		{
			code:     codeNoHost,
			expected: assert.Error,
		},
		{
			code:     codeNotFqdn,
			expected: assert.Error,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.code, func(t *testing.T) {
			t.Parallel()

			handler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				if req.Method != http.MethodPost {
					http.Error(rw, fmt.Sprintf("unsupported method: %s", req.Method), http.StatusMethodNotAllowed)
					return
				}

				if err := req.ParseForm(); err != nil {
					http.Error(rw, "failed to parse form data", http.StatusBadRequest)
					return
				}

				if req.PostForm.Encode() != "hostname=_acme-challenge.example.com&password=secret&txt=foo" {
					http.Error(rw, "invalid form data", http.StatusBadRequest)
					return
				}

				_, _ = rw.Write([]byte(test.code))
			})

			server := httptest.NewServer(handler)
			t.Cleanup(server.Close)

			client := NewClient(map[string]string{"example.com": "secret"})
			client.baseURL = server.URL

			err := client.UpdateTxtRecord(context.Background(), "_acme-challenge.example.com", "foo")
			test.expected(t, err)
		})
	}
}
