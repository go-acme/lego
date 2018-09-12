package iij

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	apiAccessKeyEnv  string
	apiSecretKeyEnv  string
	doServiceCodeEnv string
	testDomain       string
	liveTest         bool
)

func init() {
	apiAccessKeyEnv = os.Getenv("IIJ_API_ACCESS_KEY")
	apiSecretKeyEnv = os.Getenv("IIJ_API_SECRET_KEY")
	doServiceCodeEnv = os.Getenv("IIJ_DO_SERVICE_CODE")

	testDomain = os.Getenv("IIJ_API_TESTDOMAIN")

	if len(apiAccessKeyEnv) > 0 && len(apiSecretKeyEnv) > 0 && len(doServiceCodeEnv) > 0 && len(testDomain) > 0 {
		liveTest = true
	}
}

func restoreEnv() {
	os.Setenv("IIJ_API_ACCESS_KEY", apiAccessKeyEnv)
	os.Setenv("IIJ_API_SECRET_KEY", apiSecretKeyEnv)
	os.Setenv("IIJ_DO_SERVICE_CODE", doServiceCodeEnv)
	os.Setenv("IIJ_API_TESTDOMAIN", testDomain)
}

func TestSplitDomain(t *testing.T) {
	testCases := []struct {
		desc          string
		domain        string
		zones         []string
		expectedOwner string
		expectedZone  string
	}{
		{
			desc:          "domain equals zone",
			domain:        "domain.com",
			zones:         []string{"domain.com"},
			expectedOwner: "_acme-challenge",
			expectedZone:  "domain.com",
		},
		{
			desc:          "with a sub domain",
			domain:        "my.domain.com",
			zones:         []string{"domain.com"},
			expectedOwner: "_acme-challenge.my",
			expectedZone:  "domain.com",
		},
		{
			desc:          "with a sub domain in a zone",
			domain:        "my.sub.domain.com",
			zones:         []string{"sub.domain.com", "domain.com"},
			expectedOwner: "_acme-challenge.my",
			expectedZone:  "sub.domain.com",
		},
		{
			desc:          "with a sub sub domain",
			domain:        "my.sub.domain.com",
			zones:         []string{"domain1.com", "domain.com"},
			expectedOwner: "_acme-challenge.my.sub",
			expectedZone:  "domain.com",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			owner, zone, err := splitDomain(test.domain, test.zones)
			require.NoError(t, err)

			assert.Equal(t, test.expectedOwner, owner)
			assert.Equal(t, test.expectedZone, zone)
		})
	}

}

func TestNewDNSProviderMissingCredErr(t *testing.T) {
	defer restoreEnv()
	os.Setenv("IIJ_API_ACCESS_KEY", "")
	os.Setenv("IIJ_API_SECRET_KEY", "")
	os.Setenv("IIJ_DO_SERVICE_CODE", "")

	_, err := NewDNSProvider()
	assert.EqualError(t, err, "iij: some credentials information are missing: IIJ_API_ACCESS_KEY,IIJ_API_SECRET_KEY,IIJ_DO_SERVICE_CODE")
}

func TestNewDNSProvider(t *testing.T) {
	if !liveTest {
		t.Skip("skipping live test")
	}

	_, err := NewDNSProvider()
	assert.NoError(t, err)
}

func TestDNSProvider_Present(t *testing.T) {
	if !liveTest {
		t.Skip("skipping live test")
	}

	provider, err := NewDNSProvider()
	assert.NoError(t, err)

	err = provider.Present(testDomain, "", "123d==")
	assert.NoError(t, err)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	if !liveTest {
		t.Skip("skipping live test")
	}

	provider, err := NewDNSProvider()
	assert.NoError(t, err)

	err = provider.CleanUp(testDomain, "", "123d==")
	assert.NoError(t, err)
}
