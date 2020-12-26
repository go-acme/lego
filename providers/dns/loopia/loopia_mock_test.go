package loopia

// import (
// 	"errors"
// 	"fmt"
// 	"testing"
//
// 	"github.com/go-acme/lego/v4/providers/dns/loopia/internal"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/mock"
// 	"github.com/stretchr/testify/require"
// )
//
// const (
// 	exampleDomain = "example.com"
// 	exampleRdata  = "LHDhK3oGRvkiefQnx7OOczTY5Tic_xZ6HcMOc_gmtoM"
// 	acmeChallenge = "_acme-challenge"
// )
//
// type mockedClient struct {
// 	mock.Mock
// }
//
// func (c *mockedClient) RemoveTXTRecord(domain string, subdomain string, recordID int) error {
// 	args := c.Called(domain, subdomain, recordID)
// 	return args.Error(0)
// }
//
// func (c *mockedClient) AddTXTRecord(domain string, subdomain string, ttl int, value string) error {
// 	args := c.Called(domain, subdomain, ttl, value)
// 	return args.Error(0)
// }
//
// func (c *mockedClient) GetTXTRecords(domain string, subdomain string) ([]internal.RecordObj, error) {
// 	args := c.Called(domain, subdomain)
// 	return args.Get(0).([]internal.RecordObj), args.Error(1)
// }
//
// func (c *mockedClient) RemoveSubdomain(domain, subdomain string) error {
// 	args := c.Called(domain, subdomain)
// 	return args.Error(0)
// }
//
// func TestDNSProvider_Present(t *testing.T) {
// 	mockedFindZoneByFqdn := func(fqdn string) (string, error) {
// 		return exampleDomain + ".", nil
// 	}
//
// 	testCases := []struct {
// 		name                        string
// 		expectedErrorMsg            string
// 		expectedInProgressTokenInfo int
// 		getTXTRecordsError          error
// 		getTXTRecordsReturn         []internal.RecordObj
// 		addTXTRecordError           error
// 		callAddTXTRecord            bool
// 		callGetTXTRecords           bool
// 	}{
// 		{
// 			name: "Present OK",
// 			getTXTRecordsReturn: []internal.RecordObj{
// 				{Type: "TXT", Rdata: exampleRdata, RecordID: 12345678},
// 			},
// 			callAddTXTRecord:            true,
// 			callGetTXTRecords:           true,
// 			expectedInProgressTokenInfo: 12345678,
// 		},
// 		{
// 			name:              "addTXTRecord fails",
// 			addTXTRecordError: fmt.Errorf("unknown Error: 'ADDTXT'"),
// 			callAddTXTRecord:  true,
// 			expectedErrorMsg:  "unknown Error: 'ADDTXT'",
// 		},
// 		{
// 			name:               "getTXTRecords fails",
// 			getTXTRecordsError: fmt.Errorf("unknown Error: 'GETTXT'"),
// 			callAddTXTRecord:   true,
// 			callGetTXTRecords:  true,
// 			expectedErrorMsg:   "unknown Error: 'GETTXT'",
// 		},
// 		{
// 			name:              "Failed to get ID",
// 			callAddTXTRecord:  true,
// 			callGetTXTRecords: true,
// 			expectedErrorMsg:  "loopia: Failed to get id for TXT record",
// 		},
// 	}
// 	for _, test := range testCases {
// 		t.Run(test.name, func(t *testing.T) {
// 			config := NewDefaultConfig()
// 			config.APIUser = "apiuser"
// 			config.APIPassword = "password"
//
// 			client := &mockedClient{}
// 			provider, _ := NewDNSProviderConfig(config)
// 			provider.findZoneByFqdn = mockedFindZoneByFqdn
// 			provider.client = client
//
// 			if test.callAddTXTRecord {
// 				client.On("addTXTRecord", exampleDomain, acmeChallenge, config.TTL, exampleRdata).Return(test.addTXTRecordError)
// 			}
//
// 			if test.callGetTXTRecords {
// 				client.On("getTXTRecords", exampleDomain, acmeChallenge).Return(test.getTXTRecordsReturn, test.getTXTRecordsError)
// 			}
//
// 			err := provider.Present(exampleDomain, "token", "key")
// 			client.AssertExpectations(t)
//
// 			if test.expectedErrorMsg == "" {
// 				require.NoError(t, err)
// 				assert.Equal(t, test.expectedInProgressTokenInfo, provider.inProgressInfo["token"])
// 			} else {
// 				require.Error(t, err)
// 				assert.EqualError(t, err, test.expectedErrorMsg)
// 			}
// 		})
// 	}
// }
//
// func TestDNSProvider_Cleanup(t *testing.T) {
// 	mockedFindZoneByFqdn := func(fqdn string) (string, error) {
// 		return "example.com.", nil
// 	}
//
// 	testCases := []struct {
// 		name                 string
// 		expectedErrorMsg     string
// 		getTXTRecordsError   error
// 		getTXTRecordsReturn  []internal.RecordObj
// 		removeTXTRecordError error
// 		removeSubdomainError error
// 		callAddTXTRecord     bool
// 		callGetTXTRecords    bool
// 		callRemoveSubdomain  bool
// 	}{
// 		{
// 			name:                "Cleanup Ok",
// 			callAddTXTRecord:    true,
// 			callGetTXTRecords:   true,
// 			callRemoveSubdomain: true,
// 		},
// 		{
// 			name:                 "removeTXTRecord failed",
// 			removeTXTRecordError: errors.New("authentication Error"),
// 			expectedErrorMsg:     "Authentication Error",
// 			callAddTXTRecord:     true,
// 		},
// 		{
// 			name:                 "removeSubdomain failed",
// 			removeSubdomainError: fmt.Errorf("unknown Error: 'UNKNOWN_ERROR'"),
// 			expectedErrorMsg:     "unknown Error: 'UNKNOWN_ERROR'",
// 			callAddTXTRecord:     true,
// 			callGetTXTRecords:    true,
// 			callRemoveSubdomain:  true,
// 		},
// 		{
// 			name:                "Dont call removeSubdomain when records",
// 			getTXTRecordsReturn: []internal.RecordObj{{Type: "TXT", Rdata: "LEFTOVER"}},
// 			callAddTXTRecord:    true,
// 			callGetTXTRecords:   true,
// 			callRemoveSubdomain: false,
// 		},
// 		{
// 			name:                "getTXTRecords failed",
// 			getTXTRecordsError:  fmt.Errorf("unknown Error: 'UNKNOWN_ERROR'"),
// 			expectedErrorMsg:    "unknown Error: 'UNKNOWN_ERROR'",
// 			callAddTXTRecord:    true,
// 			callGetTXTRecords:   true,
// 			callRemoveSubdomain: false,
// 		},
// 	}
// 	for _, test := range testCases {
// 		t.Run(test.name, func(t *testing.T) {
// 			config := NewDefaultConfig()
// 			config.APIUser = "apiuser"
// 			config.APIPassword = "password"
//
// 			client := &mockedClient{}
//
// 			provider, _ := NewDNSProviderConfig(config)
// 			provider.findZoneByFqdn = mockedFindZoneByFqdn
// 			provider.client = client
// 			provider.inProgressInfo["token"] = 12345678
//
// 			if test.callAddTXTRecord {
// 				client.On("removeTXTRecord", "example.com", "_acme-challenge", 12345678).Return(test.removeTXTRecordError)
// 			}
//
// 			if test.callGetTXTRecords {
// 				client.On("getTXTRecords", "example.com", "_acme-challenge").Return(test.getTXTRecordsReturn, test.getTXTRecordsError)
// 			}
//
// 			if test.callRemoveSubdomain {
// 				client.On("removeSubdomain", "example.com", "_acme-challenge").Return(test.removeSubdomainError)
// 			}
//
// 			err := provider.CleanUp("example.com", "token", "key")
//
// 			client.AssertExpectations(t)
//
// 			if test.expectedErrorMsg == "" {
// 				require.NoError(t, err)
// 			} else {
// 				require.Error(t, err)
// 				assert.EqualError(t, err, test.expectedErrorMsg)
// 			}
// 		})
// 	}
// }
