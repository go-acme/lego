package internal

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupClient(server *httptest.Server) (*Client, error) {
	client := NewClient("user", "secret")
	client.HTTPClient = server.Client()
	client.baseURL, _ = url.Parse(server.URL)

	client.signer.saltShaker = func() []byte { return []byte("0123456789ABCDEF") }
	client.signer.clock = func() time.Time { return time.Unix(1692475113, 0) }

	return client, nil
}

func TestClient_AddRecord(t *testing.T) {
	client := servermock.NewBuilder[*Client](setupClient,
		servermock.CheckHeader().
			WithContentTypeFromURLEncoded().
			With(authenticationHeader, "user;1692475113;0123456789ABCDEF;24a32faf74c7bd0525f560ff12a1c1fb6545bafc"),
	).
		Route("POST /dns/example.com/addRR", nil, servermock.CheckForm().Strict().
			With("data", "txtTXTtxt").
			With("name", "sub").
			With("type", "TXT").
			With("ttl", "30"),
		).
		Build(t)

	record := Record{
		Name: "sub",
		Type: "TXT",
		Data: "txtTXTtxt",
		TTL:  30,
	}

	err := client.AddRecord(t.Context(), "example.com", record)
	require.NoError(t, err)
}

func TestClient_AddRecord_error(t *testing.T) {
	client := servermock.NewBuilder[*Client](setupClient,
		servermock.CheckHeader().
			WithContentTypeFromURLEncoded().
			With(authenticationHeader, "user;1692475113;0123456789ABCDEF;24a32faf74c7bd0525f560ff12a1c1fb6545bafc"),
	).
		Route("POST /dns/example.com/addRR",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized)).
		Build(t)

	record := Record{
		Name: "sub",
		Type: "TXT",
		Data: "txtTXTtxt",
		TTL:  30,
	}

	err := client.AddRecord(t.Context(), "example.com", record)
	require.Error(t, err)
}

func TestClient_RemoveRecord(t *testing.T) {
	client := servermock.NewBuilder[*Client](setupClient,
		servermock.CheckHeader().
			WithContentTypeFromURLEncoded().
			With(authenticationHeader, "user;1692475113;0123456789ABCDEF;699f01f077ca487bd66ac370d6dfc5b122c65522"),
	).
		Route("POST /dns/example.com/removeRR", nil,
			servermock.CheckForm().Strict().
				With("data", "txtTXTtxt").
				With("name", "sub").
				With("type", "TXT"),
		).
		Build(t)

	record := Record{
		Name: "sub",
		Type: "TXT",
		Data: "txtTXTtxt",
	}

	err := client.RemoveRecord(t.Context(), "example.com", record)
	require.NoError(t, err)
}

func TestClient_RemoveRecord_error(t *testing.T) {
	client := servermock.NewBuilder[*Client](setupClient,
		servermock.CheckHeader().
			WithContentTypeFromURLEncoded().
			With(authenticationHeader, "user;1692475113;0123456789ABCDEF;699f01f077ca487bd66ac370d6dfc5b122c65522"),
	).
		Route("POST /dns/example.com/removeRR",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized)).
		Build(t)

	record := Record{
		Name: "sub",
		Type: "TXT",
		Data: "txtTXTtxt",
	}

	err := client.RemoveRecord(t.Context(), "example.com", record)
	require.Error(t, err)
}

func TestSigner_Sign(t *testing.T) {
	testCases := []struct {
		desc     string
		path     string
		now      int64
		salt     string
		expected string
	}{
		{
			desc:     "basic",
			path:     "/path",
			now:      1692475113,
			salt:     "0123456789ABCDEF",
			expected: "user;1692475113;0123456789ABCDEF;417a9988c7ad7919b297884dd120b5808d8a1e6f",
		},
		{
			desc:     "another date",
			path:     "/path",
			now:      1692567766,
			salt:     "0123456789ABCDEF",
			expected: "user;1692567766;0123456789ABCDEF;b5c28286fd2e1a45a7c576dc2a6430116f721502",
		},
		{
			desc:     "another salt",
			path:     "/path",
			now:      1692475113,
			salt:     "FEDCBA9876543210",
			expected: "user;1692475113;FEDCBA9876543210;0f766822bda4fdc09829be4e1ea5e27ae3ae334e",
		},
		{
			desc:     "empty path",
			path:     "",
			now:      1692475113,
			salt:     "0123456789ABCDEF",
			expected: "user;1692475113;0123456789ABCDEF;c7c241a4d15d04d92805631d58d4d72ac1c339a1",
		},
		{
			desc:     "root path",
			path:     "/",
			now:      1692475113,
			salt:     "0123456789ABCDEF",
			expected: "user;1692475113;0123456789ABCDEF;c7c241a4d15d04d92805631d58d4d72ac1c339a1",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			signer := NewSigner()
			signer.saltShaker = func() []byte { return []byte(test.salt) }
			signer.clock = func() time.Time { return time.Unix(test.now, 0) }

			sign := signer.Sign(test.path, "data", "user", "secret")

			assert.Equal(t, test.expected, sign)
		})
	}
}
