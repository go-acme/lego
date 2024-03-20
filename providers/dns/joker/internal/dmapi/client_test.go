package dmapi

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

func setupTest(t *testing.T) (*http.ServeMux, string) {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	return mux, server.URL
}

func TestClient_GetZone(t *testing.T) {
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

	mux, serverURL := setupTest(t)

	mux.HandleFunc("/dns-zone-get", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)

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
			client := NewClient(AuthInfo{APIKey: "12345"})
			client.BaseURL = serverURL

			response, err := client.GetZone(mockContext(test.authSid), test.domain)
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
		expected *Response
	}{
		{
			desc:  "Empty response",
			input: "",
			expected: &Response{
				Headers:    url.Values{},
				StatusCode: -1,
			},
		},
		{
			desc:  "No headers, just body",
			input: "\n\nTest body",
			expected: &Response{
				Headers:    url.Values{},
				Body:       "Test body",
				StatusCode: -1,
			},
		},
		{
			desc:  "Headers and body",
			input: "Test-Header: value\n\nTest body",
			expected: &Response{
				Headers:    url.Values{"Test-Header": {"value"}},
				Body:       "Test body",
				StatusCode: -1,
			},
		},
		{
			desc:  "Headers and body + Auth-Sid",
			input: "Test-Header: value\nAuth-Sid: 123\n\nTest body",
			expected: &Response{
				Headers:    url.Values{"Test-Header": {"value"}, "Auth-Sid": {"123"}},
				Body:       "Test body",
				StatusCode: -1,
				AuthSid:    "123",
			},
		},
		{
			desc:  "Headers and body + Status-Text",
			input: "Test-Header: value\nStatus-Text: OK\n\nTest body",
			expected: &Response{
				Headers:    url.Values{"Test-Header": {"value"}, "Status-Text": {"OK"}},
				Body:       "Test body",
				StatusText: "OK",
				StatusCode: -1,
			},
		},
		{
			desc:  "Headers and body + Status-Code",
			input: "Test-Header: value\nStatus-Code: 2020\n\nTest body",
			expected: &Response{
				Headers:    url.Values{"Test-Header": {"value"}, "Status-Code": {"2020"}},
				Body:       "Test body",
				StatusCode: 2020,
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			response := parseResponse(test.input)

			assert.Equal(t, test.expected, response)
		})
	}
}

func Test_RemoveTxtEntryFromZone(t *testing.T) {
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
			t.Parallel()

			zone, modified := RemoveTxtEntryFromZone(test.input, "_acme-challenge")
			assert.Equal(t, test.expected, zone)
			assert.Equal(t, test.modified, modified)
		})
	}
}

func Test_AddTxtEntryToZone(t *testing.T) {
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
			zone := AddTxtEntryToZone(test.input, "_acme-challenge", "test", 120)
			assert.Equal(t, test.expected, zone)
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
			assert.Equal(t, test.expected, line)
		})
	}
}
