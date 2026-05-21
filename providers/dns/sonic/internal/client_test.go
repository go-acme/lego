package internal

import (
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/require"
)

func setupClient(server *httptest.Server) (*Client, error) {
	client, err := NewClient("user", "secret")
	if err != nil {
		return nil, err
	}

	client.BaseURL = server.URL
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
			response: "set_record.json",
			assert:   require.NoError,
		},
		{
			desc:     "failure",
			response: "error.json",
			assert:   require.Error,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			client := servermock.NewBuilder[*Client](setupClient, servermock.CheckHeader().WithJSONHeaders()).
				Route("PUT /host",
					servermock.ResponseFromFixture(test.response),
					servermock.CheckRequestJSONBodyFromFixture("set_record_add-request.json"),
				).
				Build(t)

			err := client.SetRecord(t.Context(), "_acme-challenge.example.com", "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY", 120)
			test.assert(t, err)
		})
	}
}
