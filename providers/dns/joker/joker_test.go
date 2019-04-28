package joker

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/go-acme/lego/platform/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var envTest = tester.NewEnvTest("JOKER_API_KEY").WithDomain("JOKER_DOMAIN")

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				"JOKER_API_KEY": "123",
			},
		},
		{
			desc: "missing key",
			envVars: map[string]string{
				"JOKER_API_KEY": "",
			},
			expected: "joker: some credentials information are missing: JOKER_API_KEY",
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
				assert.NotNil(t, p.config)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc            string
		authKey         string
		baseURL         string
		expected        string
		expectedBaseURL string
	}{
		{
			desc:            "success",
			authKey:         "123",
			expectedBaseURL: defaultBaseURL,
		},
		{
			desc:            "missing credentials",
			expected:        "joker: credentials missing",
			expectedBaseURL: defaultBaseURL,
		},
		{
			desc:            "Base URL should ends with /",
			authKey:         "123",
			baseURL:         "http://example.com",
			expectedBaseURL: "http://example.com/",
		},
		{
			desc:            "Base URL already ends with /",
			authKey:         "123",
			baseURL:         "http://example.com/",
			expectedBaseURL: "http://example.com/",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.APIKey = test.authKey
			if test.baseURL != "" {
				config.BaseURL = test.baseURL
			}

			p, err := NewDNSProviderConfig(config)

			if len(test.expected) == 0 {
				require.NoError(t, err)
				require.NotNil(t, p)
				assert.NotNil(t, p.config)
				assert.Equal(t, test.expectedBaseURL, p.config.BaseURL)
			} else {
				require.EqualError(t, err, test.expected)
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

func TestRemoveTxtEntryFromZone(t *testing.T) {
	testCases := []struct {
		desc     string
		input    string
		expected string
		modified bool
	}{
		{
			desc:     "empty zone",
			input:    "",
			expected: "",
			modified: false,
		},
		{
			desc:     "zone with only A entry",
			input:    "@ A 0 192.0.2.2 3600",
			expected: "@ A 0 192.0.2.2 3600",
			modified: false,
		},
		{
			desc:     "zone with only clenup entry",
			input:    "_acme-challenge TXT 0  \"old \" 120",
			expected: "",
			modified: true,
		},
		{
			desc:     "zone with one A and one cleanup entries",
			input:    "@ A 0 192.0.2.2 3600\n_acme-challenge TXT 0  \"old \" 120",
			expected: "@ A 0 192.0.2.2 3600",
			modified: true,
		},
		{
			desc:     "zone with one A and multiple cleanup entries",
			input:    "@ A 0 192.0.2.2 3600\n_acme-challenge TXT 0  \"old \" 120\n_acme-challenge TXT 0  \"another \" 120",
			expected: "@ A 0 192.0.2.2 3600",
			modified: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {

			zone, modified := removeTxtEntryFromZone(test.input, "_acme-challenge")
			assert.Equal(t, zone, test.expected)
			assert.Equal(t, modified, test.modified)
		})
	}
}

func TestAddTxtEntryToZone(t *testing.T) {
	testCases := []struct {
		desc     string
		input    string
		expected string
	}{
		{
			desc:     "empty zone",
			input:    "",
			expected: "_acme-challenge TXT 0 \"test\" 120",
		},
		{
			desc:     "zone with A entry",
			input:    "@ A 0 192.0.2.2 3600",
			expected: "@ A 0 192.0.2.2 3600\n_acme-challenge TXT 0 \"test\" 120",
		},
		{
			desc:     "zone with required clenup entry",
			input:    "_acme-challenge TXT 0  \"old \" 120",
			expected: "_acme-challenge TXT 0 \"old\" 120\n_acme-challenge TXT 0 \"test\" 120",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {

			zone := addTxtEntryToZone(test.input, "_acme-challenge", "test", 120)
			assert.Equal(t, zone, test.expected)
		})
	}
}

func TestFixTxtLines(t *testing.T) {
	testCases := []struct {
		desc     string
		input    string
		expected string
	}{
		{
			desc:     "clean-up",
			input:    `_acme-challenge TXT 0  "SrqD25Gpm3WtIGKCqhgsLeXWE_FAD5Hv9CRoLAHxlIE " 120`,
			expected: `_acme-challenge TXT 0 "SrqD25Gpm3WtIGKCqhgsLeXWE_FAD5Hv9CRoLAHxlIE" 120`,
		},
		{
			desc:     "already cleaned",
			input:    `_acme-challenge TXT 0 "SrqD25Gpm3WtIGKCqhgsLeXWE_FAD5Hv9CRoLAHxlIE" 120`,
			expected: `_acme-challenge TXT 0 "SrqD25Gpm3WtIGKCqhgsLeXWE_FAD5Hv9CRoLAHxlIE" 120`,
		},
		{
			desc:     "special DNS entry",
			input:    "$dyndns=yes:username:password",
			expected: "$dyndns=yes:username:password",
		},
		{
			desc:     "SRV entry",
			input:    "_jabber._tcp SRV 20/0 xmpp-server1.l.google.com:5269 300",
			expected: "_jabber._tcp SRV 20/0 xmpp-server1.l.google.com:5269 300",
		},
		{
			desc:     "MX entry",
			input:    "@ MX 10 ASPMX.L.GOOGLE.COM 300",
			expected: "@ MX 10 ASPMX.L.GOOGLE.COM 300",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {

			line := fixTxtLines(test.input)
			assert.Equal(t, line, test.expected)
		})
	}
}

func TestParseJokerResponse(t *testing.T) {
	testCases := []struct {
		desc               string
		input              string
		expectedHeaders    url.Values
		expectedBody       string
		expectedStatusCode int
		expectedStatusText string
		expectedAuthSid    string
	}{
		{
			desc:               "Empty response",
			input:              "",
			expectedBody:       "",
			expectedHeaders:    url.Values{},
			expectedStatusText: "",
			expectedStatusCode: -1,
		},
		{
			desc:            "No headers, just body",
			input:           "\n\nTest body",
			expectedBody:    "Test body",
			expectedHeaders: url.Values{},
		},
		{
			desc:            "Headers and body",
			input:           "Test-Header: value\n\nTest body",
			expectedBody:    "Test body",
			expectedHeaders: url.Values{"Test-Header": {"value"}},
		},
		{
			desc:            "Headers and body + Auth-Sid",
			input:           "Test-Header: value\nAuth-Sid: 123\n\nTest body",
			expectedBody:    "Test body",
			expectedHeaders: url.Values{"Test-Header": {"value"}, "Auth-Sid": {"123"}},
			expectedAuthSid: "123",
		},
		{
			desc:               "Headers and body + Status-Text",
			input:              "Test-Header: value\nStatus-Text: OK\n\nTest body",
			expectedBody:       "Test body",
			expectedHeaders:    url.Values{"Test-Header": {"value"}, "Status-Text": {"OK"}},
			expectedStatusText: "OK",
		},
		{
			desc:               "Headers and body + Status-Code",
			input:              "Test-Header: value\nStatus-Code: 2020\n\nTest body",
			expectedBody:       "Test body",
			expectedHeaders:    url.Values{"Test-Header": {"value"}, "Status-Code": {"2020"}},
			expectedStatusCode: 2020,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {

			response := parseJokerResponse(test.input)
			assert.Equal(t, test.expectedHeaders, response.Headers)
			assert.Equal(t, test.expectedBody, response.Body)
			assert.Equal(t, test.expectedAuthSid, response.AuthSid)
			assert.Equal(t, test.expectedStatusText, response.StatusText)

			if test.expectedStatusCode != 0 {
				assert.Equal(t, test.expectedStatusCode, response.StatusCode)
			}
		})
	}
}

func setup() (*http.ServeMux, *httptest.Server) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	return mux, server
}

const (
	correctAuth     = "123"
	incorrectAuth   = "321"
	serverErrorAuth = "500"
)

func TestJokerLogin(t *testing.T) {
	testCases := []struct {
		desc               string
		authKey            string
		expectedError      bool
		expectedStatusCode int
		expectedAuthSid    string
	}{
		{
			desc:               "correct key",
			authKey:            correctAuth,
			expectedStatusCode: 0,
			expectedAuthSid:    correctAuth,
		},
		{
			desc:               "incorrect key",
			authKey:            incorrectAuth,
			expectedStatusCode: 2200,
			expectedError:      true,
		},
		{
			desc:               "server error",
			authKey:            serverErrorAuth,
			expectedStatusCode: -500,
			expectedError:      true,
		},
		{
			desc:               "non-ok status code",
			authKey:            "333",
			expectedStatusCode: 2202,
			expectedError:      true,
		},
	}
	mux, server := setup()
	defer server.Close()

	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		switch r.FormValue("api-key") {
		case correctAuth:
			http.Error(w, "Status-Code: 0\nStatus-Text: OK\nAuth-Sid: 123\n\ncom\nnet", http.StatusOK)
		case incorrectAuth:
			http.Error(w, "Status-Code: 2200\nStatus-Text: Authentication error", http.StatusOK)
		case serverErrorAuth:
			http.Error(w, "Unknown", http.StatusNotFound)
		default:
			http.Error(w, "Status-Code: 2202\nStatus-Text: OK\n\ncom\nnet", http.StatusOK)
		}
	})

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.BaseURL = server.URL
			config.APIKey = test.authKey

			p, err := NewDNSProviderConfig(config)
			require.NoError(t, err)
			require.NotNil(t, p)
			response, err := p.jokerLogin()
			if test.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, response)
				assert.Equal(t, test.expectedStatusCode, response.StatusCode)
				assert.Equal(t, test.expectedAuthSid, response.AuthSid)
			}
		})
	}
}

func TestJokerLogout(t *testing.T) {
	testCases := []struct {
		desc               string
		authSid            string
		expectedError      bool
		expectedStatusCode int
	}{
		{
			desc:               "correct auth-sid",
			authSid:            correctAuth,
			expectedStatusCode: 0,
		},
		{
			desc:               "incorrect auth-sid",
			authSid:            incorrectAuth,
			expectedStatusCode: 2200,
		},
		{
			desc:          "already logged out",
			authSid:       "",
			expectedError: true,
		},
		{
			desc:          "server error",
			authSid:       serverErrorAuth,
			expectedError: true,
		},
	}
	mux, server := setup()
	defer server.Close()

	mux.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		switch r.FormValue("auth-sid") {
		case correctAuth:
			http.Error(w, "Status-Code: 0\nStatus-Text: OK\n", http.StatusOK)
		case incorrectAuth:
			http.Error(w, "Status-Code: 2200\nStatus-Text: Authentication error", http.StatusOK)
		default:
			http.Error(w, "Unknown", http.StatusNotFound)
		}
	})

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.BaseURL = server.URL
			config.APIKey = "12345"
			config.AuthSid = test.authSid

			p, err := NewDNSProviderConfig(config)
			require.NoError(t, err)
			require.NotNil(t, p)
			response, err := p.jokerLogout()
			if test.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, response)
				assert.Equal(t, test.expectedStatusCode, response.StatusCode)
			}
		})
	}
}

func TestJokerGetZone(t *testing.T) {
	const testZone = "@ A 0 192.0.2.2 3600"
	testCases := []struct {
		desc               string
		authSid            string
		domain             string
		zone               string
		expectedError      bool
		expectedStatusCode int
	}{
		{
			desc:               "correct auth-sid, known domain",
			authSid:            correctAuth,
			domain:             "known",
			zone:               testZone,
			expectedStatusCode: 0,
		},
		{
			desc:               "incorrect auth-sid, known domain",
			authSid:            incorrectAuth,
			domain:             "known",
			expectedStatusCode: 2202,
		},
		{
			desc:               "correct auth-sid, unknown domain",
			authSid:            correctAuth,
			domain:             "unknown",
			expectedStatusCode: 2202,
		},
		{
			desc:          "server error",
			authSid:       "500",
			expectedError: true,
		},
	}
	mux, server := setup()
	defer server.Close()

	mux.HandleFunc("/dns-zone-get", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		authSid := r.FormValue("auth-sid")
		domain := r.FormValue("domain")

		switch {
		case authSid == correctAuth && domain == "known":
			http.Error(w, "Status-Code: 0\nStatus-Text: OK\n\n"+testZone, http.StatusOK)
		case authSid == incorrectAuth || (authSid == correctAuth && domain == "unknown"):
			http.Error(w, "Status-Code: 2202\nStatus-Text: Authorization error", http.StatusOK)
		default:
			http.Error(w, "Unknown", http.StatusNotFound)
		}
	})

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.BaseURL = server.URL
			config.APIKey = "12345"
			config.AuthSid = test.authSid

			p, err := NewDNSProviderConfig(config)
			require.NoError(t, err)
			require.NotNil(t, p)
			response, err := p.jokerGetZone(test.domain)
			if test.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, response)
				assert.Equal(t, test.expectedStatusCode, response.StatusCode)
				assert.Equal(t, test.zone, response.Body)
			}
		})
	}
}
