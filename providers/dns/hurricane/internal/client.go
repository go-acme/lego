package internal

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const defaultBaseURL = "https://dyn.dns.he.net/nic/update"

const (
	codeGood     = "good"
	codeNoChg    = "nochg"
	codeAbuse    = "abuse"
	codeBadAgent = "badagent"
	codeBadAuth  = "badauth"
	codeNoHost   = "nohost"
	codeNotFqdn  = "notfqdn"
)

// Client the Hurricane Electric client.
type Client struct {
	HTTPClient *http.Client
	baseURL    string

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
func (c *Client) UpdateTxtRecord(domain string, txt string) error {
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

	resp, err := c.HTTPClient.PostForm(c.baseURL, data)
	if err != nil {
		return err
	}

	defer func() { _ = resp.Body.Close() }()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
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
	words := strings.SplitN(body, " ", 2)

	switch words[0] {
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
	case codeNoHost:
		return fmt.Errorf("%s: the record provided does not exist in this account: %s", body, hostname)
	case codeNotFqdn:
		return fmt.Errorf("%s: the record provided isn't an FQDN: %s", body, hostname)
	default:
		// This is basically only server errors.
		return fmt.Errorf("attempt to change TXT record %s returned %s", hostname, body)
	}
}
