package dns01

import (
	"strings"
	"testing"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
)

func TestCnameCaseInsensitive(t *testing.T) {
	const qname = "_acme-challenge.uppercase-test.example.com."
	const cnameTarget = "_acme-challenge.uppercase-test.cname-target.example.com."
	msg := new(dns.Msg)
	msg.Authoritative = true
	msg.Answer = []dns.RR{
		&dns.CNAME{
			Hdr: dns.RR_Header{
				Name:   strings.ToUpper(qname), // CNAME names are case-insensitive
				Rrtype: dns.TypeCNAME,
				Class:  dns.ClassINET,
				Ttl:    3600,
			},
			Target: cnameTarget,
		},
	}
	fqdn := updateDomainWithCName(msg, qname)
	assert.Equal(t, cnameTarget, fqdn)
}
