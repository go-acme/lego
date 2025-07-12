package internal

import (
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupClient(server *httptest.Server) (*Client, error) {
	client := NewClient("secret")
	client.BaseURL = server.URL
	client.HTTPClient = server.Client()

	return client, nil
}

func TestClient_AddRecord(t *testing.T) {
	client := servermock.NewBuilder[*Client](setupClient).
		Route("GET /", servermock.RawStringResponse(`{}`),
			servermock.CheckQueryParameter().Strict().
				With("cmd", "dns-add_record").
				With("comment", "Managed+By+lego").
				With("format", "json").
				With("key", "secret").
				With("record", "example.com").
				With("type", "TXT").
				With("value", "aaa")).
		Build(t)

	err := client.AddRecord(t.Context(), "example.com", "aaa")
	require.NoError(t, err)
}

func TestClient_RemoveRecord(t *testing.T) {
	client := servermock.NewBuilder[*Client](setupClient).
		Route("GET /", servermock.RawStringResponse(`{}`),
			servermock.CheckQueryParameter().Strict().
				With("cmd", "dns-remove_record").
				With("comment", "Managed+By+lego").
				With("format", "json").
				With("key", "secret").
				With("record", "example.com").
				With("type", "TXT").
				With("value", "aaa")).
		Build(t)

	err := client.RemoveRecord(t.Context(), "example.com", "aaa")
	require.NoError(t, err)
}

func TestClient_buildQuery(t *testing.T) {
	const fakeAPIKey = "asdf1234"

	testCases := []struct {
		desc     string
		apiKey   string
		baseURL  string
		action   string
		domain   string
		txt      string
		expected string
	}{
		{
			desc:     "success",
			apiKey:   fakeAPIKey,
			action:   cmdAddRecord,
			domain:   "domain",
			txt:      "TXTtxtTXT",
			expected: "https://api.dreamhost.com?cmd=dns-add_record&comment=Managed%2BBy%2Blego&format=json&key=asdf1234&record=domain&type=TXT&value=TXTtxtTXT",
		},
		{
			desc:    "Invalid base URL",
			apiKey:  fakeAPIKey,
			baseURL: ":",
			action:  cmdAddRecord,
			domain:  "domain",
			txt:     "TXTtxtTXT",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			client := NewClient(test.apiKey)
			if test.baseURL != "" {
				client.BaseURL = test.baseURL
			}

			endpoint, err := client.buildEndpoint(test.action, test.domain, test.txt)

			if test.expected == "" {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expected, endpoint.String())
			}
		})
	}
}
