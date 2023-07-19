package dns01

import (
	"strings"
	"testing"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
)

func Test_updateDomainWithCName_caseInsensitive(t *testing.T) {
	qname := "_acme-challenge.uppercase-test.example.com."
	cnameTarget := "_acme-challenge.uppercase-test.cname-target.example.com."

	msg := &dns.Msg{
		MsgHdr: dns.MsgHdr{
			Authoritative: true,
		},
		Answer: []dns.RR{
			&dns.CNAME{
				Hdr: dns.RR_Header{
					Name:   strings.ToUpper(qname), // CNAME names are case-insensitive
					Rrtype: dns.TypeCNAME,
					Class:  dns.ClassINET,
					Ttl:    3600,
				},
				Target: cnameTarget,
			},
		},
	}

	fqdn := updateDomainWithCName(msg, qname)

	assert.Equal(t, cnameTarget, fqdn)
}
