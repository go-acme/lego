package internal

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/require"
)

func setupClient(credentials map[string]string) func(server *httptest.Server) (*Client, error) {
	return func(server *httptest.Server) (*Client, error) {
		client := NewClient(credentials)
		client.HTTPClient = server.Client()
		client.baseURL = server.URL

		return client, nil
	}
}

func TestClient_Add(t *testing.T) {
	txtValue := "123456789012"

	client := servermock.NewBuilder[*Client](setupClient(map[string]string{"example.org": "secret"})).
		Route("POST /",
			servermock.RawStringResponse(fmt.Sprintf("%s %s", successCode, txtValue)),
			servermock.CheckQueryParameter().Strict().
				With("acme", addAction).With("txt", txtValue)).
		Build(t)

	err := client.Add(t.Context(), "example.org", txtValue)
	require.NoError(t, err)
}

func TestClient_Add_error(t *testing.T) {
	txtValue := "123456789012"

	client := servermock.NewBuilder[*Client](setupClient(map[string]string{"example.com": "secret"})).
		Route("POST /",
			servermock.RawStringResponse(fmt.Sprintf("%s %s", successCode, txtValue)),
			servermock.CheckQueryParameter().Strict().
				With("acme", addAction).With("txt", txtValue)).
		Build(t)

	err := client.Add(t.Context(), "example.org", txtValue)

	require.EqualError(t, err, "domain example.org not found in credentials, check your credentials map")
}

func TestClient_Remove(t *testing.T) {
	txtValue := "ABCDEFGHIJKL"

	client := servermock.NewBuilder[*Client](setupClient(map[string]string{"example.org": "secret"})).
		Route("POST /",
			servermock.RawStringResponse(fmt.Sprintf("%s %s", successCode, txtValue)),
			servermock.CheckQueryParameter().Strict().
				With("acme", removeAction).With("txt", txtValue)).
		Build(t)

	err := client.Remove(t.Context(), "example.org", txtValue)
	require.NoError(t, err)
}

func TestClient_Remove_error(t *testing.T) {
	txtValue := "ABCDEFGHIJKL"

	testCases := []struct {
		desc     string
		hostname string
		response string
		expected string
	}{
		{
			desc:     "response error - txt",
			hostname: "example.com",
			response: "error - no valid acme txt record",
			expected: "error - no valid acme txt record",
		},
		{
			desc:     "response error - acme",
			hostname: "example.com",
			response: "nochg 1234:1234:1234:1234:1234:1234:1234:1234",
			expected: "nochg 1234:1234:1234:1234:1234:1234:1234:1234",
		},
		{
			desc:     "credential error",
			hostname: "example.org",
			response: fmt.Sprintf("%s %s", successCode, txtValue),
			expected: "domain example.org not found in credentials, check your credentials map",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			client := servermock.NewBuilder[*Client](setupClient(map[string]string{"example.com": "secret"})).
				Route("POST /",
					servermock.RawStringResponse(test.response),
					servermock.CheckQueryParameter().Strict().
						With("acme", removeAction).With("txt", txtValue)).
				Build(t)

			err := client.Remove(t.Context(), test.hostname, txtValue)
			require.EqualError(t, err, test.expected)
		})
	}
}
