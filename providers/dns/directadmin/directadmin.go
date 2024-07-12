package directadmin

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/log"
)

// DNSProviderDirectAdmin implements the challenge.Provider interface
type DNSProviderDirectAdmin struct {
	apiURL   string
	username string
	password string
	timeout  time.Duration
}

// NewDNSProviderDirectAdmin creates a new DNSProviderDirectAdmin instance
func NewDNSProviderDirectAdmin(apiURL, username, password string, timeout time.Duration) (*DNSProviderDirectAdmin, error) {
	return &DNSProviderDirectAdmin{
		apiURL:   apiURL,
		username: username,
		password: password,
		timeout:  timeout,
	}, nil
}

// Present creates a TXT record to fulfill the DNS-01 challenge
func (d *DNSProviderDirectAdmin) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	url := fmt.Sprintf("%s/CMD_API_DNS_CONTROL?domain=%s&json=yes", d.apiURL, domain)
	data := fmt.Sprintf("action=add&type=TXT&name=_acme-challenge&value=%s", info.Value)

	req, err := http.NewRequest("POST", url, strings.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(d.username, d.password)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: d.timeout}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	log.Infof("Presented TXT record for domain %s", domain)
	return nil
}

// CleanUp removes the TXT record created for the DNS-01 challenge
func (d *DNSProviderDirectAdmin) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	url := fmt.Sprintf("%s/CMD_API_DNS_CONTROL?domain=%s&json=yes", d.apiURL, domain)
	data := fmt.Sprintf("action=delete&type=TXT&name=_acme-challenge&value=%s", info.Value)

	req, err := http.NewRequest("POST", url, strings.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(d.username, d.password)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: d.timeout}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	log.Infof("Cleaned up TXT record for domain %s", domain)
	return nil
}
