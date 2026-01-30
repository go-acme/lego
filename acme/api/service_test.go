package api

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_getLink(t *testing.T) {
	testCases := []struct {
		desc     string
		header   http.Header
		relName  string
		expected string
	}{
		{
			desc: "success",
			header: http.Header{
				"Link": []string{`<https://acme-staging-v02.api.letsencrypt.org/next>; rel="next", <https://acme-staging-v02.api.letsencrypt.org/up?query>; rel="up"`},
			},
			relName:  "up",
			expected: "https://acme-staging-v02.api.letsencrypt.org/up?query",
		},
		{
			desc: "success several lines",
			header: http.Header{
				"Link": []string{`<https://acme-staging-v02.api.letsencrypt.org/next>; rel="next"`, `<https://acme-staging-v02.api.letsencrypt.org/up?query>; rel="up"`},
			},
			relName:  "up",
			expected: "https://acme-staging-v02.api.letsencrypt.org/up?query",
		},
		{
			desc:     "no link",
			header:   http.Header{},
			relName:  "up",
			expected: "",
		},
		{
			desc:     "no header",
			relName:  "up",
			expected: "",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			link := getLink(test.header, test.relName)

			assert.Equal(t, test.expected, link)
		})
	}
}

func TestParseRetryAfter(t *testing.T) {
	testCases := []struct {
		desc     string
		value    string
		expected time.Duration
	}{
		{
			desc:     "empty header value",
			value:    "",
			expected: time.Duration(0),
		},
		{
			desc:     "delay-seconds",
			value:    "123",
			expected: 123 * time.Second,
		},
		{
			desc:     "HTTP-date",
			value:    time.Now().Add(3 * time.Second).Format(time.RFC1123),
			expected: 3 * time.Second,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			rt, err := ParseRetryAfter(test.value)
			require.NoError(t, err)

			assert.InDelta(t, test.expected.Seconds(), rt.Seconds(), 0.9)
		})
	}
}
