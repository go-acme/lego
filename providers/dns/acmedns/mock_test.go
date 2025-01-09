package acmedns

import (
	"context"
	"errors"

	"github.com/nrdcg/goacmedns"
	"github.com/nrdcg/goacmedns/storage"
)

var (
	// errorClientErr is used by the Client mocks that return an error.
	errorClientErr = errors.New("errorClient always errors")
	// errorStorageErr is used by the Storage mocks that return an error.
	errorStorageErr = errors.New("errorStorage always errors")
)

var egTestAccount = goacmedns.Account{
	FullDomain: "acme-dns." + egDomain,
	SubDomain:  "random-looking-junk." + egDomain,
	Username:   "spooky.mulder",
	Password:   "trustno1",
}

// mockClient is a mock implementing the acmeDNSClient interface that always
// returns a fixed goacmedns.Account from calls to Register.
type mockClient struct {
	mockAccount goacmedns.Account
}

// UpdateTXTRecord does nothing.
func (c mockClient) UpdateTXTRecord(_ context.Context, _ goacmedns.Account, _ string) error {
	return nil
}

// RegisterAccount returns c.mockAccount and no errors.
func (c mockClient) RegisterAccount(_ context.Context, _ []string) (goacmedns.Account, error) {
	return c.mockAccount, nil
}

// mockUpdateClient is a mock implementing the acmeDNSClient interface that
// tracks the calls to UpdateTXTRecord in the records map.
type mockUpdateClient struct {
	mockClient
	records map[goacmedns.Account]string
}

// UpdateTXTRecord saves a record value to c.records for the given acct.
func (c mockUpdateClient) UpdateTXTRecord(_ context.Context, acct goacmedns.Account, value string) error {
	c.records[acct] = value
	return nil
}

// errorUpdateClient is a mock implementing the acmeDNSClient interface that always
// returns errors from errorUpdateClient.
type errorUpdateClient struct {
	mockClient
}

// UpdateTXTRecord always returns an error.
func (c errorUpdateClient) UpdateTXTRecord(_ context.Context, _ goacmedns.Account, _ string) error {
	return errorClientErr
}

// errorRegisterClient is a mock implementing the acmeDNSClient interface that always
// returns errors from RegisterAccount.
type errorRegisterClient struct {
	mockClient
}

// RegisterAccount always returns an error.
func (c errorRegisterClient) RegisterAccount(_ context.Context, _ []string) (goacmedns.Account, error) {
	return goacmedns.Account{}, errorClientErr
}

// mockStorage is a mock implementing the goacmedns.Storage interface that
// returns static account data and ignores Save.
type mockStorage struct {
	accounts map[string]goacmedns.Account
}

// Save does nothing.
func (m mockStorage) Save(_ context.Context) error {
	return nil
}

// Put stores an account for the given domain in m.accounts.
func (m mockStorage) Put(_ context.Context, domain string, acct goacmedns.Account) error {
	m.accounts[domain] = acct
	return nil
}

// Fetch retrieves an account for the given domain from m.accounts or returns
// goacmedns.ErrDomainNotFound.
func (m mockStorage) Fetch(_ context.Context, domain string) (goacmedns.Account, error) {
	if acct, ok := m.accounts[domain]; ok {
		return acct, nil
	}
	return goacmedns.Account{}, storage.ErrDomainNotFound
}

// FetchAll returns all of m.accounts.
func (m mockStorage) FetchAll(_ context.Context) (map[string]goacmedns.Account, error) {
	return m.accounts, nil
}

// errorPutStorage is a mock implementing the goacmedns.Storage interface that
// always returns errors from Put.
type errorPutStorage struct {
	mockStorage
}

// Put always errors.
func (e errorPutStorage) Put(_ context.Context, _ string, _ goacmedns.Account) error {
	return errorStorageErr
}

// errorSaveStorage is a mock implementing the goacmedns.Storage interface that
// always returns errors from Save.
type errorSaveStorage struct {
	mockStorage
}

// Save always errors.
func (e errorSaveStorage) Save(_ context.Context) error {
	return errorStorageErr
}

// errorFetchStorage is a mock implementing the goacmedns.Storage interface that
// always returns errors from Fetch.
type errorFetchStorage struct {
	mockStorage
}

// Fetch always errors.
func (e errorFetchStorage) Fetch(_ context.Context, _ string) (goacmedns.Account, error) {
	return goacmedns.Account{}, errorStorageErr
}

// FetchAll is a nop for errorFetchStorage.
func (e errorFetchStorage) FetchAll(_ context.Context) (map[string]goacmedns.Account, error) {
	return nil, nil
}
