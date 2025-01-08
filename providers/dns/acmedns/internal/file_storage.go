package internal

import (
	"context"
	"os"

	"github.com/cpu/goacmedns"
)

// FileStorage wraps [goacmedns.Storage] for file to implement [acmedns.Storage].
type FileStorage struct {
	storage goacmedns.Storage
}

// NewFileStorage creates a new [FileStorage].
func NewFileStorage(path string, mode os.FileMode) *FileStorage {
	return &FileStorage{
		storage: goacmedns.NewFileStorage(path, mode),
	}
}

// NewStorageWrapper wraps a new [goacmedns.Storage].
func NewStorageWrapper(storage goacmedns.Storage) *FileStorage {
	return &FileStorage{
		storage: storage,
	}
}

func (s *FileStorage) Save(_ context.Context) error {
	return s.storage.Save()
}

func (s *FileStorage) Put(_ context.Context, domain string, account goacmedns.Account) error {
	return s.storage.Put(domain, account)
}

func (s *FileStorage) Fetch(_ context.Context, domain string) (goacmedns.Account, error) {
	return s.storage.Fetch(domain)
}

func (s *FileStorage) FetchAll(_ context.Context) (map[string]goacmedns.Account, error) {
	return s.storage.FetchAll(), nil
}
