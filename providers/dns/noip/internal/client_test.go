package internal

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client, err := NewClient(OAuthStaticAccessToken(server.Client(), "secret"))
			if err != nil {
				return nil, err
			}

			client.BaseURL, _ = url.Parse(server.URL)

			return client, nil
		},
		servermock.CheckHeader().
			WithJSONHeaders().
			WithAuthorization("Bearer secret"),
	)
}

func TestClient_CreateRData(t *testing.T) {
	client := mockBuilder().
		Route("POST /v1/dns/records/example.com/foo/rrsets/TXT/rdata",
			servermock.Noop(),
			servermock.CheckRequestJSONBody(`[{"value":"txtTXTtxt","label":"mylabel"}]`)).
		Build(t)

	data := RData{
		Value: "txtTXTtxt",
		Label: "mylabel",
	}

	err := client.CreateRData(t.Context(), "example.com", "foo", "TXT", data)
	require.NoError(t, err)
}

func TestClient_CreateRData_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /v1/dns/records/example.com/foo/rrsets/TXT/rdata",
			servermock.ResponseFromFixture("full.json").
				WithStatusCode(http.StatusUnauthorized)).
		Build(t)

	data := RData{
		Value: "txtTXTtxt",
		Label: "mylabel",
	}

	err := client.CreateRData(t.Context(), "example.com", "foo", "TXT", data)
	require.EqualError(t, err, "[status code 401] id: err_eXgLfYUj, code: 2451, title: invalid request query, detail: unknown variant `summary`, expected `detail` or `name_only`, location: query; id: err_oWfehpGt, code: 3809, title: invalid url encoded request body")
}

func TestClient_DeleteRDataByLabel(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /v1/dns/records/example.com/foo/rrsets/TXT/rdata/mylabel",
			servermock.Noop()).
		Build(t)

	err := client.DeleteRDataByLabel(t.Context(), "example.com", "foo", "TXT", "mylabel")
	require.NoError(t, err)
}
