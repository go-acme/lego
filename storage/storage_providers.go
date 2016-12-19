package storage

import (
	"github.com/xenolf/lego/storage/local"
	"github.com/xenolf/lego/storage/s3"
)

type StorageProvider interface {
	CheckPath(string) error
	WritePath(string, []byte) error
	ReadPath(string) ([]byte, error)
}

func NewStorageProvider(name string) (StorageProvider, error) {
	var err error
	var storage StorageProvider
	switch name {
	case "local":
		storage, err = local.NewStorageProvider()
	case "s3":
		storage, err = s3.NewStorageProvider()
	default:
		storage, err = local.NewStorageProvider()
	}
	return storage, err
}
