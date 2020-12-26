package loopia

import (
	"errors"
	"fmt"
	"testing"

	"github.com/go-acme/lego/v4/providers/dns/loopia/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	exampleDomain    = "example.com"
	exampleSubDomain = "_acme-challenge"
	exampleRdata     = "LHDhK3oGRvkiefQnx7OOczTY5Tic_xZ6HcMOc_gmtoM"
)

func TestDNSProvider_Present(t *testing.T) {
	mockedFindZoneByFqdn := func(fqdn string) (string, error) {
		return exampleDomain + ".", nil
	}

	testCases := []struct {
		desc string

		getTXTRecordsError  error
		getTXTRecordsReturn []internal.RecordObj
		addTXTRecordError   error
		callAddTXTRecord    bool
		callGetTXTRecords   bool

		expectedError               string
		expectedInProgressTokenInfo int
	}{
		{
			desc: "Present OK",

			getTXTRecordsReturn: []internal.RecordObj{{Type: "TXT", Rdata: exampleRdata, RecordID: 12345678}},
			callAddTXTRecord:    true,
			callGetTXTRecords:   true,

			expectedInProgressTokenInfo: 12345678,
		},
		{
			desc: "AddTXTRecord fails",

			addTXTRecordError: fmt.Errorf("unknown error: 'ADDTXT'"),
			callAddTXTRecord:  true,

			expectedError: "loopia: failed to add TXT record: unknown error: 'ADDTXT'",
		},
		{
			desc: "GetTXTRecords fails",

			getTXTRecordsError: fmt.Errorf("unknown error: 'GETTXT'"),
			callAddTXTRecord:   true,
			callGetTXTRecords:  true,

			expectedError: "loopia: failed to get TXT records: unknown error: 'GETTXT'",
		},
		{
			desc: "Failed to get ID",

			callAddTXTRecord:  true,
			callGetTXTRecords: true,

			expectedError: "loopia: failed to find the stored TXT record",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.APIUser = "apiuser"
			config.APIPassword = "password"

			client := &mockedClient{}

			provider, err := NewDNSProviderConfig(config)
			require.NoError(t, err)

			provider.findZoneByFqdn = mockedFindZoneByFqdn
			provider.client = client

			if test.callAddTXTRecord {
				client.On("AddTXTRecord", exampleDomain, exampleSubDomain, config.TTL, exampleRdata).Return(test.addTXTRecordError)
			}

			if test.callGetTXTRecords {
				client.On("GetTXTRecords", exampleDomain, exampleSubDomain).Return(test.getTXTRecordsReturn, test.getTXTRecordsError)
			}

			err = provider.Present(exampleDomain, "token", "key")

			client.AssertExpectations(t)

			if test.expectedError == "" {
				require.NoError(t, err)
				assert.Equal(t, test.expectedInProgressTokenInfo, provider.inProgressInfo["token"])
			} else {
				require.Error(t, err)
				assert.EqualError(t, err, test.expectedError)
			}
		})
	}
}

func TestDNSProvider_Cleanup(t *testing.T) {
	mockedFindZoneByFqdn := func(fqdn string) (string, error) {
		return "example.com.", nil
	}

	testCases := []struct {
		desc string

		getTXTRecordsError   error
		getTXTRecordsReturn  []internal.RecordObj
		removeTXTRecordError error
		removeSubdomainError error
		callAddTXTRecord     bool
		callGetTXTRecords    bool
		callRemoveSubdomain  bool

		expectedError string
	}{
		{
			desc: "Cleanup Ok",

			callAddTXTRecord:    true,
			callGetTXTRecords:   true,
			callRemoveSubdomain: true,
		},
		{
			desc: "removeTXTRecord failed",

			removeTXTRecordError: errors.New("authentication error"),
			callAddTXTRecord:     true,

			expectedError: "loopia: failed to remove TXT record: authentication error",
		},
		{
			desc: "removeSubdomain failed",

			removeSubdomainError: errors.New(`unknown error: "UNKNOWN_ERROR"`),
			callAddTXTRecord:     true,
			callGetTXTRecords:    true,
			callRemoveSubdomain:  true,

			expectedError: `loopia: failed to remove sub-domain: unknown error: "UNKNOWN_ERROR"`,
		},
		{
			desc: "Dont call removeSubdomain when records",

			getTXTRecordsReturn: []internal.RecordObj{{Type: "TXT", Rdata: "LEFTOVER"}},
			callAddTXTRecord:    true,
			callGetTXTRecords:   true,
			callRemoveSubdomain: false,
		},
		{
			desc: "getTXTRecords failed",

			getTXTRecordsError:  errors.New(`unknown error: "UNKNOWN_ERROR"`),
			callAddTXTRecord:    true,
			callGetTXTRecords:   true,
			callRemoveSubdomain: false,

			expectedError: `loopia: failed to get TXT records: unknown error: "UNKNOWN_ERROR"`,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.APIUser = "apiuser"
			config.APIPassword = "password"

			client := &mockedClient{}

			provider, err := NewDNSProviderConfig(config)
			require.NoError(t, err)

			provider.findZoneByFqdn = mockedFindZoneByFqdn
			provider.client = client
			provider.inProgressInfo["token"] = 12345678

			if test.callAddTXTRecord {
				client.On("RemoveTXTRecord", "example.com", "_acme-challenge", 12345678).Return(test.removeTXTRecordError)
			}

			if test.callGetTXTRecords {
				client.On("GetTXTRecords", "example.com", "_acme-challenge").Return(test.getTXTRecordsReturn, test.getTXTRecordsError)
			}

			if test.callRemoveSubdomain {
				client.On("RemoveSubdomain", "example.com", "_acme-challenge").Return(test.removeSubdomainError)
			}

			err = provider.CleanUp("example.com", "token", "key")

			client.AssertExpectations(t)

			if test.expectedError == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.EqualError(t, err, test.expectedError)
			}
		})
	}
}

type mockedClient struct {
	mock.Mock
}

func (c *mockedClient) RemoveTXTRecord(domain string, subdomain string, recordID int) error {
	args := c.Called(domain, subdomain, recordID)
	return args.Error(0)
}

func (c *mockedClient) AddTXTRecord(domain string, subdomain string, ttl int, value string) error {
	args := c.Called(domain, subdomain, ttl, value)
	return args.Error(0)
}

func (c *mockedClient) GetTXTRecords(domain string, subdomain string) ([]internal.RecordObj, error) {
	args := c.Called(domain, subdomain)
	return args.Get(0).([]internal.RecordObj), args.Error(1)
}

func (c *mockedClient) RemoveSubdomain(domain, subdomain string) error {
	args := c.Called(domain, subdomain)
	return args.Error(0)
}
