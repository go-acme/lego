package internal

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func setup(t *testing.T, body string) *Client {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	mux.HandleFunc("/host", func(rw http.ResponseWriter, req *http.Request) {
		_, _ = fmt.Fprintln(rw, body)
	})

	client, err := NewClient("foo", "secret")
	require.NoError(t, err)

	client.baseURL = server.URL

	return client
}

func TestClient_SetRecord(t *testing.T) {
	testCases := []struct {
		desc     string
		response string
		assert   require.ErrorAssertionFunc
	}{
		{
			desc:     "success",
			response: `{"message":"OK","result":200}`,
			assert:   require.NoError,
		},
		{
			desc:     "failure",
			response: `{"message":"Not Found :  the information you requested was not found.","result":404}`,
			assert:   require.Error,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			client := setup(t, test.response)

			err := client.SetRecord("example.com", "txttxttxt", 10)
			test.assert(t, err)
		})
	}
}
