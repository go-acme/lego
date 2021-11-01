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
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

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
		endpoint string
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

			if len(test.endpoint) > 0 {
				config.Endpoint = mustParse(test.endpoint)
			}

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
		username      string
		apiKey        string
		handlers      map[string]http.HandlerFunc
		expectedError string
	}{
		{
			desc:     "success",
			username: "bar",
			apiKey:   "foo",
			handlers: map[string]http.HandlerFunc{
				"/" + hostedZone + "/txt": mockHandlerCreateRecord,
			},
		},
		{
			desc:     "invalid auth",
			username: "nope",
			apiKey:   "foo",
			handlers: map[string]http.HandlerFunc{
				"/" + hostedZone + "/txt": mockHandlerCreateRecord,
			},
			expectedError: "zoneee: status code=401: Unauthorized\n",
		},
		{
			desc:          "error",
			username:      "bar",
			apiKey:        "foo",
			expectedError: "zoneee: status code=404: 404 page not found\n",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			mux := http.NewServeMux()
			for uri, handler := range test.handlers {
				mux.HandleFunc(uri, handler)
			}

			server := httptest.NewServer(mux)
			t.Cleanup(server.Close)

			config := NewDefaultConfig()
			config.Endpoint = mustParse(server.URL)
			config.Username = test.username
			config.APIKey = test.apiKey

			p, err := NewDNSProviderConfig(config)
			require.NoError(t, err)

			err = p.Present(domain, "token", "key")
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
		username      string
		apiKey        string
		handlers      map[string]http.HandlerFunc
		expectedError string
	}{
		{
			desc:     "success",
			username: "bar",
			apiKey:   "foo",
			handlers: map[string]http.HandlerFunc{
				"/" + hostedZone + "/txt": mockHandlerGetRecords([]txtRecord{{
					ID:          "1234",
					Name:        domain,
					Destination: "LHDhK3oGRvkiefQnx7OOczTY5Tic_xZ6HcMOc_gmtoM",
					Delete:      true,
					Modify:      true,
				}}),
				"/" + hostedZone + "/txt/1234": mockHandlerDeleteRecord,
			},
		},
		{
			desc:     "no txt records",
			username: "bar",
			apiKey:   "foo",
			handlers: map[string]http.HandlerFunc{
				"/" + hostedZone + "/txt":      mockHandlerGetRecords([]txtRecord{}),
				"/" + hostedZone + "/txt/1234": mockHandlerDeleteRecord,
			},
			expectedError: "zoneee: txt record does not exist for LHDhK3oGRvkiefQnx7OOczTY5Tic_xZ6HcMOc_gmtoM",
		},
		{
			desc:     "invalid auth",
			username: "nope",
			apiKey:   "foo",
			handlers: map[string]http.HandlerFunc{
				"/" + hostedZone + "/txt": mockHandlerGetRecords([]txtRecord{{
					ID:          "1234",
					Name:        domain,
					Destination: "LHDhK3oGRvkiefQnx7OOczTY5Tic_xZ6HcMOc_gmtoM",
					Delete:      true,
					Modify:      true,
				}}),
				"/" + hostedZone + "/txt/1234": mockHandlerDeleteRecord,
			},
			expectedError: "zoneee: status code=401: Unauthorized\n",
		},
		{
			desc:          "error",
			username:      "bar",
			apiKey:        "foo",
			expectedError: "zoneee: status code=404: 404 page not found\n",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			mux := http.NewServeMux()
			for uri, handler := range test.handlers {
				mux.HandleFunc(uri, handler)
			}

			server := httptest.NewServer(mux)
			t.Cleanup(server.Close)

			config := NewDefaultConfig()
			config.Endpoint = mustParse(server.URL)
			config.Username = test.username
			config.APIKey = test.apiKey

			p, err := NewDNSProviderConfig(config)
			require.NoError(t, err)

			err = p.CleanUp(domain, "token", "key")
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

func mustParse(rawURL string) *url.URL {
	uri, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}
	return uri
}

func mockHandlerCreateRecord(rw http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(rw, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}

	username, apiKey, ok := req.BasicAuth()
	if username != "bar" || apiKey != "foo" || !ok {
		rw.Header().Set("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s"`, "Please enter your username and API key."))
		http.Error(rw, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	record := txtRecord{}
	err := json.NewDecoder(req.Body).Decode(&record)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	record.ID = "1234"
	record.Delete = true
	record.Modify = true
	record.ResourceURL = req.URL.String() + "/1234"

	bytes, err := json.Marshal([]txtRecord{record})
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	if _, err = rw.Write(bytes); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
}

func mockHandlerGetRecords(records []txtRecord) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(rw, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}

		username, apiKey, ok := req.BasicAuth()
		if username != "bar" || apiKey != "foo" || !ok {
			rw.Header().Set("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s"`, "Please enter your username and API key."))
			http.Error(rw, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		for _, value := range records {
			if value.ResourceURL == "" {
				value.ResourceURL = req.URL.String() + "/" + value.ID
			}
		}

		bytes, err := json.Marshal(records)
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

func mockHandlerDeleteRecord(rw http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodDelete {
		http.Error(rw, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	username, apiKey, ok := req.BasicAuth()
	if username != "bar" || apiKey != "foo" || !ok {
		rw.Header().Set("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s"`, "Please enter your username and API key."))
		http.Error(rw, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	rw.WriteHeader(http.StatusNoContent)
}
