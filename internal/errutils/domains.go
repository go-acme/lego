package errutils

import (
	"fmt"
	"maps"
	"slices"
	"sort"
	"strings"
)

type DomainsError struct {
	prefix string
	data   map[string]error
}

func NewDomainsError(prefix string) *DomainsError {
	return &DomainsError{
		prefix: prefix,
		data:   make(map[string]error),
	}
}

func (e *DomainsError) Add(domain string, err error) {
	e.data[domain] = err
}

func (e *DomainsError) Has(domain string) bool {
	_, ok := e.data[domain]

	return ok
}

func (e *DomainsError) Join() error {
	if e == nil || len(e.data) == 0 {
		return nil
	}

	return e
}

func (e *DomainsError) Error() string {
	buffer := new(strings.Builder)

	_, _ = fmt.Fprintf(buffer, "%s: one or more domains had a problem:", e.prefix)

	var domains []string
	for domain := range e.data {
		domains = append(domains, domain)
	}

	sort.Strings(domains)

	for _, domain := range domains {
		_, _ = fmt.Fprintf(buffer, " [%s: %s]", domain, e.data[domain])
	}

	return buffer.String()
}

func (e *DomainsError) Unwrap() []error {
	if e == nil || len(e.data) == 0 {
		return nil
	}

	return slices.AppendSeq(make([]error, 0, len(e.data)), maps.Values(e.data))
}
