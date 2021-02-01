package edgedns

import (
	"fmt"
	"testing"

	configdns "github.com/akamai/AkamaiOPEN-edgegrid-golang/configdns-v2"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLiveTTL(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()

	provider, err := NewDNSProvider()
	require.NoError(t, err)

	domain := envTest.GetDomain()

	err = provider.Present(domain, "foo", "bar")
	require.NoError(t, err)

	defer func() {
		e := provider.CleanUp(domain, "foo", "bar")
		if e != nil {
			t.Log(e)
		}
	}()

	fqdn := "_acme-challenge." + domain + "."
	zone, err := findZone(domain)
	require.NoError(t, err)

	resourceRecordSets, err := configdns.GetRecordList(zone, fqdn, "TXT")
	require.NoError(t, err)

	for i, rrset := range resourceRecordSets.Recordsets {
		if rrset.Name != fqdn {
			continue
		}

		t.Run(fmt.Sprintf("testing record set %d", i), func(t *testing.T) {
			assert.Equal(t, rrset.Name, fqdn)
			assert.Equal(t, rrset.Type, "TXT")
			assert.Equal(t, rrset.TTL, dns01.DefaultTTL)
		})
	}
}
