package cloudru

import (
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/go-acme/lego/v5/internal/tester"
	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/go-acme/lego/v5/providers/dns/cloudru/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(
	EnvServiceInstanceID,
	EnvKeyID,
	EnvSecret).
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
				EnvServiceInstanceID: "123",
				EnvKeyID:             "user",
				EnvSecret:            "secret",
			},
		},
		{
			desc:     "missing credentials",
			envVars:  map[string]string{},
			expected: "cloudru: some credentials information are missing: CLOUDRU_SERVICE_INSTANCE_ID,CLOUDRU_KEY_ID,CLOUDRU_SECRET",
		},
		{
			desc: "missing service instance ID",
			envVars: map[string]string{
				EnvServiceInstanceID: "",
				EnvKeyID:             "user",
				EnvSecret:            "secret",
			},
			expected: "cloudru: some credentials information are missing: CLOUDRU_SERVICE_INSTANCE_ID",
		},
		{
			desc: "missing key ID",
			envVars: map[string]string{
				EnvServiceInstanceID: "123",
				EnvKeyID:             "",
				EnvSecret:            "secret",
			},
			expected: "cloudru: some credentials information are missing: CLOUDRU_KEY_ID",
		},
		{
			desc: "missing secret",
			envVars: map[string]string{
				EnvServiceInstanceID: "123",
				EnvKeyID:             "user",
				EnvSecret:            "",
			},
			expected: "cloudru: some credentials information are missing: CLOUDRU_SECRET",
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
		desc              string
		serviceInstanceID string
		keyID             string
		secret            string
		expected          string
	}{
		{
			desc:              "success",
			serviceInstanceID: "123",
			keyID:             "user",
			secret:            "secret",
		},
		{
			desc:     "missing credentials",
			expected: "cloudru: some credentials information are missing",
		},
		{
			desc:              "missing service instance ID",
			serviceInstanceID: "",
			keyID:             "user",
			secret:            "secret",
			expected:          "cloudru: some credentials information are missing",
		},
		{
			desc:              "missing key ID",
			serviceInstanceID: "123",
			keyID:             "",
			secret:            "secret",
			expected:          "cloudru: some credentials information are missing",
		},
		{
			desc:              "missing secret",
			serviceInstanceID: "123",
			keyID:             "user",
			secret:            "",
			expected:          "cloudru: some credentials information are missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.ServiceInstanceID = test.serviceInstanceID
			config.KeyID = test.keyID
			config.Secret = test.secret

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

func TestLivePresent(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()

	provider, err := NewDNSProvider()
	require.NoError(t, err)

	err = provider.Present(t.Context(), envTest.GetDomain(), "", "123d==")
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

	err = provider.CleanUp(t.Context(), envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}

func oauthServer(t *testing.T) string {
	t.Helper()

	return servermock.NewBuilder[string](
		func(server *httptest.Server) (string, error) {
			return server.URL, nil
		}).
		Route("POST /",
			servermock.ResponseFromInternal("auth.json"),
			servermock.CheckHeader().
				WithContentTypeFromURLEncoded(),
			servermock.CheckForm().Strict().
				With("client_id", "user").
				With("client_secret", "secret").
				With("grant_type", "access_key"),
		).
		Build(t)
}

func mockBuilder(tokenURL string) *servermock.Builder[*DNSProvider] {
	return servermock.NewBuilder(
		func(server *httptest.Server) (*DNSProvider, error) {
			config := NewDefaultConfig()
			config.KeyID = "user"
			config.Secret = "secret"
			config.ServiceInstanceID = "xxx"

			authConfig := internal.CreateOAuthConfig(tokenURL, config.KeyID, config.Secret)

			p, err := newDNSProviderConfig(config, authConfig)
			if err != nil {
				return nil, err
			}

			p.client.BaseURL, _ = url.Parse(server.URL)

			return p, nil
		},
		servermock.CheckHeader().
			WithJSONHeaders().
			WithAuthorization("Bearer "+
				"eyJhbGciOiJSUzI1NiIsImtpZCI6ImEyMjM3ZDhhLWQ0ZDQtNDA5Yi04ZTMxLWM3NGJhYTZhM2NjYiIsInR5cCI6IkpXVCJ9."+
				"eyJhdWQiOlsiaWFtIl0sImF1dGhfdGltZSI6MTY5MDA1Mzg1MiwiYXpwIjoiOWE0M2I0OTM1ZDRhMDc5NmRkYjE0Mjk0NjUxYjk2NzciLCJlbWFpbCI6ImxlZ29AOTlmN2I5NzItZmZlYS00OTkyLTgyN2EtY2M4MDYzOTg1MmNhLmlhbS5zYmVyY2xvdWQucnUiLCJleHAiOjE2OTAwNTc0NTIsImlhdCI6MTY5MDA1Mzg1MiwiaXNzIjoiaHR0cHM6Ly9hdXRoLmlhbS5zYmVyY2xvdWQucnUvYXV0aC9zeXN0ZW0iLCJqdGkiOiJlYzk0ZWJhNC03NzU2LTRjNjQtYmNmMC0zMzYxODIwNWM5ODkiLCJuYmYiOjE2OTAwNTM4NTIsIm5vbmNlIjoiIiwicmVhbG1fYWNjZXNzIjpudWxsLCJyZXNvdXJjZV9hY2Nlc3MiOnt9LCJzY29wZSI6Im9wZW5pZCBwcm9maWxlIGVtYWlsIHJvbGVzIiwic3ViIjoiZmVhMzRmYTUtZmE5ZS00OTdkLTk5MDAtNjdhYWJjOWNkZWJjIiwic3ViX2lkIjoiZmVhMzRmYTUtZmE5ZS00OTdkLTk5MDAtNjdhYWJjOWNkZWJjIiwic3ViX3R5cGUiOiJzZXJ2aWNlX2FjY291bnQiLCJ0eXAiOiJCZWFyZXIifQ."+
				"hhPr-Xr_NbyRwrqGoqeepthWfpfmD47RjzHUwo2lVPkeMiL8AMWzDPRxs-8gns9eTSHZCoAH0RjyrBnTaOrztInM72h8_rIIFr0MMPIIqrUkp2id_alya9eoiSWg_69PzNZ2CKWJDylL8o4Vi9_cSBYp-6H1xNcOAvO4a9xkNCoGGiogjHWNFq64qnS_P6fYY-pl9leuprCeq1GAKPODevHwzmc4gkEZIj_15SUh_ofJRJICgyLmkELQ8a0wDGYmZcdNKiGQDpd7rHaGrOvO1k8IJHfgs5aCMyuHXybTg6AMlodpYs8MBdk6K_VFY-cxSRB8ocq_Q7Hgt9qaRADg2Q",
			),
	)
}

func TestDNSProvider_Present(t *testing.T) {
	provider := mockBuilder(oauthServer(t)).
		Route("GET /zones",
			servermock.ResponseFromInternal("zones.json"),
			servermock.CheckQueryParameter().Strict().
				With("parentId", "xxx"),
		).
		Route("POST /zones/59556fcd-95ff-451f-b49b-9732f21f944a/records",
			servermock.ResponseFromInternal("record.json"),
			servermock.CheckRequestJSONBodyFromInternal("record-request.json"),
		).
		Build(t)

	err := provider.Present(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)

	assert.Equal(t, map[string]*internal.Record{"abc": {
		ZoneID:  "59556fcd-95ff-451f-b49b-9732f21f944a",
		Name:    "_acme-challenge.example.com.",
		Type:    "TXT",
		Values:  []string{"ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY"},
		TTL:     "120",
		Enables: true,
	}}, provider.records)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockBuilder(oauthServer(t)).
		Route("DELETE /zones/59556fcd-95ff-451f-b49b-9732f21f944a/records/_acme-challenge.example.com./TXT",
			servermock.ResponseFromInternal("record.json"),
		).
		Build(t)

	provider.records["abc"] = &internal.Record{
		ZoneID:  "59556fcd-95ff-451f-b49b-9732f21f944a",
		Name:    "_acme-challenge.example.com.",
		Type:    "TXT",
		Values:  []string{"ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY"},
		TTL:     "120",
		Enables: true,
	}

	err := provider.CleanUp(t.Context(), "example.com", "abc", "123d==")
	require.NoError(t, err)

	assert.Empty(t, provider.records)
}
