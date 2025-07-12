package internal

import (
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/require"
)

func setupClient(server *httptest.Server) (*Client, error) {
	client, err := NewClient("foo", "secret")
	if err != nil {
		return nil, err
	}

	client.baseURL = server.URL
	client.HTTPClient = server.Client()

	return client, nil
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
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			client := servermock.NewBuilder[*Client](setupClient, servermock.CheckHeader().WithJSONHeaders()).
				Route("PUT /host",
					servermock.RawStringResponse(test.response),
					servermock.CheckRequestJSONBody(`{"userid":"foo","apikey":"secret","hostname":"example.com","value":"txttxttxt","ttl":10,"type":"TXT"}`)).
				Build(t)

			err := client.SetRecord(t.Context(), "example.com", "txttxttxt", 10)
			test.assert(t, err)
		})
	}
}
