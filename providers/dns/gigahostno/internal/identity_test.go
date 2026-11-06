package internal

import (
	"context"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func mockTokenSource(ctx context.Context) *servermock.Builder[oauth2.TokenSource] {
	return servermock.NewBuilder[oauth2.TokenSource](func(server *httptest.Server) (oauth2.TokenSource, error) {
		cfg, err := NewAuthConfig("", "user", "secret", "")
		if err != nil {
			return nil, err
		}

		cfg.BaseURL, _ = url.Parse(server.URL)
		cfg.HTTPClient = server.Client()

		return cfg.TokenSource(ctx), nil
	})
}

func TestTokenSource_Token(t *testing.T) {
	tokenSource := mockTokenSource(t.Context()).
		Route("POST /authenticate",
			servermock.JSONEncode(APIResponse[*Token]{
				Data: &Token{
					Token:       "secrettoken",
					TokenExpire: time.Now().Add(7200 * time.Second).UTC().Unix(),
				},
			}),
			servermock.CheckRequestJSONBodyFromFixture("authenticate-request.json"),
		).
		Build(t)

	token, err := tokenSource.Token()
	require.NoError(t, err)

	assert.Equal(t, "secrettoken", token.AccessToken)
	assert.Equal(t, "Bearer", token.TokenType)
	assert.InDelta(t, 7200, token.ExpiresIn, 2)
	assert.WithinDuration(t, time.Now().Add(7200*time.Second), token.Expiry, 2*time.Second)
}

func TestTokenSource_authenticate(t *testing.T) {
	tokenSource := mockTokenSource(t.Context()).
		Route("POST /authenticate",
			servermock.ResponseFromFixture("authenticate.json"),
			servermock.CheckRequestJSONBodyFromFixture("authenticate-request.json"),
		).
		Build(t)

	token, err := tokenSource.(*TokenSource).authenticate(t.Context())
	require.NoError(t, err)

	expected := &Token{
		Token:              "secrettoken",
		TokenExpire:        1577836800,
		CustomerID:         "16030",
		ContactID:          "15182",
		CustomerName:       "Cloudline AS",
		ContactUsername:    "test@example.com",
		ContactAccessLevel: "admin",
		CustomerAddress:    "Grønland 14",
		CustomerZipcode:    "5918",
		CustomerCity:       "Frekhaug",
		CustomerProvince:   "Vestland",
		GASecret:           "ga_secret",
		GAEnabled:          "1",
		VAT:                1,
	}

	assert.Equal(t, expected, token)
}
