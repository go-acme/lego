package acmedns

import (
	"context"

	"github.com/cpu/goacmedns"
)

// Storage same as [goacmedns.Storage] but with context and errors.
type Storage interface {
	Save(ctx context.Context) error
	Put(ctx context.Context, domain string, account goacmedns.Account) error
	Fetch(ctx context.Context, domain string) (goacmedns.Account, error)
	FetchAll(ctx context.Context) (map[string]goacmedns.Account, error)
}
