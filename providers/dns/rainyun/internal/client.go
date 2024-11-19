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
	"reflect"
	"strconv"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

const defaultBaseURL = "https://api.v2.rainyun.com/product/"

// Client the Rain Yun API client.
type Client struct {
	apiKey string

	baseURL    *url.URL
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
		baseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

func (c *Client) AddRecord(ctx context.Context, domainID int, record AddRecord) error {
	endpoint := c.baseURL.JoinPath("domain", strconv.Itoa(domainID), "dns")

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, record)
	if err != nil {
		return err
	}

	err = c.do(req, nil)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) DeleteRecord(ctx context.Context, domainID, recordID int) error {
	endpoint := c.baseURL.JoinPath("domain", strconv.Itoa(domainID), "dns")

	req, err := newQueryRequest(ctx, http.MethodDelete, endpoint, DelRecord{ID: recordID})
	if err != nil {
		return err
	}

	err = c.do(req, nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) ListRecords(ctx context.Context, domainID int) ([]Record, error) {
	endpoint := c.baseURL.JoinPath("domain", strconv.Itoa(domainID), "dns")

	query := endpoint.Query()
	query.Set("limit", "100")
	query.Set("page_no", "1")
	endpoint.RawQuery = query.Encode()

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var recordData APIResponse[Record]
	err = c.do(req, &recordData)
	if err != nil {
		return nil, err
	}

	return recordData.Data.Records, nil
}

func (c *Client) ListDomains(ctx context.Context) ([]Domain, error) {
	endpoint := c.baseURL.JoinPath("domain")

	query := endpoint.Query()
	query.Set("options", `{"columnFilters":{"domains.Domain":""},"sort":[],"page":1,"perPage":100}`)
	endpoint.RawQuery = query.Encode()

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var domainData APIResponse[Domain]

	err = c.do(req, &domainData)
	if err != nil {
		return nil, err
	}

	return domainData.Data.Records, nil
}

func (c *Client) do(req *http.Request, result any) error {
	req.Header.Add("x-api-key", c.apiKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode > http.StatusBadRequest {
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

func newQueryRequest(ctx context.Context, method string, endpoint *url.URL, payload any) (*http.Request, error) {
	v := url.Values{}
	if payload != nil {
		sv := reflect.ValueOf(payload)
		for i := 0; i < sv.Type().NumField(); i++ {
			field := sv.Type().Field(i)
			tag := field.Tag.Get("form")
			if tag == "" {
				continue
			}
			value := sv.Field(i).Interface()
			switch vt := value.(type) {
			case int:
				v.Add(tag, strconv.Itoa(vt))
			case string:
				v.Add(tag, vt)
			// 可以根据需要支持更多类型
			default:
				return nil, fmt.Errorf("unsupported type %T", vt)
			}
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, fmt.Sprintf("%s?%s", endpoint.String(), v.Encode()), nil)
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

	var errAPI APIError
	err := json.Unmarshal(raw, &errAPI)
	if err != nil {
		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	return &errAPI
}
