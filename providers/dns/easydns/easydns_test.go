package easydns

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(
	EnvEndpoint,
	EnvToken,
	EnvKey).
	WithDomain(envDomain)

func mockBuilder() *servermock.Builder[*DNSProvider] {
	return servermock.NewBuilder(
		func(server *httptest.Server) (*DNSProvider, error) {
			endpoint, err := url.Parse(server.URL)
			if err != nil {
				return nil, err
			}

			config := NewDefaultConfig()
			config.Token = "TOKEN"
			config.Key = "SECRET"
			config.Endpoint = endpoint
			config.HTTPClient = server.Client()

			return NewDNSProviderConfig(config)
		},
		servermock.CheckHeader().
			WithJSONHeaders().
			WithAuthorization("Basic VE9LRU46U0VDUkVU"),
		servermock.CheckQueryParameter().Strict().
			With("format", "json"))
}

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvToken: "TOKEN",
				EnvKey:   "SECRET",
			},
		},
		{
			desc: "missing token",
			envVars: map[string]string{
				EnvKey: "SECRET",
			},
			expected: "easydns: some credentials information are missing: EASYDNS_TOKEN",
		},
		{
			desc: "missing key",
			envVars: map[string]string{
				EnvToken: "TOKEN",
			},
			expected: "easydns: some credentials information are missing: EASYDNS_KEY",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer envTest.RestoreEnv()
			envTest.ClearEnv()

			envTest.Apply(test.envVars)

			p, err := NewDNSProvider()

			if test.expected == "" {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.config)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc     string
		config   *Config
		expected string
	}{
		{
			desc: "success",
			config: &Config{
				Token: "TOKEN",
				Key:   "KEY",
			},
		},
		{
			desc:     "nil config",
			config:   nil,
			expected: "easydns: the configuration of the DNS provider is nil",
		},
		{
			desc: "missing token",
			config: &Config{
				Key: "KEY",
			},
			expected: "easydns: the API token is missing",
		},
		{
			desc: "missing key",
			config: &Config{
				Token: "TOKEN",
			},
			expected: "easydns: the API key is missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			p, err := NewDNSProviderConfig(test.config)

			if test.expected == "" {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.config)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestDNSProvider_Present(t *testing.T) {
	provider := mockBuilder().
		Route("GET /zones/records/all/example.com",
			servermock.RawStringResponse(`{
		  "msg": "string",
		  "status": 200,
		  "tm": 0,
		  "data": [{
		    "id": "60898922",
		    "domain": "example.com",
		    "host": "hosta",
		    "ttl": "300",
		    "prio": "0",
		    "geozone_id": "0",
		    "type": "A",
		    "rdata": "1.2.3.4",
		    "last_mod": "2019-08-28 19:09:50"
		  }],
		  "count": 0,
		  "total": 0,
		  "start": 0,
		  "max": 0
		}
		`),
			servermock.CheckQueryParameter().Strict().
				With("format", "json")).
		Route("PUT /zones/records/add/example.com/TXT",
			servermock.RawStringResponse(`{
				"msg": "OK",
				"tm": 1554681934,
				"data": {
					"host": "_acme-challenge",
					"geozone_id": 0,
					"ttl": "120",
					"prio": "0",
					"rdata": "pW9ZKG0xz_PCriK-nCMOjADy9eJcgGWIzkkj2fN4uZM",
					"revoked": 0,
					"id": "123456789",
					"new_host": "_acme-challenge.example.com"
				},
				"status": 201
			}`),
			servermock.CheckRequestJSONBody(`{"domain":"example.com","host":"_acme-challenge","ttl":"120","prio":"0","type":"TXT","rdata":"pW9ZKG0xz_PCriK-nCMOjADy9eJcgGWIzkkj2fN4uZM"}
`)).
		Build(t)

	err := provider.Present("example.com", "token", "keyAuth")
	require.NoError(t, err)
	require.Contains(t, provider.recordIDs, "_acme-challenge.example.com.|pW9ZKG0xz_PCriK-nCMOjADy9eJcgGWIzkkj2fN4uZM")
}

func TestDNSProvider_Cleanup_WhenRecordIdNotSet_NoOp(t *testing.T) {
	provider := mockBuilder().
		Route("GET /zones/records/all/_acme-challenge.example.com",
			servermock.RawStringResponse(`{
	  "msg": "string",
	  "status": 200,
	  "tm": 0,
	  "data": [{
	    "id": "60898922",
	    "domain": "example.com",
	    "host": "hosta",
	    "ttl": "300",
	    "prio": "0",
	    "geozone_id": "0",
	    "type": "A",
	    "rdata": "1.2.3.4",
	    "last_mod": "2019-08-28 19:09:50"
	  }],
	  "count": 0,
	  "total": 0,
	  "start": 0,
	  "max": 0
	}
	`)).
		Build(t)

	err := provider.CleanUp("example.com", "token", "keyAuth")
	require.NoError(t, err)
}

func TestDNSProvider_Cleanup_WhenRecordIdSet_DeletesTxtRecord(t *testing.T) {
	provider := mockBuilder().
		Route("GET /zones/records/all/_acme-challenge.example.com",
			servermock.RawStringResponse(`{
	  "msg": "string",
	  "status": 200,
	  "tm": 0,
	  "data": [{
	    "id": "60898922",
	    "domain": "example.com",
	    "host": "hosta",
	    "ttl": "300",
	    "prio": "0",
	    "geozone_id": "0",
	    "type": "A",
	    "rdata": "1.2.3.4",
	    "last_mod": "2019-08-28 19:09:50"
	  }],
	  "count": 0,
	  "total": 0,
	  "start": 0,
	  "max": 0
	}
	`)).
		Route("DELETE /zones/records/_acme-challenge.example.com/123456",
			servermock.RawStringResponse(`{
				"msg": "OK",
				"data": {
					"domain": "example.com",
					"id": "123456"
				},
				"status": 200
			}`)).
		Build(t)

	provider.recordIDs["_acme-challenge.example.com.|pW9ZKG0xz_PCriK-nCMOjADy9eJcgGWIzkkj2fN4uZM"] = "123456"

	err := provider.CleanUp("example.com", "token", "keyAuth")
	require.NoError(t, err)
}

func TestDNSProvider_Cleanup_WhenHttpError_ReturnsError(t *testing.T) {
	errorMessage := `{
		"error": {
			"code": 406,
			"message": "Provided id is invalid or you do not have permission to access it."
		}
	}`

	provider := mockBuilder().
		Route("GET /zones/records/all/example.com",
			servermock.RawStringResponse(`{
  "msg": "string",
  "status": 200,
  "tm": 0,
  "data": [{
    "id": "60898922",
    "domain": "example.com",
    "host": "hosta",
    "ttl": "300",
    "prio": "0",
    "geozone_id": "0",
    "type": "A",
    "rdata": "1.2.3.4",
    "last_mod": "2019-08-28 19:09:50"
  }],
  "count": 0,
  "total": 0,
  "start": 0,
  "max": 0
}
`)).
		Route("DELETE /zones/records/example.com/123456",
			servermock.RawStringResponse(errorMessage).
				WithStatusCode(http.StatusNotAcceptable)).
		Build(t)

	provider.recordIDs["_acme-challenge.example.com.|pW9ZKG0xz_PCriK-nCMOjADy9eJcgGWIzkkj2fN4uZM"] = "123456"

	err := provider.CleanUp("example.com", "token", "keyAuth")

	expectedError := fmt.Sprintf("easydns: unexpected status code: [status code: 406] body: %v", errorMessage)
	require.EqualError(t, err, expectedError)
}

func TestLivePresent(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()
	provider, err := NewDNSProvider()
	require.NoError(t, err)

	err = provider.Present(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}

func TestLiveCleanUp(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()
	provider, err := NewDNSProvider()
	require.NoError(t, err)

	time.Sleep(2 * time.Second)

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}
