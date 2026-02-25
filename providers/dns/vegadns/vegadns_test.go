package vegadns

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-acme/lego/v5/internal/tester"
	servermock2 "github.com/go-acme/lego/v5/internal/tester/servermock"
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
		builder       *servermock2.Builder[*DNSProvider]
		expectedError string
	}{
		{
			desc: "success",
			builder: mockBuilder().
				Route("POST /1.0/token",
					servermock2.ResponseFromFixture("token.json")).
				Route("GET /1.0/domains", getDomainHandler()).
				Route("POST /1.0/records",
					servermock2.ResponseFromFixture("create_record.json").
						WithStatusCode(http.StatusCreated)),
		},
		{
			desc: "fail to find the zone",
			builder: mockBuilder().
				Route("POST /1.0/token",
					servermock2.ResponseFromFixture("token.json")).
				Route("GET /1.0/domains",
					servermock2.Noop().
						WithStatusCode(http.StatusNotFound)),
			expectedError: "vegadns: find domain ID for _acme-challenge.example.com.: domain not found",
		},
		{
			desc: "fail to create TXT record",
			builder: mockBuilder().
				Route("POST /1.0/token",
					servermock2.ResponseFromFixture("token.json")).
				Route("GET /1.0/domains", getDomainHandler()).
				Route("POST /1.0/records",
					servermock2.Noop().
						WithStatusCode(http.StatusBadRequest)),
			expectedError: "vegadns: create TXT record: bad answer from VegaDNS (code: 400, message: )",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer envTest.RestoreEnv()

			envTest.ClearEnv()

			provider := test.builder.Build(t)

			err := provider.Present(t.Context(), testDomain, "token", "keyAuth")
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
		builder       *servermock2.Builder[*DNSProvider]
		expectedError string
	}{
		{
			desc: "success",
			builder: mockBuilder().
				Route("POST /1.0/token",
					servermock2.ResponseFromFixture("token.json")).
				Route("GET /1.0/domains", getDomainHandler()).
				Route("GET /1.0/records",
					servermock2.ResponseFromFixture("records.json"),
					servermock2.CheckQueryParameter().With("domain_id", "1")).
				Route("DELETE /1.0/records/3",
					servermock2.ResponseFromFixture("record_delete.json")),
		},
		{
			desc: "fail to find the zone",
			builder: mockBuilder().
				Route("POST /1.0/token",
					servermock2.ResponseFromFixture("token.json")).
				Route("GET /1.0/domains",
					servermock2.Noop().
						WithStatusCode(http.StatusNotFound)),
			expectedError: "vegadns: find domain ID for _acme-challenge.example.com.: domain not found",
		},
		{
			desc: "fail to get record ID",
			builder: mockBuilder().
				Route("POST /1.0/token",
					servermock2.ResponseFromFixture("token.json")).
				Route("GET /1.0/domains", getDomainHandler()).
				Route("GET /1.0/records",
					servermock2.Noop().
						WithStatusCode(http.StatusNotFound),
					servermock2.CheckQueryParameter().With("domain_id", "1")),
			expectedError: "vegadns: find record ID for 1: get records: bad answer from VegaDNS (code: 404, message: )",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer envTest.RestoreEnv()

			envTest.ClearEnv()

			provider := test.builder.Build(t)

			err := provider.CleanUp(t.Context(), testDomain, "token", "keyAuth")
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

func mockBuilder() *servermock2.Builder[*DNSProvider] {
	return servermock2.NewBuilder(func(server *httptest.Server) (*DNSProvider, error) {
		envTest.Apply(map[string]string{
			EnvKey:    "key",
			EnvSecret: "secret",
			EnvURL:    server.URL,
		})

		return NewDNSProvider()
	})
}
