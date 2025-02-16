package acmedns

import (
	"context"
	"testing"

	"github.com/nrdcg/goacmedns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	// Fixed test data for unit tests.
	egDomain  = "example.com"
	egFQDN    = "_acme-challenge." + egDomain + "."
	egKeyAuth = "⚷"
)

// TestPresent tests that the ACME-DNS Present function for updating a DNS-01
// challenge response TXT record works as expected.
func TestPresent(t *testing.T) {
	// validAccountStorage is a mockStorage configured to return the egTestAccount.
	validAccountStorage := mockStorage{
		map[string]goacmedns.Account{
			egDomain: egTestAccount,
		},
	}
	// validUpdateClient is a mockClient configured with the egTestAccount that will
	// track TXT updates in a map.
	validUpdateClient := mockUpdateClient{
		mockClient{egTestAccount},
		make(map[goacmedns.Account]string),
	}

	testCases := []struct {
		Name          string
		Client        acmeDNSClient
		Storage       goacmedns.Storage
		ExpectedError error
	}{
		{
			Name:          "present when client storage returns unexpected error",
			Client:        mockClient{egTestAccount},
			Storage:       errorFetchStorage{},
			ExpectedError: errorStorageErr,
		},
		{
			Name:   "present when client storage returns ErrDomainNotFound",
			Client: mockClient{egTestAccount},
			ExpectedError: ErrCNAMERequired{
				Domain: egDomain,
				FQDN:   egFQDN,
				Target: egTestAccount.FullDomain,
			},
		},
		{
			Name:          "present when client UpdateTXTRecord returns unexpected error",
			Client:        errorUpdateClient{},
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
				storage: mockStorage{make(map[string]goacmedns.Account)},
			}

			// override the storage mock if required by the test case.
			if test.Storage != nil {
				p.storage = test.Storage
			}

			// call Present. The token argument can be garbage because the ACME-DNS
			// provider does not use it.
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

// TestRegister tests that the ACME-DNS register function works correctly.
func TestRegister(t *testing.T) {
	testCases := []struct {
		Name          string
		Client        acmeDNSClient
		Storage       goacmedns.Storage
		Domain        string
		FQDN          string
		ExpectedError error
	}{
		{
			Name:          "register when acme-dns client returns an error",
			Client:        errorRegisterClient{},
			ExpectedError: errorClientErr,
		},
		{
			Name:          "register when acme-dns storage put returns an error",
			Client:        mockClient{egTestAccount},
			Storage:       errorPutStorage{mockStorage{make(map[string]goacmedns.Account)}},
			ExpectedError: errorStorageErr,
		},
		{
			Name:          "register when acme-dns storage save returns an error",
			Client:        mockClient{egTestAccount},
			Storage:       errorSaveStorage{mockStorage{make(map[string]goacmedns.Account)}},
			ExpectedError: errorStorageErr,
		},
		{
			Name:   "register when everything works",
			Client: mockClient{egTestAccount},
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
				storage: mockStorage{make(map[string]goacmedns.Account)},
			}

			// override the storage mock if required by the testcase.
			if test.Storage != nil {
				p.storage = test.Storage
			}

			// Call register for the example domain/fqdn.
			err := p.register(context.Background(), egDomain, egFQDN)
			if test.ExpectedError != nil {
				assert.Equal(t, test.ExpectedError, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
