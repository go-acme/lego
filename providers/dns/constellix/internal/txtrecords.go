package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

// TxtRecordService API access to Record.
type TxtRecordService service

// Create a TXT record.
// https://api-docs.constellix.com/?version=latest#22e24d5b-9ec0-49a7-b2b0-5ff0a28e71be
func (s *TxtRecordService) Create(ctx context.Context, domainID int64, record RecordRequest) ([]Record, error) {
	endpoint, err := s.client.createEndpoint(defaultVersion, "domains", strconv.FormatInt(domainID, 10), "records", "txt")
	if err != nil {
		return nil, fmt.Errorf("failed to create request endpoint: %w", err)
	}

	body, err := json.Marshal(record)
	if err != nil {
		return nil, fmt.Errorf("failed to create request JSON body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	var records []Record
	err = s.client.do(req, &records)
	if err != nil {
		return nil, err
	}

	return records, nil
}

// GetAll TXT records.
// https://api-docs.constellix.com/?version=latest#e7103c53-2ad8-4bc8-b5b3-4c22c4b571b2
func (s *TxtRecordService) GetAll(ctx context.Context, domainID int64) ([]Record, error) {
	endpoint, err := s.client.createEndpoint(defaultVersion, "domains", strconv.FormatInt(domainID, 10), "records", "txt")
	if err != nil {
		return nil, fmt.Errorf("failed to create endpoint: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	var records []Record
	err = s.client.do(req, &records)
	if err != nil {
		return nil, err
	}

	return records, nil
}

// Get a TXT record.
// https://api-docs.constellix.com/?version=latest#e7103c53-2ad8-4bc8-b5b3-4c22c4b571b2
func (s *TxtRecordService) Get(ctx context.Context, domainID, recordID int64) (*Record, error) {
	endpoint, err := s.client.createEndpoint(defaultVersion, "domains", strconv.FormatInt(domainID, 10), "records", "txt", strconv.FormatInt(recordID, 10))
	if err != nil {
		return nil, fmt.Errorf("failed to create request endpoint: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	var records Record
	err = s.client.do(req, &records)
	if err != nil {
		return nil, err
	}

	return &records, nil
}

// Update a TXT record.
// https://api-docs.constellix.com/?version=latest#d4e9ab2e-fac0-45a6-b0e4-cf62a2d2e3da
func (s *TxtRecordService) Update(ctx context.Context, domainID, recordID int64, record RecordRequest) (*SuccessMessage, error) {
	endpoint, err := s.client.createEndpoint(defaultVersion, "domains", strconv.FormatInt(domainID, 10), "records", "txt", strconv.FormatInt(recordID, 10))
	if err != nil {
		return nil, fmt.Errorf("failed to create request endpoint: %w", err)
	}

	body, err := json.Marshal(record)
	if err != nil {
		return nil, fmt.Errorf("failed to create request JSON body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	var msg SuccessMessage
	err = s.client.do(req, &msg)
	if err != nil {
		return nil, err
	}

	return &msg, nil
}

// Delete a TXT record.
// https://api-docs.constellix.com/?version=latest#135947f7-d6c8-481a-83c7-4d387b0bdf9e
func (s *TxtRecordService) Delete(ctx context.Context, domainID, recordID int64) (*SuccessMessage, error) {
	endpoint, err := s.client.createEndpoint(defaultVersion, "domains", strconv.FormatInt(domainID, 10), "records", "txt", strconv.FormatInt(recordID, 10))
	if err != nil {
		return nil, fmt.Errorf("failed to create request endpoint: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	var msg *SuccessMessage
	err = s.client.do(req, &msg)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

// Search searches for a TXT record by name.
// https://api-docs.constellix.com/?version=latest#81003e4f-bd3f-413f-a18d-6d9d18f10201
func (s *TxtRecordService) Search(ctx context.Context, domainID int64, filter searchFilter, value string) ([]Record, error) {
	endpoint, err := s.client.createEndpoint(defaultVersion, "domains", strconv.FormatInt(domainID, 10), "records", "txt", "search")
	if err != nil {
		return nil, fmt.Errorf("failed to create request endpoint: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	query := req.URL.Query()
	query.Set(string(filter), value)
	req.URL.RawQuery = query.Encode()

	var records []Record

	err = s.client.do(req, &records)
	if err != nil {
		var nf *NotFound
		if !errors.As(err, &nf) {
			return nil, err
		}
	}

	return records, nil
}
