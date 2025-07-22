package edgedns

import (
	"context"
	"fmt"
	"testing"
	"time"

	edgegriddns "github.com/akamai/AkamaiOPEN-edgegrid-golang/v11/pkg/dns"
	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v11/pkg/session"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLivePresent(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()
	provider, err := NewDNSProvider()
	require.NoError(t, err)

	err = provider.Present(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)

	// Present Twice to handle create / update
	err = provider.Present(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}

func TestLiveCleanUp(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()
	provider, err := NewDNSProvider()
	require.NoError(t, err)

	time.Sleep(1 * time.Second)

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}

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
	zone, err := getZone(fqdn)
	require.NoError(t, err)

	ctx := context.Background()

	sess, err := session.New(session.WithSigner(provider.config))
	require.NoError(t, err)

	client := edgegriddns.Client(sess)

	resourceRecordSets, err := client.GetRecordList(ctx, edgegriddns.GetRecordListRequest{
		Zone:       zone,
		RecordType: "TXT",
	})

	require.NoError(t, err)

	for i, rrset := range resourceRecordSets.RecordSets {
		if rrset.Name != fqdn {
			continue
		}

		t.Run(fmt.Sprintf("testing record set %d", i), func(t *testing.T) {
			assert.Equal(t, fqdn, rrset.Name)
			assert.Equal(t, "TXT", rrset.Type)
			assert.Equal(t, dns01.DefaultTTL, rrset.TTL)
		})
	}
}
