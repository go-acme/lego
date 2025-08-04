package zoneee

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/go-acme/lego/v4/providers/dns/zoneee/internal"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

const (
	fakeUsername = "user"
	fakeAPIKey   = "secret"
)

var envTest = tester.NewEnvTest(EnvEndpoint, EnvAPIUser, EnvAPIKey).
	WithLiveTestRequirements(EnvAPIUser, EnvAPIKey).
	WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvAPIUser: "123",
				EnvAPIKey:  "456",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				EnvAPIUser: "",
				EnvAPIKey:  "",
			},
			expected: "zoneee: some credentials information are missing: ZONEEE_API_USER,ZONEEE_API_KEY",
		},
		{
			desc: "missing username",
			envVars: map[string]string{
				EnvAPIUser: "",
				EnvAPIKey:  "456",
			},
			expected: "zoneee: some credentials information are missing: ZONEEE_API_USER",
		},
		{
			desc: "missing API key",
			envVars: map[string]string{
				EnvAPIUser: "123",
				EnvAPIKey:  "",
			},
			expected: "zoneee: some credentials information are missing: ZONEEE_API_KEY",
		},
		{
			desc: "invalid URL",
			envVars: map[string]string{
				EnvAPIUser:  "123",
				EnvAPIKey:   "456",
				EnvEndpoint: ":",
			},
			expected: `zoneee: parse ":": missing protocol scheme`,
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
		apiUser  string
		apiKey   string
		expected string
	}{
		{
			desc:    "success",
			apiKey:  "123",
			apiUser: "456",
		},
		{
			desc:     "missing credentials",
			expected: "zoneee: credentials missing: username",
		},
		{
			desc:     "missing api key",
			apiUser:  "456",
			expected: "zoneee: credentials missing: API key",
		},
		{
			desc:     "missing username",
			apiKey:   "123",
			expected: "zoneee: credentials missing: username",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.APIKey = test.apiKey
			config.Username = test.apiUser

			p, err := NewDNSProviderConfig(config)

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
	hostedZone := "example.com"
	domain := "prefix." + hostedZone

	testCases := []struct {
		desc          string
		builder       *servermock.Builder[*DNSProvider]
		expectedError string
	}{
		{
			desc: "success",
			builder: mockBuilder(fakeUsername, fakeAPIKey).
				Route("POST /dns/"+hostedZone+"/txt",
					mockHandlerCreateRecord()),
		},
		{
			desc: "invalid auth",
			builder: mockBuilder("nope", "nope").
				Route("POST /dns/"+hostedZone+"/txt", nil),
			expectedError: "zoneee: unexpected status code: [status code: 401] body: Unauthorized",
		},
		{
			desc:          "error",
			builder:       mockBuilder(fakeUsername, fakeAPIKey),
			expectedError: "zoneee: unexpected status code: [status code: 404] body: 404 page not found",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			provider := test.builder.Build(t)

			err := provider.Present(domain, "token", "key")
			if test.expectedError == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, test.expectedError)
			}
		})
	}
}

func TestDNSProvider_Cleanup(t *testing.T) {
	hostedZone := "example.com"
	domain := "prefix." + hostedZone

	testCases := []struct {
		desc          string
		builder       *servermock.Builder[*DNSProvider]
		expectedError string
	}{
		{
			desc: "success",
			builder: mockBuilder(fakeUsername, fakeAPIKey).
				Route("GET /dns/"+hostedZone+"/txt",
					mockHandlerGetRecords([]internal.TXTRecord{{
						ID:          "1234",
						Name:        domain,
						Destination: "LHDhK3oGRvkiefQnx7OOczTY5Tic_xZ6HcMOc_gmtoM",
						Delete:      true,
						Modify:      true,
					}})).
				Route("DELETE /dns/"+hostedZone+"/txt/1234",
					servermock.Noop().
						WithStatusCode(http.StatusNoContent)),
		},
		{
			desc: "no txt records",
			builder: mockBuilder(fakeUsername, fakeAPIKey).
				Route("GET /dns/"+hostedZone+"/txt",
					mockHandlerGetRecords([]internal.TXTRecord{})),
			expectedError: "zoneee: txt record does not exist for LHDhK3oGRvkiefQnx7OOczTY5Tic_xZ6HcMOc_gmtoM",
		},
		{
			desc: "invalid auth",
			builder: mockBuilder("nope", "nope").
				Route("GET /dns/"+hostedZone+"/txt", nil),
			expectedError: "zoneee: unexpected status code: [status code: 401] body: Unauthorized",
		},
		{
			desc:          "error",
			builder:       mockBuilder(fakeUsername, fakeAPIKey),
			expectedError: "zoneee: unexpected status code: [status code: 404] body: 404 page not found",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			provider := test.builder.Build(t)

			err := provider.CleanUp(domain, "token", "key")
			if test.expectedError == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, test.expectedError)
			}
		})
	}
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

func mockBuilder(username, apiKey string) *servermock.Builder[*DNSProvider] {
	return servermock.NewBuilder(
		func(server *httptest.Server) (*DNSProvider, error) {
			config := NewDefaultConfig()
			config.HTTPClient = server.Client()
			config.Endpoint, _ = url.Parse(server.URL)
			config.Username = username
			config.APIKey = apiKey

			return NewDNSProviderConfig(config)
		},
		checkBasicAuth())
}

func mockHandlerCreateRecord() http.HandlerFunc {
	return encodeJSONHandler(func(req *http.Request, rw http.ResponseWriter) (any, error) {
		record := internal.TXTRecord{}
		err := json.NewDecoder(req.Body).Decode(&record)
		if err != nil {
			return nil, err
		}

		record.ID = "1234"
		record.Delete = true
		record.Modify = true
		record.ResourceURL = req.URL.String() + "/1234"

		return []internal.TXTRecord{record}, nil
	})
}

func mockHandlerGetRecords(records []internal.TXTRecord) http.HandlerFunc {
	return encodeJSONHandler(func(req *http.Request, rw http.ResponseWriter) (any, error) {
		for _, record := range records {
			if record.ResourceURL == "" {
				record.ResourceURL = req.URL.String() + "/" + record.ID
			}
		}

		return records, nil
	})
}

func encodeJSONHandler(build func(req *http.Request, rw http.ResponseWriter) (any, error)) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		data, err := build(req, rw)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		bytes, err := json.Marshal(data)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		if _, err = rw.Write(bytes); err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func checkBasicAuth() servermock.LinkFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			username, apiKey, ok := req.BasicAuth()
			if username != fakeUsername || apiKey != fakeAPIKey || !ok {
				rw.Header().Set("WWW-Authenticate", fmt.Sprintf(`Basic realm=%q`, "Please enter your username and API key."))
				http.Error(rw, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(rw, req)
		})
	}
}
