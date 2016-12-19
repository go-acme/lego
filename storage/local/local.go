package local

import (
	"io/ioutil"
	"os"
)

type Storage struct{}

func NewStorageProvider() (*Storage, error) {
	return &Storage{}, nil
}

func (s *Storage) CheckPath(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, 0700)
	}
	return nil
}

func (s *Storage) WritePath(path string, data []byte) error {
	return ioutil.WriteFile(path, data, 0600)
}

func (s *Storage) ReadPath(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}
