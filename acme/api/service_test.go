package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
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
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			link := getLink(test.header, test.relName)

			assert.Equal(t, test.expected, link)
		})
	}
}
