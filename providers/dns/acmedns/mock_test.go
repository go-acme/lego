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

type mockClient struct {
	records map[goacmedns.Account]string

	updateTXTRecordCalled bool
	updateTXTRecord       func(ctx context.Context, acct goacmedns.Account, value string) error

	registerAccountCalled bool
	registerAccount       func(ctx context.Context, allowFrom []string) (goacmedns.Account, error)
}

func newMockClient() *mockClient {
	return &mockClient{
		records: make(map[goacmedns.Account]string),
		updateTXTRecord: func(_ context.Context, _ goacmedns.Account, _ string) error {
			return nil
		},
		registerAccount: func(_ context.Context, _ []string) (goacmedns.Account, error) {
			return goacmedns.Account{}, nil
		},
	}
}

func (c *mockClient) UpdateTXTRecord(ctx context.Context, acct goacmedns.Account, value string) error {
	c.updateTXTRecordCalled = true
	c.records[acct] = value

	return c.updateTXTRecord(ctx, acct, value)
}

func (c *mockClient) RegisterAccount(ctx context.Context, allowFrom []string) (goacmedns.Account, error) {
	c.registerAccountCalled = true
	return c.registerAccount(ctx, allowFrom)
}

func (c *mockClient) WithUpdateTXTRecordError(err error) *mockClient {
	c.updateTXTRecord = func(_ context.Context, _ goacmedns.Account, _ string) error {
		return err
	}

	return c
}

func (c *mockClient) WithRegisterAccount(acct goacmedns.Account) *mockClient {
	c.registerAccount = func(_ context.Context, _ []string) (goacmedns.Account, error) {
		return acct, nil
	}

	return c
}

func (c *mockClient) WithRegisterAccountError(err error) *mockClient {
	c.registerAccount = func(_ context.Context, _ []string) (goacmedns.Account, error) {
		return goacmedns.Account{}, err
	}

	return c
}

type mockStorage struct {
	accounts map[string]goacmedns.Account
	fetchAll func(ctx context.Context) (map[string]goacmedns.Account, error)
	fetch    func(ctx context.Context, domain string) (goacmedns.Account, error)
	put      func(ctx context.Context, domain string, acct goacmedns.Account) error
	save     func(ctx context.Context) error
}

func newMockStorage() *mockStorage {
	m := &mockStorage{
		accounts: make(map[string]goacmedns.Account),
		put: func(_ context.Context, _ string, _ goacmedns.Account) error {
			return nil
		},
		save: func(_ context.Context) error {
			return nil
		},
	}

	m.fetchAll = func(ctx context.Context) (map[string]goacmedns.Account, error) {
		return m.accounts, nil
	}

	m.fetch = func(_ context.Context, domain string) (goacmedns.Account, error) {
		if acct, ok := m.accounts[domain]; ok {
			return acct, nil
		}
		return goacmedns.Account{}, storage.ErrDomainNotFound
	}

	return m
}

func (m *mockStorage) FetchAll(ctx context.Context) (map[string]goacmedns.Account, error) {
	return m.fetchAll(ctx)
}

func (m *mockStorage) Fetch(ctx context.Context, domain string) (goacmedns.Account, error) {
	return m.fetch(ctx, domain)
}

func (m *mockStorage) Put(ctx context.Context, domain string, account goacmedns.Account) error {
	return m.put(ctx, domain, account)
}

func (m *mockStorage) Save(ctx context.Context) error {
	return m.save(ctx)
}

func (m *mockStorage) WithAccount(domain string, acct goacmedns.Account) *mockStorage {
	m.accounts[domain] = acct

	return m
}

func (m *mockStorage) WithFetchError(err error) *mockStorage {
	m.fetch = func(_ context.Context, _ string) (goacmedns.Account, error) {
		return goacmedns.Account{}, err
	}

	return m
}

func (m *mockStorage) WithPutError(err error) *mockStorage {
	m.put = func(_ context.Context, _ string, _ goacmedns.Account) error {
		return err
	}

	return m
}

func (m *mockStorage) WithSaveError(err error) *mockStorage {
	m.save = func(ctx context.Context) error {
		return err
	}

	return m
}
