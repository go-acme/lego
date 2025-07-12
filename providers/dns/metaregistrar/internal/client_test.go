package internal

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client, err := NewClient("secret")
			if err != nil {
				return nil, err
			}

			client.HTTPClient = server.Client()
			client.baseURL, _ = url.Parse(server.URL)

			return client, nil
		},
		servermock.CheckHeader().WithJSONHeaders().
			With(tokenHeader, "secret"))
}

func TestClient_UpdateDNSZone(t *testing.T) {
	client := mockBuilder().
		Route("PATCH /dnszone/example.com",
			servermock.ResponseFromFixture("update-dns-zone.json"),
			servermock.CheckRequestJSONBody(`{"add":[{"name":"@","type":"TXT","ttl":60,"content":"value"}]}`)).
		Build(t)

	updateRequest := DNSZoneUpdateRequest{
		Add: []Record{{
			Name:    "@",
			Type:    "TXT",
			TTL:     60,
			Content: "value",
		}},
	}

	response, err := client.UpdateDNSZone(t.Context(), "example.com", updateRequest)
	require.NoError(t, err)

	expected := &DNSZoneUpdateResponse{
		ResponseID: "mapi1_cb46ad8790b62b76535bd3102bd282aec83b894c",
		Status:     "ok",
		Message:    "Command completed successfully",
	}

	assert.Equal(t, expected, response)
}

func TestClient_UpdateDNSZone_error(t *testing.T) {
	testCases := []struct {
		desc     string
		filename string
		expected string
	}{
		{
			desc:     "authentication error",
			filename: "error.json",
			expected: "invalid_token: the supplied token is invalid",
		},
		{
			desc:     "API error",
			filename: "error-response.json",
			expected: "error: does_not_exist: This server does not exist",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			client := mockBuilder().
				Route("PATCH /dnszone/example.com",
					servermock.ResponseFromFixture(test.filename).
						WithStatusCode(http.StatusUnprocessableEntity)).
				Build(t)

			updateRequest := DNSZoneUpdateRequest{
				Add: []Record{{
					Name:    "@",
					Type:    "TXT",
					TTL:     60,
					Content: "value",
				}},
			}

			_, err := client.UpdateDNSZone(t.Context(), "example.com", updateRequest)
			require.EqualError(t, err, test.expected)
		})
	}
}
