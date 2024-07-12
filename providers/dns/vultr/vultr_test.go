package vultr

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vultr/govultr/v2"
)

const envDomain = envNamespace + "TEST_DOMAIN"

var envTest = tester.NewEnvTest(EnvAPIKey).
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
				EnvAPIKey: "123",
			},
		},
		{
			desc: "missing api key",
			envVars: map[string]string{
				EnvAPIKey: "",
			},
			expected: "vultr: some credentials information are missing: VULTR_API_KEY",
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
			expected: "vultr: credentials missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.APIKey = test.apiKey

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

func TestDNSProvider_getHostedZone(t *testing.T) {
	testCases := []struct {
		desc              string
		domain            string
		expected          string
		expectedPageCount int
	}{
		{
			desc:              "exact match, in latest page",
			domain:            "test.my.example.com",
			expected:          "test.my.example.com",
			expectedPageCount: 5,
		},
		{
			desc:              "exact match, in the middle",
			domain:            "my.example.com",
			expected:          "my.example.com",
			expectedPageCount: 3,
		},
		{
			desc:              "exact match, first page",
			domain:            "example.com",
			expected:          "example.com",
			expectedPageCount: 1,
		},
		{
			desc:              "match on apex",
			domain:            "test.example.org",
			expected:          "example.org",
			expectedPageCount: 5,
		},
		{
			desc:              "match on parent",
			domain:            "test.my.example.net",
			expected:          "my.example.net",
			expectedPageCount: 5,
		},
	}

	domains := []govultr.Domain{{Domain: "example.com"}, {Domain: "example.org"}, {Domain: "example.net"}}

	for i := range 50 {
		domains = append(domains, govultr.Domain{Domain: fmt.Sprintf("my%02d.example.com", i)})
	}

	domains = append(domains, govultr.Domain{Domain: "my.example.com"}, govultr.Domain{Domain: "my.example.net"})

	for i := 50; i < 100; i++ {
		domains = append(domains, govultr.Domain{Domain: fmt.Sprintf("my%02d.example.com", i)})
	}

	domains = append(domains, govultr.Domain{Domain: "test.my.example.com"})

	type domainsBase struct {
		Domains []govultr.Domain `json:"domains"`
		Meta    *govultr.Meta    `json:"meta"`
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			mux := http.NewServeMux()
			server := httptest.NewServer(mux)
			t.Cleanup(server.Close)

			client := govultr.NewClient(nil)
			err := client.SetBaseURL(server.URL)
			require.NoError(t, err)

			p := &DNSProvider{client: client}

			var pageCount int

			mux.HandleFunc("/v2/domains", func(rw http.ResponseWriter, req *http.Request) {
				pageCount++

				query := req.URL.Query()
				cursor, _ := strconv.Atoi(query.Get("cursor"))
				perPage, _ := strconv.Atoi(query.Get("per_page"))

				var next string
				if len(domains)/perPage > cursor {
					next = strconv.Itoa(cursor + 1)
				}

				start := cursor * perPage
				if len(domains) < start {
					start = cursor * len(domains)
				}

				end := (cursor + 1) * perPage
				if len(domains) < end {
					end = len(domains)
				}

				db := domainsBase{
					Domains: domains[start:end],
					Meta: &govultr.Meta{
						Total: len(domains),
						Links: &govultr.Links{Next: next},
					},
				}

				err = json.NewEncoder(rw).Encode(db)
				if err != nil {
					http.Error(rw, err.Error(), http.StatusInternalServerError)
					return
				}
			})

			zone, err := p.getHostedZone(context.Background(), test.domain)
			require.NoError(t, err)

			assert.Equal(t, test.expected, zone)
			assert.Equal(t, test.expectedPageCount, pageCount)
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
