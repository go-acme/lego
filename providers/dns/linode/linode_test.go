package linode

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/linode/linodego"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var envTest = tester.NewEnvTest(EnvToken)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvToken: "123",
			},
		},
		{
			desc: "missing api key",
			envVars: map[string]string{
				EnvToken: "",
			},
			expected: "linode: some credentials information are missing: LINODE_TOKEN",
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
		desc     string
		apiKey   string
		expected string
	}{
		{
			desc:   "success",
			apiKey: "123",
		},
		{
			desc:     "missing credentials",
			expected: "linode: Linode Access Token missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.Token = test.apiKey

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

func TestDNSProvider_Present(t *testing.T) {
	defer envTest.RestoreEnv()

	os.Setenv(EnvToken, "testing")

	domain := "example.com"
	keyAuth := "dGVzdGluZw=="

	testCases := []struct {
		desc          string
		builder       *servermock.Builder[*DNSProvider]
		expectedError string
	}{
		{
			desc: "Success",
			builder: mockBuilder().
				Route("GET /v4/domains",
					servermock.JSONEncode(linodego.DomainsPagedResponse{
						PageOptions: &linodego.PageOptions{
							Pages:   1,
							Results: 1,
							Page:    1,
						},
						Data: []linodego.Domain{{
							Domain: domain,
							ID:     1234,
						}},
					})).
				Route("POST /v4/domains/1234/records", servermock.JSONEncode(linodego.DomainRecord{
					ID: 1234,
				})),
		},
		{
			desc: "NoDomain",
			builder: mockBuilder().
				Route("GET /v4/domains",
					servermock.JSONEncode(linodego.APIError{
						Errors: []linodego.APIErrorReason{{
							Reason: "Not found",
						}},
					}).
						WithStatusCode(http.StatusNotFound)),
			expectedError: "[404] Not found",
		},
		{
			desc: "CreateFailed",
			builder: mockBuilder().
				Route("GET /v4/domains",
					servermock.JSONEncode(&linodego.DomainsPagedResponse{
						PageOptions: &linodego.PageOptions{
							Pages:   1,
							Results: 1,
							Page:    1,
						},
						Data: []linodego.Domain{{
							Domain: "example.com",
							ID:     1234,
						}},
					})).
				Route("POST /v4/domains/1234/records",
					servermock.JSONEncode(linodego.APIError{
						Errors: []linodego.APIErrorReason{{
							Reason: "Failed to create domain resource",
							Field:  "somefield",
						}},
					}).
						WithStatusCode(http.StatusBadRequest)),
			expectedError: "[400] [somefield] Failed to create domain resource",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			provider := test.builder.Build(t)

			err := provider.Present(domain, "", keyAuth)
			if test.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, test.expectedError)
			}
		})
	}
}

func TestDNSProvider_CleanUp(t *testing.T) {
	defer envTest.RestoreEnv()

	os.Setenv(EnvToken, "testing")

	domain := "example.com"
	keyAuth := "dGVzdGluZw=="

	testCases := []struct {
		desc          string
		builder       *servermock.Builder[*DNSProvider]
		expectedError string
	}{
		{
			desc: "Success",
			builder: mockBuilder().
				Route("GET /v4/domains",
					servermock.JSONEncode(&linodego.DomainsPagedResponse{
						PageOptions: &linodego.PageOptions{
							Pages:   1,
							Results: 1,
							Page:    1,
						},
						Data: []linodego.Domain{{
							Domain: "foobar.com",
							ID:     1234,
						}},
					})).
				Route("GET /v4/domains/1234/records",
					servermock.JSONEncode(&linodego.DomainRecordsPagedResponse{
						PageOptions: &linodego.PageOptions{
							Pages:   1,
							Results: 1,
							Page:    1,
						},
						Data: []linodego.DomainRecord{{
							ID:     1234,
							Name:   "_acme-challenge",
							Target: "ElbOJKOkFWiZLQeoxf-wb3IpOsQCdvoM0y_wn0TEkxM",
							Type:   "TXT",
						}},
					})).
				Route("DELETE /v4/domains/1234/records/1234",
					servermock.RawStringResponse("{}").WithHeader("Content-Type", "application/json")),
		},
		{
			desc: "NoDomain",
			builder: mockBuilder().
				Route("GET /v4/domains",
					servermock.JSONEncode(linodego.APIError{
						Errors: []linodego.APIErrorReason{{
							Reason: "Not found",
						}},
					}).
						WithStatusCode(http.StatusNotFound)).
				Route("GET /v4/domains/1234/records",
					servermock.JSONEncode(linodego.APIError{
						Errors: []linodego.APIErrorReason{{
							Reason: "Not found",
						}},
					},
					).
						WithStatusCode(http.StatusNotFound)),
			expectedError: "[404] Not found",
		},
		{
			desc: "DeleteFailed",
			builder: mockBuilder().
				Route("GET /v4/domains",
					servermock.JSONEncode(linodego.DomainsPagedResponse{
						PageOptions: &linodego.PageOptions{
							Pages:   1,
							Results: 1,
							Page:    1,
						},
						Data: []linodego.Domain{{
							ID:     1234,
							Domain: "example.com",
						}},
					})).
				Route("GET /v4/domains/1234/records",
					servermock.JSONEncode(linodego.DomainRecordsPagedResponse{
						PageOptions: &linodego.PageOptions{
							Pages:   1,
							Results: 1,
							Page:    1,
						},
						Data: []linodego.DomainRecord{{
							ID:     1234,
							Name:   "_acme-challenge",
							Target: "ElbOJKOkFWiZLQeoxf-wb3IpOsQCdvoM0y_wn0TEkxM",
							Type:   "TXT",
						}},
					})).
				Route("DELETE /v4/domains/1234/records/1234",
					servermock.JSONEncode(linodego.APIError{
						Errors: []linodego.APIErrorReason{{
							Reason: "Failed to delete domain resource",
						}},
					}).
						WithStatusCode(http.StatusBadRequest)),
			expectedError: "[400] Failed to delete domain resource",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			provider := test.builder.Build(t)

			err := provider.CleanUp(domain, "", keyAuth)
			if test.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, test.expectedError)
			}
		})
	}
}

func TestLivePresent(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("Skipping live test")
	}
	// TODO implement this test
}

func TestLiveCleanUp(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("Skipping live test")
	}
	// TODO implement this test
}

func mockBuilder() *servermock.Builder[*DNSProvider] {
	return servermock.NewBuilder(func(server *httptest.Server) (*DNSProvider, error) {
		p, err := NewDNSProvider()
		if err != nil {
			return nil, err
		}

		p.client.SetBaseURL(server.URL)

		return p, nil
	})
}
