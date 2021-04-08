package vinyldns

import (
	"testing"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/stretchr/testify/require"
)

const (
	liveKeyauth = "foo"
	liveToken   = "bar"
)

func TestLiveTTL(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()

	provider, err := NewDNSProvider()
	require.NoError(t, err)

	domain := envTest.GetDomain()
	fqdn, value := dns01.GetRecord(domain, liveKeyauth)

	err = provider.Present(domain, liveToken, liveKeyauth)
	require.NoError(t, err)

	_, domainName, err := provider.fqdnSplit(fqdn)
	require.NoError(t, err)

	defer func() {
		errC := provider.CleanUp(domain, liveToken, liveKeyauth)
		if errC != nil {
			t.Log(errC)
		}
	}()

	_, err = provider.getZone(domainName)
	require.NoError(t, err)

	recordset, err := provider.getExistingRecordSet(fqdn)
	require.NoError(t, err)
	require.NotEqual(t, "", recordset.Name)

	var found bool
	for _, i := range recordset.Records {
		if i.Text == value {
			found = true
		}
	}
	require.Equal(t, true, found)
}
