package vegadns

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testDomain = "example.com"

var envTest = tester.NewEnvTest(EnvKey, EnvSecret, EnvURL)

func TestNewDNSProvider_Fail(t *testing.T) {
	defer envTest.RestoreEnv()

	envTest.ClearEnv()

	_, err := NewDNSProvider()
	require.Error(t, err, "VEGADNS_URL env missing")
}

func TestDNSProvider_TimeoutSuccess(t *testing.T) {
	defer envTest.RestoreEnv()

	envTest.ClearEnv()

	provider := mockBuilder().Build(t)

	timeout, interval := provider.Timeout()
	assert.Equal(t, 12*time.Minute, timeout)
	assert.Equal(t, 1*time.Minute, interval)
}

func TestDNSProvider_Present(t *testing.T) {
	testCases := []struct {
		desc          string
		handler       http.Handler
		builder       *servermock.Builder[*DNSProvider]
		expectedError string
	}{
		{
			desc: "success",
			builder: mockBuilder().
				Route("POST /1.0/token",
					servermock.ResponseFromFixture("token.json")).
				Route("GET /1.0/domains", getDomainHandler()).
				Route("POST /1.0/records",
					servermock.ResponseFromFixture("create_record.json").
						WithStatusCode(http.StatusCreated)),
		},
		{
			desc: "fail to find the zone",
			builder: mockBuilder().
				Route("POST /1.0/token",
					servermock.ResponseFromFixture("token.json")).
				Route("GET /1.0/domains",
					servermock.Noop().
						WithStatusCode(http.StatusNotFound)),
			expectedError: "vegadns: find domain ID for _acme-challenge.example.com.: domain not found",
		},
		{
			desc: "fail to create TXT record",
			builder: mockBuilder().
				Route("POST /1.0/token",
					servermock.ResponseFromFixture("token.json")).
				Route("GET /1.0/domains", getDomainHandler()).
				Route("POST /1.0/records",
					servermock.Noop().
						WithStatusCode(http.StatusBadRequest)),
			expectedError: "vegadns: create TXT record: bad answer from VegaDNS (code: 400, message: )",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer envTest.RestoreEnv()

			envTest.ClearEnv()

			provider := test.builder.Build(t)

			err := provider.Present(testDomain, "token", "keyAuth")
			if test.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, test.expectedError)
			}
		})
	}
}

func TestDNSProvider_CleanUp(t *testing.T) {
	testCases := []struct {
		desc          string
		builder       *servermock.Builder[*DNSProvider]
		expectedError string
	}{
		{
			desc: "success",
			builder: mockBuilder().
				Route("POST /1.0/token",
					servermock.ResponseFromFixture("token.json")).
				Route("GET /1.0/domains", getDomainHandler()).
				Route("GET /1.0/records",
					servermock.ResponseFromFixture("records.json"),
					servermock.CheckQueryParameter().With("domain_id", "1")).
				Route("DELETE /1.0/records/3",
					servermock.ResponseFromFixture("record_delete.json")),
		},
		{
			desc: "fail to find the zone",
			builder: mockBuilder().
				Route("POST /1.0/token",
					servermock.ResponseFromFixture("token.json")).
				Route("GET /1.0/domains",
					servermock.Noop().
						WithStatusCode(http.StatusNotFound)),
			expectedError: "vegadns: find domain ID for _acme-challenge.example.com.: domain not found",
		},
		{
			desc: "fail to get record ID",
			builder: mockBuilder().
				Route("POST /1.0/token",
					servermock.ResponseFromFixture("token.json")).
				Route("GET /1.0/domains", getDomainHandler()).
				Route("GET /1.0/records",
					servermock.Noop().
						WithStatusCode(http.StatusNotFound),
					servermock.CheckQueryParameter().With("domain_id", "1")),
			expectedError: "vegadns: find record ID for 1: get records: bad answer from VegaDNS (code: 404, message: )",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer envTest.RestoreEnv()

			envTest.ClearEnv()

			provider := test.builder.Build(t)

			err := provider.CleanUp(testDomain, "token", "keyAuth")
			if test.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, test.expectedError)
			}
		})
	}
}

func getDomainHandler() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Query().Get("search") == testDomain {
			fmt.Fprint(rw, `
{
  "domains":[
    {
      "domain_id":1,
      "domain":"example.com",
      "status":"active",
      "owner_id":0
    }
  ]
}
`)

			return
		}

		rw.WriteHeader(http.StatusNotFound)
	}
}

func mockBuilder() *servermock.Builder[*DNSProvider] {
	return servermock.NewBuilder(func(server *httptest.Server) (*DNSProvider, error) {
		envTest.Apply(map[string]string{
			EnvKey:    "key",
			EnvSecret: "secret",
			EnvURL:    server.URL,
		})

		return NewDNSProvider()
	})
}
