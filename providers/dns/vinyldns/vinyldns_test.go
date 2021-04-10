package vinyldns

import (
	"net/http"
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

const (
	targetHostname    = "host"
	targetDomain      = "vinyldns.io"
	zoneID            = "00000000-0000-0000-0000-000000000000"
	newRecordSetID    = "11000000-0000-0000-0000-000000000000"
	newCreateChangeID = "20000000-0000-0000-0000-000000000000"
	recordID          = "30000000-0000-0000-0000-000000000000"
)

var envTest = tester.NewEnvTest(
	EnvAccessKey,
	EnvSecretKey,
	EnvHost).
	WithDomain(envDomain)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvAccessKey: "123",
				EnvSecretKey: "456",
				EnvHost:      "https://example.com",
			},
		},
		{
			desc: "missing all credentials",
			envVars: map[string]string{
				EnvHost: "https://example.com",
			},
			expected: "vinyldns: some credentials information are missing: VINYLDNS_ACCESS_KEY,VINYLDNS_SECRET_KEY",
		},
		{
			desc: "missing access key",
			envVars: map[string]string{
				EnvSecretKey: "456",
				EnvHost:      "https://example.com",
			},
			expected: "vinyldns: some credentials information are missing: VINYLDNS_ACCESS_KEY",
		},
		{
			desc: "missing secret key",
			envVars: map[string]string{
				EnvAccessKey: "123",
				EnvHost:      "https://example.com",
			},
			expected: "vinyldns: some credentials information are missing: VINYLDNS_SECRET_KEY",
		},
		{
			desc: "missing host",
			envVars: map[string]string{
				EnvAccessKey: "123",
				EnvSecretKey: "456",
			},
			expected: "vinyldns: some credentials information are missing: VINYLDNS_HOST",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer envTest.RestoreEnv()
			envTest.ClearEnv()

			envTest.Apply(test.envVars)

			p, err := NewDNSProvider()

			if test.expected == "" {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.config)
				require.NotNil(t, p.client)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc      string
		accessKey string
		secretKey string
		host      string
		expected  string
	}{
		{
			desc:      "success",
			accessKey: "123",
			secretKey: "456",
			host:      "https://example.com",
		},
		{
			desc:     "missing all credentials",
			host:     "https://example.com",
			expected: "vinyldns: credentials are missing",
		},
		{
			desc:      "missing access key",
			secretKey: "456",
			host:      "https://example.com",
			expected:  "vinyldns: credentials are missing",
		},
		{
			desc:      "missing secret key",
			accessKey: "123",
			host:      "https://example.com",
			expected:  "vinyldns: credentials are missing",
		},
		{
			desc:      "missing host",
			accessKey: "123",
			secretKey: "456",
			expected:  "vinyldns: host is missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.AccessKey = test.accessKey
			config.SecretKey = test.secretKey
			config.Host = test.host

			p, err := NewDNSProviderConfig(config)

			if test.expected == "" {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.config)
				require.NotNil(t, p.client)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestDNSProvider_Present_new(t *testing.T) {
	mux, p := setup(t)

	mux.Handle("/", newMockRouter().
		Get("/zones/name/"+targetDomain+".", http.StatusOK, "zoneByName").
		Get("/zones/"+zoneID+"/recordsets", http.StatusOK, "recordSetsListAll-empty").
		Post("/zones/"+zoneID+"/recordsets", http.StatusAccepted, "recordSetUpdate-create").
		Get("/zones/"+zoneID+"/recordsets/"+newRecordSetID+"/changes/"+newCreateChangeID, http.StatusOK, "recordSetChange-create"),
	)

	err := p.Present(targetHostname+"."+targetDomain, "123456d==", "123456d==")
	require.NoError(t, err)
}

func TestDNSProvider_Present_existing(t *testing.T) {
	mux, p := setup(t)

	mux.Handle("/", newMockRouter().
		Get("/zones/name/"+targetDomain+".", http.StatusOK, "zoneByName").
		Get("/zones/"+zoneID+"/recordsets", http.StatusOK, "recordSetsListAll"),
	)

	err := p.Present(targetHostname+"."+targetDomain, "123456d==", "123456d==")
	require.NoError(t, err)
}

func TestDNSProvider_Present_duplicateKey(t *testing.T) {
	mux, p := setup(t)

	mux.Handle("/", newMockRouter().
		Get("/zones/name/"+targetDomain+".", http.StatusOK, "zoneByName").
		Get("/zones/"+zoneID+"/recordsets", http.StatusOK, "recordSetsListAll").
		Put("/zones/"+zoneID+"/recordsets/"+recordID, http.StatusAccepted, "recordSetUpdate-create").
		Get("/zones/"+zoneID+"/recordsets/"+newRecordSetID+"/changes/"+newCreateChangeID, http.StatusOK, "recordSetChange-create"),
	)

	err := p.Present(targetHostname+"."+targetDomain, "abc123!!", "abc123!!")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	mux, p := setup(t)

	mux.Handle("/", newMockRouter().
		Get("/zones/name/"+targetDomain+".", http.StatusOK, "zoneByName").
		Get("/zones/"+zoneID+"/recordsets", http.StatusOK, "recordSetsListAll").
		Delete("/zones/"+zoneID+"/recordsets/"+recordID, http.StatusAccepted, "recordSetDelete").
		Get("/zones/"+zoneID+"/recordsets/"+newRecordSetID+"/changes/"+newCreateChangeID, http.StatusOK, "recordSetChange-delete"),
	)

	err := p.CleanUp(targetHostname+"."+targetDomain, "123456d==", "123456d==")
	require.NoError(t, err)
}

func TestLivePresent(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()
	provider, err := NewDNSProvider()
	require.NoError(t, err)

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

	time.Sleep(2 * time.Second)

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}
