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
	"strconv"
	"time"

	"github.com/go-acme/lego/v5/internal/errutils"
	"github.com/go-acme/lego/v5/internal/useragent"
	querystring "github.com/google/go-querystring/query"
)

const defaultBaseURL = "https://api.netcup.com/v1"

// Client the Netcup v1 API client.
type Client struct {
	apiKey string

	BaseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(apiKey string) (*Client, error) {
	if apiKey == "" {
		return nil, errors.New("credentials missing")
	}

	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		apiKey:     apiKey,
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// AddChallenge creates a TXT record related to a challenge.
// https://www.netcup.com/en/helpcenter/documentation/netcup-api#/v1%2Fdomain/NetcupApiV1DomainAcmePost
func (c *Client) AddChallenge(ctx context.Context, domainID int64, scope, value string) ([]ChallengeData, error) {
	endpoint := c.BaseURL.JoinPath("domains", strconv.FormatInt(domainID, 10), "acme", "challenge")

	if scope == "" {
		scope = "@"
	}

	data := ChallengeData{Scope: scope, Value: value}

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, data)
	if err != nil {
		return nil, err
	}

	result := new(APIResponse[ChallengeData])

	err = c.do(req, result)
	if err != nil {
		return nil, err
	}

	return result.Result, nil
}

// DeleteChallenge deletes a TXT record related to a challenge.
// https://www.netcup.com/en/helpcenter/documentation/netcup-api#/v1%2Fdomain/NetcupApiV1DomainAcmeDelete
func (c *Client) DeleteChallenge(ctx context.Context, domainID int64, scope, value string) error {
	endpoint := c.BaseURL.JoinPath("domains", strconv.FormatInt(domainID, 10), "acme", "challenge", scope, value)

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	err = c.do(req, nil)
	if err != nil {
		return err
	}

	return nil
}

// GetDomains gets all domains.
// https://www.netcup.com/en/helpcenter/documentation/netcup-api#/v1%2Fdomain/NetcupApiV1DomainGetList
func (c *Client) GetDomains(ctx context.Context, request *DomainsRequest) ([]Domain, error) {
	endpoint := c.BaseURL.JoinPath("domains")

	values, err := querystring.Values(request)
	if err != nil {
		return nil, err
	}

	endpoint.RawQuery = values.Encode()

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	result := new(APIResponse[Domain])

	err = c.do(req, result)
	if err != nil {
		return nil, err
	}

	return result.Result, nil
}

func (c *Client) do(req *http.Request, result any) error {
	useragent.SetHeader(req.Header)

	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode/100 != 2 {
		return parseError(req, resp)
	}

	if result == nil {
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

func parseError(req *http.Request, resp *http.Response) error {
	raw, _ := io.ReadAll(resp.Body)

	apiResp := new(APIResponse[any])

	err := json.Unmarshal(raw, apiResp)
	if err != nil {
		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	var errAPI error
	for _, apiError := range apiResp.Errors {
		errAPI = errors.Join(errAPI, apiError)
	}

	return errAPI
}
