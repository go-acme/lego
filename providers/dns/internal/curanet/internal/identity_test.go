package internal

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

func TestGetRecords_oauthClient(t *testing.T) {
	oa := servermock.NewBuilder[*http.Client](
		func(server *httptest.Server) (*http.Client, error) {
			config := &clientcredentials.Config{
				TokenURL:     server.URL,
				ClientID:     "123",
				ClientSecret: "xxx",
				Scopes:       []string{"dns"},
				AuthStyle:    oauth2.AuthStyleInParams,
			}

			return config.Client(t.Context()), nil
		}).
		Route("POST /",
			servermock.ResponseFromFixture("oauth.json"),
			servermock.CheckForm().Strict().UsePostForm().
				With("client_id", "123").
				With("client_secret", "xxx").
				With("grant_type", "client_credentials").
				With("scope", "dns"),
		).
		Build(t)

	client := mockBuilderOAuth(oa).
		Route("GET /dns/v2/Domains/example.com/Records",
			servermock.ResponseFromFixture("records_get.json"),
			servermock.CheckQueryParameter().Strict().
				With("name", "_acme-challenge").
				With("type", "TXT"),
		).
		Build(t)

	records, err := client.GetRecords(t.Context(), "example.com", "_acme-challenge", "TXT")
	require.NoError(t, err)

	assert.NotNil(t, records)
}

func mockBuilderOAuth(oauthClient *http.Client) *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client, err := NewClient(oauthClient)
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
