package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
	"github.com/pquerna/otp/totp"
)

const (
	defaultBaseURL  = "https://api.nicmanager.com/v1"
	headerTOTPToken = "X-Auth-Token"
)

// Modes.
const (
	ModeAnycast = "anycast"
	ModeZone    = "zones"
)

// Options the Client options.
type Options struct {
	Login    string
	Username string

	Email string

	Password string
	OTP      string

	Mode string
}

// Client a nicmanager DNS client.
type Client struct {
	username string
	password string
	otp      string

	mode string

	baseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient create a new Client.
func NewClient(opts Options) *Client {
	c := &Client{
		mode:       ModeAnycast,
		username:   opts.Email,
		password:   opts.Password,
		otp:        opts.OTP,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}

	c.baseURL, _ = url.Parse(defaultBaseURL)

	if opts.Mode != "" {
		c.mode = opts.Mode
	}

	if opts.Login != "" && opts.Username != "" {
		c.username = fmt.Sprintf("%s.%s", opts.Login, opts.Username)
	}

	return c
}

func (c *Client) GetZone(ctx context.Context, name string) (*Zone, error) {
	endpoint := c.baseURL.JoinPath(c.mode, name)

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var zone Zone
	err = c.do(req, http.StatusOK, &zone)
	if err != nil {
		return nil, err
	}

	return &zone, nil
}

func (c *Client) AddRecord(ctx context.Context, zone string, payload RecordCreateUpdate) error {
	endpoint := c.baseURL.JoinPath(c.mode, zone, "records")

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, payload)
	if err != nil {
		return err
	}

	err = c.do(req, http.StatusAccepted, nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) DeleteRecord(ctx context.Context, zone string, record int) error {
	endpoint := c.baseURL.JoinPath(c.mode, zone, "records", strconv.Itoa(record))

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	err = c.do(req, http.StatusAccepted, nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) do(req *http.Request, expectedStatusCode int, result any) error {
	req.SetBasicAuth(c.username, c.password)

	if c.otp != "" {
		tan, err := totp.GenerateCode(c.otp, time.Now())
		if err != nil {
			return err
		}

		req.Header.Set(headerTOTPToken, tan)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != expectedStatusCode {
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

	return err
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

	errAPI := APIError{StatusCode: resp.StatusCode}
	if err := json.Unmarshal(raw, &errAPI); err != nil {
		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	return errAPI
}
