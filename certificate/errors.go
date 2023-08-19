package certificate

import (
	"errors"
	"fmt"
)

type obtainError struct {
	data map[string]error
}

func newObtainError() *obtainError {
	return &obtainError{data: make(map[string]error)}
}

func (e *obtainError) Add(domain string, err error) {
	e.data[domain] = err
}

func (e *obtainError) Join() error {
	if e == nil {
		return nil
	}

	if len(e.data) == 0 {
		return nil
	}

	var err error
	for d, e := range e.data {
		err = errors.Join(err, fmt.Errorf("%s: %w", d, e))
	}

	return fmt.Errorf("error: one or more domains had a problem:\n%w", err)
}

type domainError struct {
	Domain string
	Error  error
}
