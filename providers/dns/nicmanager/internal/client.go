package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"

	"github.com/pquerna/otp/totp"
)

const (
	defaultBaseURL  = "https://api.nicmanager.com/v1"
	headerTOTPToken = "X-Auth-Token"
)

// Modes.
const (
	ModeAnycast = "anycast"
	ModeZone    = "zone"
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
	HTTPClient *http.Client
	baseURL    *url.URL

	username string
	password string
	otp      string

	mode string
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

func (c Client) GetZone(name string) (*Zone, error) {
	resp, err := c.do(http.MethodGet, name, nil)
	if err != nil {
		return nil, err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= http.StatusBadRequest {
		b, _ := ioutil.ReadAll(resp.Body)

		msg := APIError{StatusCode: resp.StatusCode}
		if err = json.Unmarshal(b, &msg); err != nil {
			return nil, fmt.Errorf("failed to get zone info for %s", name)
		}

		return nil, msg
	}

	var zone Zone
	err = json.NewDecoder(resp.Body).Decode(&zone)
	if err != nil {
		return nil, err
	}

	return &zone, nil
}

func (c Client) AddRecord(zone string, req RecordCreateUpdate) error {
	resp, err := c.do(http.MethodPost, path.Join(zone, "records"), req)
	if err != nil {
		return err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusAccepted {
		b, _ := ioutil.ReadAll(resp.Body)

		msg := APIError{StatusCode: resp.StatusCode}
		if err = json.Unmarshal(b, &msg); err != nil {
			return fmt.Errorf("records create should've returned %d but returned %d", http.StatusAccepted, resp.StatusCode)
		}

		return msg
	}

	return nil
}

func (c Client) DeleteRecord(zone string, record int) error {
	resp, err := c.do(http.MethodDelete, path.Join(zone, "records", strconv.Itoa(record)), nil)
	if err != nil {
		return err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusAccepted {
		b, _ := ioutil.ReadAll(resp.Body)

		msg := APIError{StatusCode: resp.StatusCode}
		if err = json.Unmarshal(b, &msg); err != nil {
			return fmt.Errorf("records delete should've returned %d but returned %d", http.StatusAccepted, resp.StatusCode)
		}

		return msg
	}

	return nil
}

func (c Client) do(method, uri string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonValue, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}

		reqBody = bytes.NewBuffer(jsonValue)
	}

	endpoint, err := c.baseURL.Parse(path.Join(c.baseURL.Path, c.mode, uri))
	if err != nil {
		return nil, err
	}

	r, err := http.NewRequest(method, endpoint.String(), reqBody)
	if err != nil {
		return nil, err
	}

	r.Header.Set("Accept", "application/json")
	r.Header.Set("Content-Type", "application/json")

	r.SetBasicAuth(c.username, c.password)

	if c.otp != "" {
		tan, err := totp.GenerateCode(c.otp, time.Now())
		if err != nil {
			return nil, err
		}

		r.Header.Set(headerTOTPToken, tan)
	}

	return c.HTTPClient.Do(r)
}
