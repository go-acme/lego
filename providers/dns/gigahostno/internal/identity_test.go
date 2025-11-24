package internal

import (
	"context"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupIdentifierClient(server *httptest.Server) (*Identifier, error) {
	client, err := NewIdentifier("user", "secret", "")
	if err != nil {
		return nil, err
	}

	client.BaseURL, _ = url.Parse(server.URL)
	client.HTTPClient = server.Client()

	return client, nil
}

func mockContext(t *testing.T) context.Context {
	t.Helper()

	return context.WithValue(t.Context(), tokenKey, "secret")
}

func TestIdentifier_Authenticate(t *testing.T) {
	identifier := servermock.NewBuilder[*Identifier](setupIdentifierClient).
		Route("POST /authenticate",
			servermock.ResponseFromFixture("authenticate.json"),
			servermock.CheckRequestJSONBodyFromFixture("authenticate-request.json")).
		Build(t)

	token, err := identifier.Authenticate(context.Background())
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

func TestToken_IsExpired(t *testing.T) {
	testCases := []struct {
		desc   string
		token  *Token
		assert assert.BoolAssertionFunc
	}{
		{
			desc:   "nil",
			assert: assert.True,
		},
		{
			desc:   "empty",
			token:  &Token{},
			assert: assert.True,
		},
		{
			desc: "not expired",
			token: &Token{
				TokenExpire: 65322892800, // 2040-01-01
			},
			assert: assert.False,
		},
		{
			desc: "now",
			token: &Token{
				TokenExpire: time.Now().Unix(),
			},
			assert: assert.True,
		},
		{
			desc: "now + 2 minutes",
			token: &Token{
				TokenExpire: time.Now().Add(2 * time.Minute).Unix(),
			},
			assert: assert.False,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			test.assert(t, test.token.IsExpired())
		})
	}
}
