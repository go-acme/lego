package dnsmock

import (
	"testing"
	"time"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSOA_self(t *testing.T) {
	addr := NewServer().
		Query("example.com. SOA", SOA("")).
		Build(t)

	client := &dns.Client{Timeout: 1 * time.Second}

	m := new(dns.Msg).SetQuestion("example.com.", dns.TypeSOA)

	r, _, err := client.Exchange(m, addr.String())
	require.NoError(t, err)

	expectedSOA := []dns.RR{&dns.SOA{
		Hdr:     dns.RR_Header{Name: "example.com.", Rrtype: dns.TypeSOA, Class: dns.ClassINET, Ttl: 120, Rdlength: 56},
		Ns:      "ns1.example.com.",
		Mbox:    "admin.example.com.",
		Serial:  2016022801,
		Refresh: 28800,
		Retry:   7200,
		Expire:  2419200,
		Minttl:  1200,
	}}

	require.Equal(t, dns.RcodeSuccess, r.Rcode)
	assert.Equal(t, expectedSOA, r.Answer)
	assert.Equal(t, m.Question, r.Question)
}

func TestSOA_differentDomain(t *testing.T) {
	addr := NewServer().
		Query("example.com. SOA", SOA("example.org.")).
		Build(t)

	client := &dns.Client{Timeout: 1 * time.Second}

	m := new(dns.Msg).SetQuestion("example.com.", dns.TypeSOA)

	r, _, err := client.Exchange(m, addr.String())
	require.NoError(t, err)

	require.Equalf(t, dns.RcodeSuccess, r.Rcode,
		"expected %s, got %s", dns.RcodeToString[dns.RcodeSuccess], dns.RcodeToString[r.Rcode])

	expectedSOA := []dns.RR{&dns.SOA{
		Hdr:     dns.RR_Header{Name: "example.org.", Rrtype: dns.TypeSOA, Class: dns.ClassINET, Ttl: 120, Rdlength: 56},
		Ns:      "ns1.example.org.",
		Mbox:    "admin.example.org.",
		Serial:  2016022801,
		Refresh: 28800,
		Retry:   7200,
		Expire:  2419200,
		Minttl:  1200,
	}}

	assert.Equal(t, expectedSOA, r.Answer)
	assert.Equal(t, m.Question, r.Question)
}

func TestSOA_tld(t *testing.T) {
	addr := NewServer().
		Query("com. SOA", SOA("")).
		Build(t)

	client := &dns.Client{Timeout: 1 * time.Second}

	m := new(dns.Msg).SetQuestion("com.", dns.TypeSOA)

	r, _, err := client.Exchange(m, addr.String())
	require.NoError(t, err)

	require.Equalf(t, dns.RcodeSuccess, r.Rcode,
		"expected %s, got %s", dns.RcodeToString[dns.RcodeSuccess], dns.RcodeToString[r.Rcode])

	expectedSOA := []dns.RR{&dns.SOA{
		Hdr:     dns.RR_Header{Name: "com.", Rrtype: dns.TypeSOA, Class: dns.ClassINET, Ttl: 120, Rdlength: 48},
		Ns:      "ns1.nic.com.",
		Mbox:    "admin.nic.com.",
		Serial:  2016022801,
		Refresh: 28800,
		Retry:   7200,
		Expire:  2419200,
		Minttl:  1200,
	}}

	assert.Equal(t, expectedSOA, r.Answer)
	assert.Equal(t, m.Question, r.Question)
}

func TestCNAME(t *testing.T) {
	addr := NewServer().
		Query("example.com. CNAME", CNAME("example.org.")).
		Build(t)

	client := &dns.Client{Timeout: 1 * time.Second}

	m := new(dns.Msg).SetQuestion("example.com.", dns.TypeCNAME)

	r, _, err := client.Exchange(m, addr.String())
	require.NoError(t, err)

	require.Equalf(t, dns.RcodeSuccess, r.Rcode,
		"expected %s, got %s", dns.RcodeToString[dns.RcodeSuccess], dns.RcodeToString[r.Rcode])

	expectedCNAME := []dns.RR{&dns.CNAME{
		Hdr:    dns.RR_Header{Name: "example.com.", Rrtype: dns.TypeCNAME, Class: dns.ClassINET, Ttl: 1, Rdlength: 13},
		Target: "example.org.",
	}}

	assert.Equal(t, expectedCNAME, r.Answer)
	assert.Equal(t, m.Question, r.Question)
}

func TestNoop(t *testing.T) {
	addr := NewServer().
		Query("example.com. CNAME", Noop).
		Build(t)

	client := &dns.Client{Timeout: 1 * time.Second}

	m := new(dns.Msg).SetQuestion("example.com.", dns.TypeCNAME)

	r, _, err := client.Exchange(m, addr.String())
	require.NoError(t, err)

	require.Equalf(t, dns.RcodeSuccess, r.Rcode,
		"expected %s, got %s", dns.RcodeToString[dns.RcodeSuccess], dns.RcodeToString[r.Rcode])
	assert.Equal(t, m.Question, r.Question)
}

func TestError(t *testing.T) {
	addr := NewServer().
		Query("example.com. CNAME", Error(dns.RcodeNameError)).
		Build(t)

	client := &dns.Client{Timeout: 1 * time.Second}

	m := new(dns.Msg).SetQuestion("example.com.", dns.TypeCNAME)

	r, _, err := client.Exchange(m, addr.String())
	require.NoError(t, err)

	require.Equalf(t, dns.RcodeNameError, r.Rcode,
		"expected %s, got %s", dns.RcodeToString[dns.RcodeNameError], dns.RcodeToString[r.Rcode])
	assert.Equal(t, m.Question, r.Question)
}
