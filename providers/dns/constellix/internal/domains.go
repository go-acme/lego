package internal

import (
	"fmt"
	"net/http"

	querystring "github.com/google/go-querystring/query"
)

// DomainService API access to Domain.
type DomainService service

// GetID for a domain name.
func (s *DomainService) GetID(domainName string) (int64, error) {
	params := &PaginationParameters{
		Offset: 0,
		Max:    100,
		Sort:   "name",
		Order:  "asc",
	}

	domains, err := s.GetAll(params)
	if err != nil {
		return 0, err
	}

	for len(domains) > 0 {
		for _, domain := range domains {
			if domain.Name == domainName {
				return domain.ID, nil
			}
		}

		if params.Max > len(domains) {
			break
		}

		params = &PaginationParameters{
			Offset: params.Max,
			Max:    100,
			Sort:   "name",
			Order:  "asc",
		}

		domains, err = s.GetAll(params)
		if err != nil {
			return 0, err
		}
	}

	return 0, fmt.Errorf("domain not found: %s", domainName)
}

// GetAll domains.
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
