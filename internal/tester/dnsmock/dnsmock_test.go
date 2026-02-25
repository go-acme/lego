package dnsmock

import (
	"testing"
	"time"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServer_Query_matchType(t *testing.T) {
	addr := NewServer().
		Query("example.com. SOA", Noop).
		Build(t)

	client := &dns.Client{Timeout: 1 * time.Second}

	m := new(dns.Msg).SetQuestion("example.com.", dns.TypeSOA)

	r, _, err := client.Exchange(m, addr.String())
	require.NoError(t, err)

	require.Equalf(t, dns.RcodeSuccess, r.Rcode,
		"expected %s, got %s", dns.RcodeToString[dns.RcodeSuccess], dns.RcodeToString[r.Rcode])
	assert.Equal(t, m.Question, r.Question)
}

func TestServer_Query_noType(t *testing.T) {
	addr := NewServer().
		Query("example.com.", Noop).
		Build(t)

	client := &dns.Client{Timeout: 1 * time.Second}

	m := new(dns.Msg).SetQuestion("example.com.", dns.TypeSOA)

	r, _, err := client.Exchange(m, addr.String())
	require.NoError(t, err)

	require.Equalf(t, dns.RcodeSuccess, r.Rcode,
		"expected %s, got %s", dns.RcodeToString[dns.RcodeSuccess], dns.RcodeToString[r.Rcode])
	assert.Equal(t, m.Question, r.Question)
}

func TestServer_Query_noMatch_domain(t *testing.T) {
	addr := NewServer().
		Query("example.com. SOA", Noop).
		Build(t)

	client := &dns.Client{Timeout: 1 * time.Second}

	m := new(dns.Msg).SetQuestion("example.org.", dns.TypeSOA)

	r, _, err := client.Exchange(m, addr.String())
	require.NoError(t, err)

	require.Equalf(t, dns.RcodeRefused, r.Rcode,
		"expected %s, got %s", dns.RcodeToString[dns.RcodeRefused], dns.RcodeToString[r.Rcode])
	assert.Equal(t, m.Question, r.Question)
}

func TestServer_Query_noMatch_type(t *testing.T) {
	addr := NewServer().
		Query("example.com. SOA", Noop).
		Build(t)

	client := &dns.Client{Timeout: 1 * time.Second}

	m := new(dns.Msg).SetQuestion("example.com.", dns.TypeTXT)

	r, _, err := client.Exchange(m, addr.String())
	require.NoError(t, err)

	require.Equalf(t, dns.RcodeNotImplemented, r.Rcode,
		"expected %s, got %s", dns.RcodeToString[dns.RcodeNotImplemented], dns.RcodeToString[r.Rcode])
	assert.Equal(t, m.Question, r.Question)
}

func TestServer_Query_noMatch_opType(t *testing.T) {
	addr := NewServer().
		Query("example.com.", Noop).
		Build(t)

	client := &dns.Client{Timeout: 1 * time.Second}

	m := new(dns.Msg).SetUpdate("example.com.")
	m.Insert([]dns.RR{
		&dns.TXT{
			Hdr: dns.RR_Header{Name: "example.com.", Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 1},
			Txt: []string{"foo"},
		},
	})

	r, _, err := client.Exchange(m, addr.String())
	require.NoError(t, err)

	require.Equalf(t, dns.RcodeNotImplemented, r.Rcode,
		"expected %s, got %s", dns.RcodeToString[dns.RcodeNotImplemented], dns.RcodeToString[r.Rcode])
	assert.Equal(t, m.Question, r.Question)
}

func TestServer_Query_unknownType(t *testing.T) {
	assert.PanicsWithValue(t, "QUERY: unknown type: ABC", func() {
		NewServer().
			Query("example.com. ABC", Noop).
			Build(t)
	})
}

func TestServer_Query_duplicate(t *testing.T) {
	assert.PanicsWithValue(t, "QUERY: duplicate route: example.com. SOA", func() {
		NewServer().
			Query("example.com. SOA", Noop).
			Query("example.com. SOA", Noop).
			Build(t)
	})
}

func TestServer_Query_duplicateGlobal(t *testing.T) {
	assert.PanicsWithValue(t, "QUERY: a global route already exists for the domain: example.com.", func() {
		NewServer().
			Query("example.com.", Noop).
			Query("example.com.", Noop).
			Build(t)
	})
}

func TestServer_Query_mixed(t *testing.T) {
	assert.PanicsWithValue(t, "QUERY: global route and specific routes cannot be mixed for the same domain: example.com.", func() {
		NewServer().
			Query("example.com. SOA", Noop).
			Query("example.com.", Noop).
			Build(t)
	})
}

func TestServer_Query_invalidDomain(t *testing.T) {
	assert.PanicsWithValue(t, "QUERY: invalid domain: .example.com.", func() {
		NewServer().
			Query(".example.com. SOA", Noop).
			Build(t)
	})
}

func TestServer_Query_invalidPattern(t *testing.T) {
	assert.PanicsWithValue(t, "QUERY: invalid pattern: example.com. SOA 13", func() {
		NewServer().
			Query("example.com. SOA 13", Noop).
			Build(t)
	})
}

func TestServer_Update(t *testing.T) {
	addr := NewServer().
		Update("example.com.", Noop).
		Build(t)

	client := &dns.Client{Timeout: 1 * time.Second}

	m := new(dns.Msg).SetUpdate("example.com.")
	m.Insert([]dns.RR{
		&dns.TXT{
			Hdr: dns.RR_Header{Name: "example.com.", Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 1},
			Txt: []string{"foo"},
		},
	})

	r, _, err := client.Exchange(m, addr.String())
	require.NoError(t, err)

	require.Equalf(t, dns.RcodeSuccess, r.Rcode,
		"expected %s, got %s", dns.RcodeToString[dns.RcodeSuccess], dns.RcodeToString[r.Rcode])
	assert.Equal(t, m.Question, r.Question)
}

func TestServer_Update_noMatch_domain(t *testing.T) {
	addr := NewServer().
		Update("example.com.", Noop).
		Build(t)

	client := &dns.Client{Timeout: 1 * time.Second}

	m := new(dns.Msg).SetUpdate("example.org.")
	m.Insert([]dns.RR{
		&dns.TXT{
			Hdr: dns.RR_Header{Name: "example.org.", Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 1},
			Txt: []string{"foo"},
		},
	})

	r, _, err := client.Exchange(m, addr.String())
	require.NoError(t, err)

	require.Equalf(t, dns.RcodeRefused, r.Rcode,
		"expected %s, got %s", dns.RcodeToString[dns.RcodeRefused], dns.RcodeToString[r.Rcode])
	assert.Equal(t, m.Question, r.Question)
}

func TestServer_Update_noMatch_opType(t *testing.T) {
	addr := NewServer().
		Update("example.com.", Noop).
		Build(t)

	client := &dns.Client{Timeout: 1 * time.Second}

	m := new(dns.Msg).SetQuestion("example.com.", dns.TypeTXT)

	r, _, err := client.Exchange(m, addr.String())
	require.NoError(t, err)

	require.Equalf(t, dns.RcodeNotImplemented, r.Rcode,
		"expected %s, got %s", dns.RcodeToString[dns.RcodeNotImplemented], dns.RcodeToString[r.Rcode])
	assert.Equal(t, m.Question, r.Question)
}

func TestServer_Update_duplicate(t *testing.T) {
	assert.PanicsWithValue(t, "UPDATE: a global route already exists for the domain: example.com.", func() {
		NewServer().
			Update("example.com.", Noop).
			Update("example.com.", Noop).
			Build(t)
	})
}

func TestServer_Update_invalidDomain(t *testing.T) {
	assert.PanicsWithValue(t, "UPDATE: invalid domain: .example.com.", func() {
		NewServer().
			Update(".example.com.", Noop).
			Build(t)
	})
}

func TestServer_Update_invalidPattern(t *testing.T) {
	assert.PanicsWithValue(t, "UPDATE: invalid pattern: example.com. SOA 13", func() {
		NewServer().
			Update("example.com. SOA 13", Noop).
			Build(t)
	})
}
