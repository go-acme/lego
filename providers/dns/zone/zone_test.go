package zone

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/xenolf/lego/platform/tester"
)

var envTest = tester.NewEnvTest("ZONE_ENDPOINT", "ZONE_API_USER", "ZONE_API_KEY")

const testResponse = `[
  {
    "destination": "LHDhK3oGRvkiefQnx7OOczTY5Tic_xZ6HcMOc_gmtoM",
    "resource_url": "http://localhost:8090/v2/dns/example.com/txt/1234",
    "id": "123",
    "name": "prefix.example.com",
    "delete": true,
    "modify": true
  }
]`

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				"ZONE_ENDPOINT": "http://localhost:8090/v2/dns/",
			},
		},
		{
			desc: "invalid URL",
			envVars: map[string]string{
				"ZONE_ENDPOINT": ":",
			},
			expected: "zone: parse :: missing protocol scheme",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer envTest.RestoreEnv()
			envTest.ClearEnv()

			envTest.Apply(test.envVars)

			p, err := NewDNSProvider()

			if len(test.expected) == 0 {
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
		endpoint *url.URL
		expected string
	}{
		{
			desc:     "success",
			endpoint: mustParse("http://localhost:8090/v2/dns/"),
		},
		{
			desc:     "missing endpoint",
			expected: "zone: the endpoint is missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.Endpoint = test.endpoint

			p, err := NewDNSProviderConfig(config)

			if len(test.expected) == 0 {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.config)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestNewDNSProvider_Present(t *testing.T) {
	envTest.RestoreEnv()

	testCases := []struct {
		desc          string
		username      string
		apikey        string
		handler       http.HandlerFunc
		expectedError string
	}{
		{
			desc:    "success",
			handler: successHandler,
		},
		{
			desc:          "error",
			handler:       http.NotFound,
			expectedError: "zone: 404: request failed: 404 page not found\n",
		},
		{
			desc:     "basic auth",
			username: "bar",
			apikey:   "foo",
			handler: func(rw http.ResponseWriter, req *http.Request) {
				username, apikey, ok := req.BasicAuth()
				if username != "bar" || apikey != "foo" || !ok {
					rw.Header().Set("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s"`, "Please enter your username and api key."))
					http.Error(rw, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
					return
				}

				fmt.Fprint(rw, testResponse)
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			mux := http.NewServeMux()
			mux.HandleFunc("/", test.handler)
			server := httptest.NewServer(mux)

			config := NewDefaultConfig()
			config.Endpoint = mustParse(server.URL)
			config.Username = test.username
			config.APIKey = test.apikey

			p, err := NewDNSProviderConfig(config)
			require.NoError(t, err)

			err = p.Present("domain.com", "token", "key")
			if test.expectedError == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, test.expectedError)
			}
		})
	}
}

func TestNewDNSProvider_Cleanup(t *testing.T) {
	envTest.RestoreEnv()

	testCases := []struct {
		desc          string
		username      string
		apikey        string
		handler       http.HandlerFunc
		expectedError string
	}{
		{
			desc:    "success",
			handler: successHandler,
		},
		{
			desc:          "error",
			handler:       http.NotFound,
			expectedError: "zone: 404: request failed: 404 page not found\n",
		},
		{
			desc:     "basic auth",
			username: "bar",
			apikey:   "foo",
			handler: func(rw http.ResponseWriter, req *http.Request) {
				username, apikey, ok := req.BasicAuth()
				if username != "bar" || apikey != "foo" || !ok {
					rw.Header().Set("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s"`, "Please enter your username and apikey."))
					http.Error(rw, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
					return
				}
				fmt.Fprint(rw, testResponse)
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			mux := http.NewServeMux()
			mux.HandleFunc("/", test.handler)
			server := httptest.NewServer(mux)

			config := NewDefaultConfig()
			config.Endpoint = mustParse(server.URL)
			config.Username = test.username
			config.APIKey = test.apikey

			p, err := NewDNSProviderConfig(config)
			require.NoError(t, err)

			err = p.CleanUp("domain.com", "token", "key")
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

func successHandler(rw http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodDelete {
		fmt.Fprint(rw, `[]`)
		return
	}

	if req.Method == http.MethodGet {
		fmt.Fprint(rw, testResponse)
		return
	}

	if req.Method != http.MethodPost {
		http.Error(rw, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	msg := &message{}
	err := json.NewDecoder(req.Body).Decode(msg)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Fprint(rw, testResponse)
}

func mustParse(rawURL string) *url.URL {
	uri, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}
	return uri
}
