package acmedns

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nrdcg/goacmedns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	egDomain  = "example.com"
	egFQDN    = "_acme-challenge." + egDomain + "."
	egKeyAuth = "⚷"
)

func TestPresent(t *testing.T) {
	// validAccountStorage is a mockStorage configured to return the egTestAccount.
	validAccountStorage := newMockStorage().WithAccount(egDomain, egTestAccount)

	// validUpdateClient is a mockClient configured with the egTestAccount that will track TXT updates in a map.
	validUpdateClient := newMockClient()

	testCases := []struct {
		Name          string
		Client        acmeDNSClient
		Storage       goacmedns.Storage
		ExpectedError error
	}{
		{
			Name:          "present when client storage returns unexpected error",
			Client:        newMockClient().WithRegisterAccount(egTestAccount),
			Storage:       newMockStorage().WithFetchError(errorStorageErr),
			ExpectedError: errorStorageErr,
		},
		{
			Name:   "present when client storage returns ErrDomainNotFound",
			Client: newMockClient().WithRegisterAccount(egTestAccount),
			ExpectedError: ErrCNAMERequired{
				Domain: egDomain,
				FQDN:   egFQDN,
				Target: egTestAccount.FullDomain,
			},
		},
		{
			Name:          "present when client UpdateTXTRecord returns unexpected error",
			Client:        newMockClient().WithUpdateTXTRecordError(errorClientErr),
			Storage:       validAccountStorage,
			ExpectedError: errorClientErr,
		},
		{
			Name:    "present when everything works",
			Storage: validAccountStorage,
			Client:  validUpdateClient,
		},
	}

	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {
			p := &DNSProvider{
				config:  NewDefaultConfig(),
				client:  test.Client,
				storage: newMockStorage(),
			}

			if test.Storage != nil {
				p.storage = test.Storage
			}

			err := p.Present(egDomain, "foo", egKeyAuth)
			if test.ExpectedError != nil {
				assert.Equal(t, test.ExpectedError, err)
			} else {
				require.NoError(t, err)
			}
		})
	}

	// Check that the success test case set a record.
	assert.Len(t, validUpdateClient.records, 1)

	// Check that the success test case set the right record for the right account.
	assert.Len(t, validUpdateClient.records[egTestAccount], 43)
}

func TestRegister(t *testing.T) {
	testCases := []struct {
		Name          string
		Client        acmeDNSClient
		Storage       goacmedns.Storage
		ExpectedError error
	}{
		{
			Name:          "register when acme-dns client returns an error",
			Client:        newMockClient().WithRegisterAccountError(errorClientErr),
			ExpectedError: errorClientErr,
		},
		{
			Name:          "register when acme-dns storage put returns an error",
			Client:        newMockClient().WithRegisterAccount(egTestAccount),
			Storage:       newMockStorage().WithPutError(errorStorageErr),
			ExpectedError: errorStorageErr,
		},
		{
			Name:          "register when acme-dns storage save returns an error",
			Client:        newMockClient().WithRegisterAccount(egTestAccount),
			Storage:       newMockStorage().WithSaveError(errorStorageErr),
			ExpectedError: errorStorageErr,
		},
		{
			Name:   "register when everything works",
			Client: newMockClient().WithRegisterAccount(egTestAccount),
			ExpectedError: ErrCNAMERequired{
				Domain: egDomain,
				FQDN:   egFQDN,
				Target: egTestAccount.FullDomain,
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {
			p := &DNSProvider{
				config:  NewDefaultConfig(),
				client:  test.Client,
				storage: newMockStorage(),
			}

			if test.Storage != nil {
				p.storage = test.Storage
			}

			acc, err := p.register(t.Context(), egDomain, egFQDN)
			if test.ExpectedError != nil {
				assert.Equal(t, test.ExpectedError, err)
			} else {
				assert.Equal(t, goacmedns.Account{}, acc)
				require.NoError(t, err)
			}
		})
	}
}

func TestPresent_httpStorage(t *testing.T) {
	testCases := []struct {
		desc          string
		StatusCode    int
		ExpectedError error
	}{
		{
			desc:       "the CNAME is not handled by the storage",
			StatusCode: http.StatusOK,
			ExpectedError: ErrCNAMERequired{
				Domain: egDomain,
				FQDN:   egFQDN,
				Target: egTestAccount.FullDomain,
			},
		},
		{
			desc:       "the CNAME is handled by the storage",
			StatusCode: http.StatusCreated,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			mux := http.NewServeMux()
			server := httptest.NewServer(mux)

			config := NewDefaultConfig()
			config.StorageBaseURL = server.URL

			p, err := NewDNSProviderConfig(config)
			require.NoError(t, err)

			client := newMockClient().WithRegisterAccount(egTestAccount)
			p.client = client

			// Fetch
			mux.HandleFunc("GET /example.com", func(rw http.ResponseWriter, reg *http.Request) {
				rw.WriteHeader(http.StatusNotFound)
			})

			// Put
			mux.HandleFunc("POST /example.com", func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(test.StatusCode)
			})

			err = p.Present(egDomain, "foo", egKeyAuth)
			if test.ExpectedError != nil {
				assert.Equal(t, test.ExpectedError, err)
				assert.True(t, client.registerAccountCalled)
				assert.False(t, client.updateTXTRecordCalled)
			} else {
				require.NoError(t, err)
				assert.True(t, client.registerAccountCalled)
				assert.True(t, client.updateTXTRecordCalled)
			}
		})
	}
}

func TestRegister_httpStorage(t *testing.T) {
	testCases := []struct {
		Name          string
		StatusCode    int
		ExpectedError error
	}{
		{
			Name:       "status code 200",
			StatusCode: http.StatusOK,
			ExpectedError: ErrCNAMERequired{
				Domain: egDomain,
				FQDN:   egFQDN,
				Target: egTestAccount.FullDomain,
			},
		},
		{
			Name:       "status code 201",
			StatusCode: http.StatusCreated,
		},
	}

	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {
			mux := http.NewServeMux()
			server := httptest.NewServer(mux)

			config := NewDefaultConfig()
			config.StorageBaseURL = server.URL

			p, err := NewDNSProviderConfig(config)
			require.NoError(t, err)

			p.client = newMockClient().WithRegisterAccount(egTestAccount)

			// Put
			mux.HandleFunc("POST /example.com", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(test.StatusCode)
			})

			acc, err := p.register(t.Context(), egDomain, egFQDN)
			if test.ExpectedError != nil {
				assert.Equal(t, test.ExpectedError, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, egTestAccount, acc)
			}
		})
	}
}
