package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const fakeAPIKey = "asdf1234"

func TestClient_buildQuery(t *testing.T) {
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
