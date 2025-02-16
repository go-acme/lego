package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
	"github.com/nrdcg/goacmedns"
	"github.com/nrdcg/goacmedns/storage"
)

var _ goacmedns.Storage = (*HTTPStorage)(nil)

var ErrCNAMEAlreadyCreated = errors.New("the CNAME has already been created")

// HTTPStorage is an implementation of [acmedns.Storage] over HTTP.
type HTTPStorage struct {
	client  *http.Client
	baseURL *url.URL
}

// NewHTTPStorage created a new [HTTPStorage].
func NewHTTPStorage(baseURL string) (*HTTPStorage, error) {
	endpoint, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	return &HTTPStorage{
		client:  &http.Client{Timeout: 2 * time.Minute},
		baseURL: endpoint,
	}, nil
}

func (s *HTTPStorage) Save(_ context.Context) error {
	return nil
}

func (s *HTTPStorage) Put(ctx context.Context, domain string, account goacmedns.Account) error {
	req, err := newJSONRequest(ctx, http.MethodPost, s.baseURL.JoinPath(domain), account)
	if err != nil {
		return fmt.Errorf("unable to create request: %w", err)
	}

	return s.do(req, nil)
}

func (s *HTTPStorage) Fetch(ctx context.Context, domain string) (goacmedns.Account, error) {
	req, err := newJSONRequest(ctx, http.MethodGet, s.baseURL.JoinPath(domain), nil)
	if err != nil {
		return goacmedns.Account{}, fmt.Errorf("unable to create request: %w", err)
	}

	var account goacmedns.Account

	err = s.do(req, &account)
	if err != nil {
		return goacmedns.Account{}, err
	}

	return account, nil
}

func (s *HTTPStorage) FetchAll(ctx context.Context) (map[string]goacmedns.Account, error) {
	req, err := newJSONRequest(ctx, http.MethodGet, s.baseURL, nil)
	if err != nil {
		return nil, err
	}

	var mapping map[string]goacmedns.Account

	err = s.do(req, &mapping)
	if err != nil {
		return nil, err
	}

	return mapping, nil
}

func (s *HTTPStorage) do(req *http.Request, result any) error {
	resp, err := s.client.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return storage.ErrDomainNotFound
	}

	if resp.StatusCode/100 != 2 {
		return errutils.NewUnexpectedResponseStatusCodeError(req, resp)
	}

	if result == nil {
		// Hack related to `Put`.
		if resp.StatusCode == http.StatusCreated {
			return ErrCNAMEAlreadyCreated
		}

		return nil
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	err = json.Unmarshal(raw, result)
	if err != nil {
		return errutils.NewUnmarshalError(req, resp.StatusCode, raw, err)
	}

	return nil
}

func newJSONRequest(ctx context.Context, method string, endpoint *url.URL, payload any) (*http.Request, error) {
	buf := new(bytes.Buffer)

	if payload != nil {
		err := json.NewEncoder(buf).Encode(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to create request JSON body: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint.String(), buf)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}
