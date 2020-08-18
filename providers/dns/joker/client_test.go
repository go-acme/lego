package joker

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	correctAPIKey     = "123"
	incorrectAPIKey   = "321"
	serverErrorAPIKey = "500"
)

const (
	correctUsername     = "lego"
	incorrectUsername   = "not_lego"
	serverErrorUsername = "error"
)

func setup() (*http.ServeMux, *httptest.Server) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	return mux, server
}

func TestDNSProvider_login_api_key(t *testing.T) {
	testCases := []struct {
		desc               string
		apiKey             string
		expectedError      bool
		expectedStatusCode int
		expectedAuthSid    string
	}{
		{
			desc:               "correct key",
			apiKey:             correctAPIKey,
			expectedStatusCode: 0,
			expectedAuthSid:    correctAPIKey,
		},
		{
			desc:               "incorrect key",
			apiKey:             incorrectAPIKey,
			expectedStatusCode: 2200,
			expectedError:      true,
		},
		{
			desc:               "server error",
			apiKey:             serverErrorAPIKey,
			expectedStatusCode: -500,
			expectedError:      true,
		},
		{
			desc:               "non-ok status code",
			apiKey:             "333",
			expectedStatusCode: 2202,
			expectedError:      true,
		},
	}

	mux, server := setup()
	defer server.Close()

	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "POST", r.Method)

		switch r.FormValue("api-key") {
		case correctAPIKey:
			_, _ = io.WriteString(w, "Status-Code: 0\nStatus-Text: OK\nAuth-Sid: 123\n\ncom\nnet")
		case incorrectAPIKey:
			_, _ = io.WriteString(w, "Status-Code: 2200\nStatus-Text: Authentication error")
		case serverErrorAPIKey:
			http.NotFound(w, r)
		default:
			_, _ = io.WriteString(w, "Status-Code: 2202\nStatus-Text: OK\n\ncom\nnet")
		}
	})

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig(nil)
			config.BaseURL = server.URL
			config.APIKey = test.apiKey

			p, err := NewDNSProviderConfig(config)
			require.NoError(t, err)
			require.NotNil(t, p)

			response, err := p.login()
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

func TestDNSProvider_login_username(t *testing.T) {
	testCases := []struct {
		desc               string
		username           string
		password           string
		expectedError      bool
		expectedStatusCode int
		expectedAuthSid    string
	}{
		{
			desc:               "correct username and password",
			username:           correctUsername,
			password:           "go-acme",
			expectedError:      false,
			expectedStatusCode: 0,
			expectedAuthSid:    correctAPIKey,
		},
		{
			desc:               "incorrect username",
			username:           incorrectUsername,
			password:           "go-acme",
			expectedStatusCode: 2200,
			expectedError:      true,
		},
		{
			desc:               "server error",
			username:           serverErrorUsername,
			password:           "go-acme",
			expectedStatusCode: -500,
			expectedError:      true,
		},
		{
			desc:               "non-ok status code",
			username:           "random",
			password:           "go-acme",
			expectedStatusCode: 2202,
			expectedError:      true,
		},
	}

	mux, server := setup()
	defer server.Close()

	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "POST", r.Method)

		switch r.FormValue("username") {
		case correctUsername:
			_, _ = io.WriteString(w, "Status-Code: 0\nStatus-Text: OK\nAuth-Sid: 123\n\ncom\nnet")
		case incorrectUsername:
			_, _ = io.WriteString(w, "Status-Code: 2200\nStatus-Text: Authentication error")
		case serverErrorUsername:
			http.NotFound(w, r)
		default:
			_, _ = io.WriteString(w, "Status-Code: 2202\nStatus-Text: OK\n\ncom\nnet")
		}
	})

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig(nil)
			config.BaseURL = server.URL
			config.Username = test.username
			config.Password = test.password

			p, err := NewDNSProviderConfig(config)
			require.NoError(t, err)
			require.NotNil(t, p)

			response, err := p.login()
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

func TestDNSProvider_logout(t *testing.T) {
	testCases := []struct {
		desc               string
		authSid            string
		expectedError      bool
		expectedStatusCode int
	}{
		{
			desc:               "correct auth-sid",
			authSid:            correctAPIKey,
			expectedStatusCode: 0,
		},
		{
			desc:               "incorrect auth-sid",
			authSid:            incorrectAPIKey,
			expectedStatusCode: 2200,
		},
		{
			desc:          "already logged out",
			authSid:       "",
			expectedError: true,
		},
		{
			desc:          "server error",
			authSid:       serverErrorAPIKey,
			expectedError: true,
		},
	}

	mux, server := setup()
	defer server.Close()

	mux.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "POST", r.Method)

		switch r.FormValue("auth-sid") {
		case correctAPIKey:
			_, _ = io.WriteString(w, "Status-Code: 0\nStatus-Text: OK\n")
		case incorrectAPIKey:
			_, _ = io.WriteString(w, "Status-Code: 2200\nStatus-Text: Authentication error")
		default:
			http.NotFound(w, r)
		}
	})

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig(nil)
			config.BaseURL = server.URL
			config.APIKey = "12345"
			config.AuthSid = test.authSid

			p, err := NewDNSProviderConfig(config)
			require.NoError(t, err)
			require.NotNil(t, p)

			response, err := p.logout()
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

func TestDNSProvider_getZone(t *testing.T) {
	testZone := "@ A 0 192.0.2.2 3600"

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
			authSid:            correctAPIKey,
			domain:             "known",
			zone:               testZone,
			expectedStatusCode: 0,
		},
		{
			desc:               "incorrect auth-sid, known domain",
			authSid:            incorrectAPIKey,
			domain:             "known",
			expectedStatusCode: 2202,
		},
		{
			desc:               "correct auth-sid, unknown domain",
			authSid:            correctAPIKey,
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
		require.Equal(t, "POST", r.Method)

		authSid := r.FormValue("auth-sid")
		domain := r.FormValue("domain")

		switch {
		case authSid == correctAPIKey && domain == "known":
			_, _ = io.WriteString(w, "Status-Code: 0\nStatus-Text: OK\n\n"+testZone)
		case authSid == incorrectAPIKey || (authSid == correctAPIKey && domain == "unknown"):
			_, _ = io.WriteString(w, "Status-Code: 2202\nStatus-Text: Authorization error")
		default:
			http.NotFound(w, r)
		}
	})

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig(nil)
			config.BaseURL = server.URL
			config.APIKey = "12345"
			config.AuthSid = test.authSid

			p, err := NewDNSProviderConfig(config)
			require.NoError(t, err)
			require.NotNil(t, p)

			response, err := p.getZone(test.domain)
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

func Test_parseResponse(t *testing.T) {
	testCases := []struct {
		desc     string
		input    string
		expected *response
	}{
		{
			desc:  "Empty response",
			input: "",
			expected: &response{
				Headers:    url.Values{},
				StatusCode: -1,
			},
		},
		{
			desc:  "No headers, just body",
			input: "\n\nTest body",
			expected: &response{
				Headers:    url.Values{},
				Body:       "Test body",
				StatusCode: -1,
			},
		},
		{
			desc:  "Headers and body",
			input: "Test-Header: value\n\nTest body",
			expected: &response{
				Headers:    url.Values{"Test-Header": {"value"}},
				Body:       "Test body",
				StatusCode: -1,
			},
		},
		{
			desc:  "Headers and body + Auth-Sid",
			input: "Test-Header: value\nAuth-Sid: 123\n\nTest body",
			expected: &response{
				Headers:    url.Values{"Test-Header": {"value"}, "Auth-Sid": {"123"}},
				Body:       "Test body",
				StatusCode: -1,
				AuthSid:    "123",
			},
		},
		{
			desc:  "Headers and body + Status-Text",
			input: "Test-Header: value\nStatus-Text: OK\n\nTest body",
			expected: &response{
				Headers:    url.Values{"Test-Header": {"value"}, "Status-Text": {"OK"}},
				Body:       "Test body",
				StatusText: "OK",
				StatusCode: -1,
			},
		},
		{
			desc:  "Headers and body + Status-Code",
			input: "Test-Header: value\nStatus-Code: 2020\n\nTest body",
			expected: &response{
				Headers:    url.Values{"Test-Header": {"value"}, "Status-Code": {"2020"}},
				Body:       "Test body",
				StatusCode: 2020,
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			response := parseResponse(test.input)

			assert.Equal(t, test.expected, response)
		})
	}
}

func Test_removeTxtEntryFromZone(t *testing.T) {
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
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			zone, modified := removeTxtEntryFromZone(test.input, "_acme-challenge")
			assert.Equal(t, zone, test.expected)
			assert.Equal(t, modified, test.modified)
		})
	}
}

func Test_addTxtEntryToZone(t *testing.T) {
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
			desc:     "zone with required cleanup entry",
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

func Test_fixTxtLines(t *testing.T) {
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
