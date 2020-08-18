package stackpath

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/go-acme/lego/v3/platform/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const envDomain = envNamespace + "DOMAIN"

var envTest = tester.NewEnvTest(
	EnvClientID,
	EnvClientSecret,
	EnvStackID).
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
				EnvClientID:     "test@example.com",
				EnvClientSecret: "123",
				EnvStackID:      "ID",
			},
		},
		{
			desc: "missing credentials",
			envVars: map[string]string{
				EnvClientID:     "",
				EnvClientSecret: "",
				EnvStackID:      "",
			},
			expected: "stackpath: some credentials information are missing: STACKPATH_CLIENT_ID,STACKPATH_CLIENT_SECRET,STACKPATH_STACK_ID",
		},
		{
			desc: "missing client id",
			envVars: map[string]string{
				EnvClientID:     "",
				EnvClientSecret: "123",
				EnvStackID:      "ID",
			},
			expected: "stackpath: some credentials information are missing: STACKPATH_CLIENT_ID",
		},
		{
			desc: "missing client secret",
			envVars: map[string]string{
				EnvClientID:     "test@example.com",
				EnvClientSecret: "",
				EnvStackID:      "ID",
			},
			expected: "stackpath: some credentials information are missing: STACKPATH_CLIENT_SECRET",
		},
		{
			desc: "missing stack id",
			envVars: map[string]string{
				EnvClientID:     "test@example.com",
				EnvClientSecret: "123",
				EnvStackID:      "",
			},
			expected: "stackpath: some credentials information are missing: STACKPATH_STACK_ID",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer envTest.RestoreEnv()
			envTest.ClearEnv()

			envTest.Apply(test.envVars)

			p, err := NewDNSProvider(nil)

			if len(test.expected) == 0 {
				require.NoError(t, err)
				assert.NotNil(t, p)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := map[string]struct {
		config      *Config
		expectedErr string
	}{
		"no_config": {
			config:      nil,
			expectedErr: "stackpath: the configuration of the DNS provider is nil",
		},
		"no_client_id": {
			config: &Config{
				ClientSecret: "secret",
				StackID:      "stackID",
			},
			expectedErr: "stackpath: credentials missing",
		},
		"no_client_secret": {
			config: &Config{
				ClientID: "clientID",
				StackID:  "stackID",
			},
			expectedErr: "stackpath: credentials missing",
		},
		"no_stack_id": {
			config: &Config{
				ClientID:     "clientID",
				ClientSecret: "secret",
			},
			expectedErr: "stackpath: stack id missing",
		},
	}

	for desc, test := range testCases {
		test := test
		t.Run(desc, func(t *testing.T) {
			t.Parallel()

			p, err := NewDNSProviderConfig(test.config)
			require.EqualError(t, err, test.expectedErr)
			assert.Nil(t, p)
		})
	}
}

func setupMockAPITest() (*DNSProvider, *http.ServeMux, func()) {
	apiHandler := http.NewServeMux()
	server := httptest.NewServer(apiHandler)

	config := NewDefaultConfig(nil)
	config.ClientID = "CLIENT_ID"
	config.ClientSecret = "CLIENT_SECRET"
	config.StackID = "STACK_ID"

	provider, err := NewDNSProviderConfig(config)
	if err != nil {
		panic(err)
	}

	provider.client = http.DefaultClient
	provider.BaseURL, _ = url.Parse(server.URL + "/")

	return provider, apiHandler, server.Close
}

func TestDNSProvider_getZoneRecords(t *testing.T) {
	provider, mux, tearDown := setupMockAPITest()
	defer tearDown()

	mux.HandleFunc("/STACK_ID/zones/A/records", func(w http.ResponseWriter, _ *http.Request) {
		content := `
			{
				"records": [
					{"id":"1","name":"foo1","type":"TXT","ttl":120,"data":"txtTXTtxt"},
					{"id":"2","name":"foo2","type":"TXT","ttl":121,"data":"TXTtxtTXT"}
				]
			}`

		_, err := w.Write([]byte(content))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	records, err := provider.getZoneRecords("foo1", &Zone{ID: "A", Domain: "test"})
	require.NoError(t, err)

	expected := []Record{
		{ID: "1", Name: "foo1", Type: "TXT", TTL: 120, Data: "txtTXTtxt"},
		{ID: "2", Name: "foo2", Type: "TXT", TTL: 121, Data: "TXTtxtTXT"},
	}

	assert.Equal(t, expected, records)
}

func TestDNSProvider_getZoneRecords_apiError(t *testing.T) {
	provider, mux, tearDown := setupMockAPITest()
	defer tearDown()

	mux.HandleFunc("/STACK_ID/zones/A/records", func(w http.ResponseWriter, _ *http.Request) {
		content := `
{
	"code": 401,
	"error": "an unauthorized request is attempted."
}`

		w.WriteHeader(http.StatusUnauthorized)
		_, err := w.Write([]byte(content))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	_, err := provider.getZoneRecords("foo1", &Zone{ID: "A", Domain: "test"})

	expected := &ErrorResponse{Code: 401, Message: "an unauthorized request is attempted."}
	assert.Equal(t, expected, err)
}

func TestDNSProvider_getZones(t *testing.T) {
	provider, mux, tearDown := setupMockAPITest()
	defer tearDown()

	mux.HandleFunc("/STACK_ID/zones", func(w http.ResponseWriter, _ *http.Request) {
		content := `
{
  "pageInfo": {
    "totalCount": "5",
    "hasPreviousPage": false,
    "hasNextPage": false,
    "startCursor": "1",
    "endCursor": "1"
  },
  "zones": [
    {
      "stackId": "my_stack",
      "accountId": "my_account",
      "id": "A",
      "domain": "foo.com",
      "version": "1",
      "labels": {
        "property1": "val1",
        "property2": "val2"
      },
      "created": "2018-10-07T02:31:49Z",
      "updated": "2018-10-07T02:31:49Z",
      "nameservers": [
        "1.1.1.1"
      ],
      "verified": "2018-10-07T02:31:49Z",
      "status": "ACTIVE",
      "disabled": false
    }
  ]
}`

		_, err := w.Write([]byte(content))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	zone, err := provider.getZones("sub.foo.com")
	require.NoError(t, err)

	expected := &Zone{ID: "A", Domain: "foo.com"}

	assert.Equal(t, expected, zone)
}

func TestLivePresent(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()
	provider, err := NewDNSProvider(nil)
	require.NoError(t, err)

	err = provider.Present(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}

func TestLiveCleanUp(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()
	provider, err := NewDNSProvider(nil)
	require.NoError(t, err)

	time.Sleep(1 * time.Second)

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}
