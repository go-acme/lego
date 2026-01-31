package zerossl

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-acme/lego/v5/internal/errutils"
	"github.com/go-acme/lego/v5/internal/useragent"
)

const EnvZeroSSLAccessKey = "ZERO_SSL_ACCESS_KEY"

const defaultBaseURL = "https://api.zerossl.com"

// Client is a ZeroSSL API client.
type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
}

// NewClient returns a new ZeroSSL API client.
func NewClient() *Client {
	baseURL, _ := url.Parse(defaultBaseURL)

	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// GenerateEAB generates a new EAB credential.
func (c *Client) GenerateEAB(ctx context.Context, accessKey string) (*APIResponse, error) {
	endpoint := c.baseURL.JoinPath("acme", "eab-credentials")

	query := endpoint.Query()
	query.Set("access_key", accessKey)

	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	return c.do(req)
}

func (c *Client) GenerateEABFromEmail(ctx context.Context, email string) (*APIResponse, error) {
	if email == "" {
		return nil, errors.New("no email provided")
	}

	endpoint := c.baseURL.JoinPath("acme", "eab-credentials-email")

	payload := url.Values{}
	payload.Set("email", email)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), strings.NewReader(payload.Encode()))
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return c.do(req)
}

func (c *Client) do(req *http.Request) (*APIResponse, error) {
	useragent.SetHeader(req.Header)

	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode/100 != 2 {
		return nil, parseError(req, resp)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	result := new(APIResponse)

	err = json.Unmarshal(raw, result)
	if err != nil {
		return nil, errutils.NewUnmarshalError(req, resp.StatusCode, raw, err)
	}

	if !result.Success {
		return nil, result.Error
	}

	return result, nil
}

func parseError(req *http.Request, resp *http.Response) error {
	raw, _ := io.ReadAll(resp.Body)

	errAPI := new(APIResponse)

	err := json.Unmarshal(raw, errAPI)
	if err != nil {
		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	return errAPI.Error
}

func IsZeroSSL(server string) bool {
	return strings.HasPrefix(server, "https://acme.zerossl.com/")
}
