package internal

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

const defaultBaseURL = "https://dyn.dns.he.net/nic/update"

const (
	codeGood     = "good"
	codeNoChg    = "nochg"
	codeAbuse    = "abuse"
	codeBadAgent = "badagent"
	codeBadAuth  = "badauth"
	codeInterval = "interval"
	codeNoHost   = "nohost"
	codeNotFqdn  = "notfqdn"
)

const defaultBurst = 5

// Client the Hurricane Electric client.
type Client struct {
	HTTPClient   *http.Client
	rateLimiters sync.Map

	baseURL string

	credentials map[string]string
	credMu      sync.Mutex
}

// NewClient Creates a new Client.
func NewClient(credentials map[string]string) *Client {
	return &Client{
		HTTPClient:  &http.Client{Timeout: 5 * time.Second},
		baseURL:     defaultBaseURL,
		credentials: credentials,
	}
}

// UpdateTxtRecord updates a TXT record.
func (c *Client) UpdateTxtRecord(ctx context.Context, domain string, txt string) error {
	hostname := fmt.Sprintf("_acme-challenge.%s", domain)

	c.credMu.Lock()
	token, ok := c.credentials[domain]
	c.credMu.Unlock()

	if !ok {
		return fmt.Errorf("hurricane: Domain %s not found in credentials, check your credentials map", domain)
	}

	data := url.Values{}
	data.Set("password", token)
	data.Set("hostname", hostname)
	data.Set("txt", txt)

	rl, _ := c.rateLimiters.LoadOrStore(hostname, rate.NewLimiter(limit(defaultBurst), defaultBurst))

	err := rl.(*rate.Limiter).Wait(ctx)
	if err != nil {
		return err
	}

	resp, err := c.HTTPClient.PostForm(c.baseURL, data)
	if err != nil {
		return err
	}

	defer func() { _ = resp.Body.Close() }()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	body := string(bytes.TrimSpace(bodyBytes))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%d: attempt to change TXT record %s returned %s", resp.StatusCode, hostname, body)
	}

	return evaluateBody(body, hostname)
}

func evaluateBody(body string, hostname string) error {
	code, _, _ := strings.Cut(body, " ")

	switch code {
	case codeGood:
		return nil
	case codeNoChg:
		log.Printf("%s: unchanged content written to TXT record %s", body, hostname)
		return nil
	case codeAbuse:
		return fmt.Errorf("%s: blocked hostname for abuse: %s", body, hostname)
	case codeBadAgent:
		return fmt.Errorf("%s: user agent not sent or HTTP method not recognized; open an issue on go-acme/lego on Github", body)
	case codeBadAuth:
		return fmt.Errorf("%s: wrong authentication token provided for TXT record %s", body, hostname)
	case codeInterval:
		return fmt.Errorf("%s: TXT records update exceeded API rate limit", body)
	case codeNoHost:
		return fmt.Errorf("%s: the record provided does not exist in this account: %s", body, hostname)
	case codeNotFqdn:
		return fmt.Errorf("%s: the record provided isn't an FQDN: %s", body, hostname)
	default:
		// This is basically only server errors.
		return fmt.Errorf("attempt to change TXT record %s returned %s", hostname, body)
	}
}

// limit computes the rate based on burst.
// The API rate limit per-record is 10 reqs / 2 minutes.
//
//	10 reqs / 2 minutes = freq 1/12 (burst = 1)
//	6 reqs / 2 minutes = freq 1/20 (burst = 5)
//
// https://github.com/go-acme/lego/issues/1415
func limit(burst int) rate.Limit {
	return 1 / rate.Limit(120/(10-burst+1))
}
