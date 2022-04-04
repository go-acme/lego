package localserver

import (
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester"
)

var envTest = tester.NewEnvTest(EnvListen)
var listenAddr = "127.0.0.1:5352"

func TestDnsProvider(t *testing.T) {
	envTest.Apply(map[string]string{EnvListen: listenAddr})
	provider, err := NewDNSProvider()
	assert.Nil(t, err)

	domain := "foo.com"
	keyAuth := "12d=="

	fqdn, txtExpectedValue := dns01.GetRecord(domain, keyAuth)
	err = provider.Present(domain, "", keyAuth)
	assert.Nil(t, err)

	r, err := makeDnsQuery(fqdn, dns.TypeTXT, listenAddr)
	assert.Nil(t, err)
	assert.Equal(t, txtExpectedValue, r.Answer[0].(*dns.TXT).Txt[0])

	r, err = makeDnsQuery("bar.com.", dns.TypeTXT, listenAddr)
	assert.Nil(t, err)
	assert.Equal(t, dns.RcodeNameError, r.MsgHdr.Rcode)

	err = provider.CleanUp(domain, "", keyAuth)
	assert.Nil(t, err)
}

func makeDnsQuery(fqdn string, dnsType uint16, dnsServer string) (*dns.Msg, error) {
	c := new(dns.Client)
	m := new(dns.Msg)
	m.SetQuestion(fqdn, dnsType)
	r, _, err := c.Exchange(m, dnsServer)
	return r, err
}
