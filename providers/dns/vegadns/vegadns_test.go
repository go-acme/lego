package vegadns

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/go-acme/lego/v4/platform/tester/clientmock"
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
		builder       *clientmock.Builder[*DNSProvider]
		expectedError string
	}{
		{
			desc: "Success",
			builder: mockBuilder().
				Route("POST /1.0/token",
					clientmock.ResponseFromFixture("token.json")).
				Route("GET /1.0/domains", getDomainHandler()).
				Route("POST /1.0/records",
					clientmock.ResponseFromFixture("create_record.json").
						WithStatusCode(http.StatusCreated)),
		},
		{
			desc: "FailToFindZone",
			builder: mockBuilder().
				Route("POST /1.0/token",
					clientmock.ResponseFromFixture("token.json")).
				Route("GET /1.0/domains",
					clientmock.Noop().
						WithStatusCode(http.StatusNotFound)),
			expectedError: "vegadns: can't find Authoritative Zone for _acme-challenge.example.com. in Present: Unable to find auth zone for fqdn _acme-challenge.example.com",
		},
		{
			desc: "FailToCreateTXT",
			builder: mockBuilder().
				Route("POST /1.0/token",
					clientmock.ResponseFromFixture("token.json")).
				Route("GET /1.0/domains", getDomainHandler()).
				Route("POST /1.0/records",
					clientmock.Noop().
						WithStatusCode(http.StatusBadRequest)),
			expectedError: "vegadns: Got bad answer from VegaDNS on CreateTXT. Code: 400. Message: ",
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
		builder       *clientmock.Builder[*DNSProvider]
		expectedError string
	}{
		{
			desc: "Success",
			builder: mockBuilder().
				Route("POST /1.0/token",
					clientmock.ResponseFromFixture("token.json")).
				Route("GET /1.0/domains", getDomainHandler()).
				Route("GET /1.0/records",
					clientmock.ResponseFromFixture("records.json"),
					clientmock.CheckQueryParameter().With("domain_id", "1")).
				Route("DELETE /1.0/records/3",
					clientmock.ResponseFromFixture("record_delete.json")),
		},
		{
			desc: "FailToFindZone",
			builder: mockBuilder().
				Route("POST /1.0/token",
					clientmock.ResponseFromFixture("token.json")).
				Route("GET /1.0/domains",
					clientmock.Noop().
						WithStatusCode(http.StatusNotFound)),
			expectedError: "vegadns: can't find Authoritative Zone for _acme-challenge.example.com. in CleanUp: Unable to find auth zone for fqdn _acme-challenge.example.com",
		},
		{
			desc: "FailToGetRecordID",
			builder: mockBuilder().
				Route("POST /1.0/token",
					clientmock.ResponseFromFixture("token.json")).
				Route("GET /1.0/domains", getDomainHandler()).
				Route("GET /1.0/records",
					clientmock.Noop().
						WithStatusCode(http.StatusNotFound),
					clientmock.CheckQueryParameter().With("domain_id", "1")),
			expectedError: "vegadns: couldn't get Record ID in CleanUp: Got bad answer from VegaDNS on GetRecordID. Code: 404. Message: ",
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

func mockBuilder() *clientmock.Builder[*DNSProvider] {
	return clientmock.NewBuilder(func(server *httptest.Server) (*DNSProvider, error) {
		envTest.Apply(map[string]string{
			EnvKey:    "key",
			EnvSecret: "secret",
			EnvURL:    server.URL,
		})

		return NewDNSProvider()
	})
}
