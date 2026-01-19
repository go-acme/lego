package dns01

import (
	"errors"
	"testing"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
)

func TestDNSError_Error(t *testing.T) {
	msgIn := createDNSMsg("example.com.", dns.TypeTXT, true)

	msgOut := createDNSMsg("example.org.", dns.TypeSOA, true)
	msgOut.Rcode = dns.RcodeNameError

	testCases := []struct {
		desc     string
		err      *DNSError
		expected string
	}{
		{
			desc:     "empty error",
			err:      &DNSError{},
			expected: "DNS error",
		},
		{
			desc: "all fields",
			err: &DNSError{
				Message: "Oops",
				NS:      "example.com.",
				MsgIn:   msgIn,
				MsgOut:  msgOut,
				Err:     errors.New("I did it again"),
			},
			expected: "Oops: I did it again [ns=example.com., question='example.com. IN  TXT', code=NXDOMAIN]",
		},
		{
			desc: "only NS",
			err: &DNSError{
				NS: "example.com.",
			},
			expected: "DNS error [ns=example.com.]",
		},
		{
			desc: "only MsgIn",
			err: &DNSError{
				MsgIn: msgIn,
			},
			expected: "DNS error [question='example.com. IN  TXT']",
		},
		{
			desc: "only MsgOut",
			err: &DNSError{
				MsgOut: msgOut,
			},
			expected: "DNS error [question='example.org. IN  SOA', code=NXDOMAIN]",
		},
		{
			desc: "only Err",
			err: &DNSError{
				Err: errors.New("I did it again"),
			},
			expected: "DNS error: I did it again",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			assert.EqualError(t, test.err, test.expected)
		})
	}
}
