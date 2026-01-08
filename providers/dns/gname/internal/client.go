package internal

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-acme/lego/v5/internal/errutils"
	"github.com/go-acme/lego/v5/internal/useragent"
	querystring "github.com/google/go-querystring/query"
)

const defaultBaseURL = "https://api.gname.com"

// Client the GName API client.
type Client struct {
	appID  string
	appKey string

	clock func() time.Time

	BaseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new Client.
func NewClient(appID, appKey string) (*Client, error) {
	if appID == "" || appKey == "" {
		return nil, errors.New("credentials missing")
	}

	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		appID:      appID,
		appKey:     appKey,
		clock:      time.Now,
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

func (c *Client) AddRecord(ctx context.Context, record Record) (int, error) {
	endpoint := c.BaseURL.JoinPath("api", "resolution", "add")

	req, err := c.newRequest(ctx, endpoint, record)
	if err != nil {
		return 0, err
	}

	var result int

	err = c.do(req, &result)
	if err != nil {
		return 0, err
	}

	return result, nil
}

func (c *Client) DeleteRecord(ctx context.Context, domain string, recordID int) error {
	endpoint := c.BaseURL.JoinPath("api", "resolution", "delete")

	record := Record{
		Domain:   domain,
		RecordID: recordID,
	}

	req, err := c.newRequest(ctx, endpoint, record)
	if err != nil {
		return err
	}

	err = c.do(req, nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) do(req *http.Request, result any) error {
	useragent.SetHeader(req.Header)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode/100 != 2 {
		raw, _ := io.ReadAll(resp.Body)

		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	response := &APIResponse{}

	err = json.Unmarshal(raw, response)
	if err != nil {
		return errutils.NewUnmarshalError(req, resp.StatusCode, raw, err)
	}

	if result == nil {
		return nil
	}

	if response.Code != 1 {
		return fmt.Errorf("%d: %s", response.Code, response.Message)
	}

	err = json.Unmarshal(response.Data, result)
	if err != nil {
		return errutils.NewUnmarshalError(req, resp.StatusCode, response.Data, err)
	}

	return nil
}

func (c *Client) newRequest(ctx context.Context, endpoint *url.URL, payload any) (*http.Request, error) {
	values, err := querystring.Values(payload)
	if err != nil {
		return nil, err
	}

	values.Set("appid", c.appID)
	values.Set("gntime", strconv.FormatInt(c.clock().UTC().Unix(), 10))

	sum := md5.Sum([]byte(values.Encode() + c.appKey))
	signature := strings.ToUpper(hex.EncodeToString(sum[:]))

	values.Set("gntoken", signature)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), strings.NewReader(values.Encode()))
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return req, nil
}
