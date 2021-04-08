package vinyldns

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vinyldns/go-vinyldns/vinyldns"
)

const (
	envDomain         = envNamespace + "DOMAIN"
	targetHostname    = "host"
	targetDomain      = "vinyldns.invalid"
	zoneID            = "00000000-0000-0000-0000-000000000000"
	recordsetID       = "10000000-0000-0000-0000-000000000000"
	newRecordSetID    = "11000000-0000-0000-0000-000000000000"
	newCreateChangeID = "20000000-0000-0000-0000-000000000000"
	deleteChangeID    = "21000000-0000-0000-0000-000000000000"
	recordName        = "_acme-challenge" + targetHostname
	recordID          = "30000000-0000-0000-0000-000000000000"
)

func getDefaultMockMapping() MockResponseMap {
	defaultMockMapping := MockResponseMap{
		"GET": {
			"/zones/name/" + targetDomain + ".":                                                    {StatusCode: 200, Body: GetZoneResponse},
			"/zones/" + zoneID + "/recordsets/" + newRecordSetID + "/changes/" + newCreateChangeID: {StatusCode: 200, Body: GetCreateRRSetStatusResponse},
			"/zones/" + zoneID + "/recordsets/" + recordsetID + "/changes/" + deleteChangeID:       {StatusCode: 200, Body: GetDeleteRRSetStatusResponse},
			"/zones/" + zoneID + "/recordsets?recordNameFilter=" + recordName:                      {StatusCode: 200, Body: FindRRSetResponse},
			"/zones/" + zoneID + "/recordsets":                                                     {StatusCode: 200, Body: FindRRSetResponse},
		},
		"POST": {
			"/zones/" + zoneID + "/recordsets": {StatusCode: 202, Body: CreateRRSetResponse},
		},
		"PUT": {
			"/zones/" + zoneID + "/recordsets/" + recordID: {StatusCode: 202, Body: CreateRRSetResponse},
		},
		"DELETE": {
			"/zones/" + zoneID + "/recordsets/" + recordID: {StatusCode: 202, Body: DeleteRRSetResponse},
		},
	}

	return defaultMockMapping
}

var envTest = tester.NewEnvTest(
	envDomain,
	EnvAccessKey,
	EnvSecretKey,
	EnvHost,
	EnvTTL,
	EnvPropagationTimeout,
	EnvPollingInterval).
	WithDomain(envDomain).
	WithLiveTestRequirements(EnvAccessKey, EnvSecretKey, EnvHost, envDomain)

func makeTestProvider(serverURL string) *DNSProvider {
	config := vinyldns.ClientConfiguration{
		AccessKey: "foo",
		SecretKey: "bar",
		Host:      serverURL,
		UserAgent: "go-acme/lego",
	}

	return &DNSProvider{
		client: vinyldns.NewClient(config),
		config: NewDefaultConfig(),
	}
}

func makeTestProviderFull(accessKey, secretKey, host, userAgent string, ttl int, propagationTimeout, pollingInterval time.Duration) *DNSProvider {
	vinyldnsConfig := vinyldns.ClientConfiguration{
		AccessKey: accessKey,
		SecretKey: secretKey,
		Host:      host,
		UserAgent: userAgent,
	}
	providerConfig := Config{
		AccessKey:          accessKey,
		SecretKey:          secretKey,
		Host:               host,
		PropagationTimeout: propagationTimeout,
		PollingInterval:    pollingInterval,
		TTL:                ttl,
	}

	return &DNSProvider{
		client: vinyldns.NewClient(vinyldnsConfig),
		config: &providerConfig,
	}
}

func Test_loadConfig_fromEnv(t *testing.T) {
	defer envTest.RestoreEnv()
	envTest.ClearEnv()

	_ = os.Setenv(EnvAccessKey, "123")
	_ = os.Setenv(EnvSecretKey, "456")
	_ = os.Setenv(EnvHost, "http://host.company.invalid")
	_ = os.Setenv(EnvTTL, "60")
	_ = os.Setenv(EnvPropagationTimeout, fmt.Sprintf("%.0f", (3*time.Minute).Seconds()))
	_ = os.Setenv(EnvPollingInterval, fmt.Sprintf("%.0f", (5*time.Second).Seconds()))

	provider := makeTestProviderFull("123", "456", "http://host.company.invalid", "go-acme/lego", 60, 3*time.Minute, 5*time.Second)

	expected, _ := NewDNSProvider()
	assert.Equal(t, expected.config.AccessKey, provider.config.AccessKey)
	assert.Equal(t, expected.config.SecretKey, provider.config.SecretKey)
	assert.Equal(t, expected.config.Host, provider.config.Host)
	assert.Equal(t, expected.config.PollingInterval, provider.config.PollingInterval)
	assert.Equal(t, expected.config.PropagationTimeout, provider.config.PropagationTimeout)
	assert.Equal(t, expected.config.TTL, provider.config.TTL)
	assert.Equal(t, expected.client.AccessKey, provider.client.AccessKey)
	assert.Equal(t, expected.client.SecretKey, provider.client.SecretKey)
	assert.Equal(t, expected.client.Host, provider.client.Host)
	assert.Equal(t, expected.client.UserAgent, provider.client.UserAgent)
}

func TestNewDefaultConfig(t *testing.T) {
	defer envTest.RestoreEnv()

	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected *Config
	}{
		{
			desc: "default configuration",
			expected: &Config{
				TTL:                30,
				PropagationTimeout: 2 * time.Minute,
				PollingInterval:    4 * time.Second,
			},
		},
		{
			desc: "non-default configuration",
			envVars: map[string]string{
				EnvTTL:                "99",
				EnvPropagationTimeout: "60",
				EnvPollingInterval:    "60",
			},
			expected: &Config{
				TTL:                99,
				PropagationTimeout: 60 * time.Second,
				PollingInterval:    60 * time.Second,
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			envTest.ClearEnv()
			for key, value := range test.envVars {
				os.Setenv(key, value)
			}

			config := NewDefaultConfig()

			assert.Equal(t, test.expected, config)
		})
	}
}

func TestDNSProvider_Present_ExistingACME(t *testing.T) {
	mockResponses := getDefaultMockMapping()

	serverURL := newMockServer(t, mockResponses)

	defer envTest.RestoreEnv()
	envTest.ClearEnv()
	provider := makeTestProvider(serverURL)

	err := provider.Present(targetHostname+"."+targetDomain, "123456d==", "123456d==")
	require.NoError(t, err, "Expected Present to return no error")
}

func TestDNSProvider_Present_DuplicateKeyACME(t *testing.T) {
	mockResponses := getDefaultMockMapping()

	serverURL := newMockServer(t, mockResponses)

	defer envTest.RestoreEnv()
	envTest.ClearEnv()
	provider := makeTestProvider(serverURL)

	err := provider.Present(targetHostname+"."+targetDomain, "abc123!!", "abc123!!")
	require.NoError(t, err, "Expected Present to return no error")
}

func TestDNSProvider_Present_NewACME(t *testing.T) {
	mockResponses := getDefaultMockMapping()
	missingResponse := MockResponse{StatusCode: 200, Body: FindEmptyRRSetResponse}
	mockResponses["GET"]["/zones/"+zoneID+"/recordsets?recordNameFilter="+recordName] = missingResponse
	mockResponses["GET"]["/zones/"+zoneID+"/recordsets"] = missingResponse

	serverURL := newMockServer(t, mockResponses)

	defer envTest.RestoreEnv()
	envTest.ClearEnv()
	provider := makeTestProvider(serverURL)

	err := provider.Present(targetHostname+"."+targetDomain, "123456d==", "123456d==")
	require.NoError(t, err, "Expected Present to return no error")
}

func TestDNSProvider_CleanUp(t *testing.T) {
	mockResponses := getDefaultMockMapping()

	serverURL := newMockServer(t, mockResponses)

	defer envTest.RestoreEnv()
	envTest.ClearEnv()
	provider := makeTestProvider(serverURL)

	err := provider.CleanUp(targetHostname+"."+targetDomain, "123456d==", "123456d==")
	require.NoError(t, err, "Expected Present to return no error")
}
