package internal

import (
	"errors"
	"fmt"
	"net/http"

	querystring "github.com/google/go-querystring/query"
)

// DomainService API access to Domain.
type DomainService service

// GetAll domains.
// https://api-docs.constellix.com/?version=latest#484c3f21-d724-4ee4-a6fa-ab22c8eb9e9b
func (s *DomainService) GetAll(params *PaginationParameters) ([]Domain, error) {
	endpoint, err := s.client.createEndpoint(defaultVersion, "domains")
	if err != nil {
		return nil, fmt.Errorf("failed to create request endpoint: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if params != nil {
		v, errQ := querystring.Values(params)
		if errQ != nil {
			return nil, errQ
		}
		req.URL.RawQuery = v.Encode()
	}

	var domains []Domain
	err = s.client.do(req, &domains)
	if err != nil {
		return nil, err
	}

	return domains, nil
}

// GetByName Gets domain by name.
func (s *DomainService) GetByName(domainName string) (Domain, error) {
	domains, err := s.Search(Exact, domainName)
	if err != nil {
		return Domain{}, err
	}

	if len(domains) == 0 {
		return Domain{}, fmt.Errorf("domain not found: %s", domainName)
	}

	if len(domains) > 1 {
		return Domain{}, fmt.Errorf("multiple domains found: %v", domains)
	}

	return domains[0], nil
}

// Search searches for a domain by name.
// https://api-docs.constellix.com/?version=latest#3d7b2679-2209-49f3-b011-b7d24e512008
func (s *DomainService) Search(filter searchFilter, value string) ([]Domain, error) {
	endpoint, err := s.client.createEndpoint(defaultVersion, "domains", "search")
	if err != nil {
		return nil, fmt.Errorf("failed to create request endpoint: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	query := req.URL.Query()
	query.Set(string(filter), value)
	req.URL.RawQuery = query.Encode()

	var domains []Domain
	err = s.client.do(req, &domains)
	if err != nil {
		var nf *NotFound
		if !errors.As(err, &nf) {
			return nil, err
		}
	}

	return domains, nil
}
