package rackspace

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(
	EnvUser,
	EnvAPIKey).
	WithDomain(envDomain)

func TestNewDNSProviderConfig(t *testing.T) {
	provider := mockBuilder().Build(t)

	assert.Equal(t, "testToken", provider.token, "The token should match")
}

func TestNewDNSProviderConfig_MissingCredErr(t *testing.T) {
	_, err := NewDNSProviderConfig(NewDefaultConfig())
	require.EqualError(t, err, "rackspace: credentials missing")
}

func TestDNSProvider_Present(t *testing.T) {
	provider := mockBuilder().
		Route("GET /123456/domains",
			servermock.ResponseFromFixture("zone_details.json"),
			servermock.CheckQueryParameter().Strict().
				With("name", "example.com")).
		Route("POST /123456/domains/112233/records",
			servermock.ResponseFromFixture("record.json").
				WithStatusCode(http.StatusAccepted),
			servermock.CheckRequestJSONBody(`{"records":[{"name":"_acme-challenge.example.com","type":"TXT","data":"pW9ZKG0xz_PCriK-nCMOjADy9eJcgGWIzkkj2fN4uZM","ttl":300}]}`)).
		Build(t)

	err := provider.Present("example.com", "token", "keyAuth")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockBuilder().
		Route("GET /123456/domains",
			servermock.ResponseFromFixture("zone_details.json"),
			servermock.CheckQueryParameter().Strict().
				With("name", "example.com")).
		Route("GET /123456/domains/112233/records",
			servermock.ResponseFromFixture("record_details.json"),
			servermock.CheckQueryParameter().Strict().
				With("type", "TXT").
				With("name", "_acme-challenge.example.com")).
		Route("DELETE /123456/domains/112233/records",
			servermock.ResponseFromFixture("delete.json"),
			servermock.CheckQueryParameter().Strict().
				With("id", "TXT-654321")).
		Build(t)

	err := provider.CleanUp("example.com", "token", "keyAuth")
	require.NoError(t, err)
}

func TestLiveNewDNSProvider_ValidEnv(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()
	provider, err := NewDNSProvider()
	require.NoError(t, err)

	assert.Contains(t, provider.cloudDNSEndpoint, "https://dns.api.rackspacecloud.com/v1.0/", "The endpoint URL should contain the base")
}

func TestLivePresent(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()
	provider, err := NewDNSProvider()
	require.NoError(t, err)

	err = provider.Present(envTest.GetDomain(), "", "112233445566==")
	require.NoError(t, err)
}

func TestLiveCleanUp(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()
	provider, err := NewDNSProvider()
	require.NoError(t, err)

	time.Sleep(15 * time.Second)

	err = provider.CleanUp(envTest.GetDomain(), "", "112233445566==")
	require.NoError(t, err)
}

func mockBuilder() *servermock.Builder[*DNSProvider] {
	return servermock.NewBuilder(
		func(server *httptest.Server) (*DNSProvider, error) {
			config := NewDefaultConfig()
			config.APIUser = "testUser"
			config.APIKey = "testKey"
			config.HTTPClient = server.Client()
			config.BaseURL = server.URL + "/v2.0/tokens"

			return NewDNSProviderConfig(config)
		},
		servermock.CheckHeader().WithJSONHeaders(),
	).
		Route("POST /v2.0/tokens",
			http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				apiURL := fmt.Sprintf("http://%s/123456", req.Context().Value(http.LocalAddrContextKey))

				resp := strings.Replace(`
{
  "access": {
    "token": {
      "id": "testToken",
      "expires": "1970-01-01T00:00:00.000Z",
      "tenant": {
        "id": "123456",
        "name": "123456"
      },
      "RAX-AUTH:authenticatedBy": [
        "APIKEY"
      ]
    },
    "serviceCatalog": [
      {
        "type": "rax:dns",
        "endpoints": [
          {
            "publicURL": "https://dns.api.rackspacecloud.com/v1.0/123456",
            "tenantId": "123456"
          }
        ],
        "name": "cloudDNS"
      }
    ],
    "user": {
      "id": "fakeUseID",
      "name": "testUser"
    }
  }
}
`, "https://dns.api.rackspacecloud.com/v1.0/123456", apiURL, 1)

				rw.WriteHeader(http.StatusOK)
				_, _ = fmt.Fprint(rw, resp)
			}),
			servermock.CheckRequestJSONBody(`{"auth":{"RAX-KSKEY:apiKeyCredentials":{"username":"testUser","apiKey":"testKey"}}}`))
}
