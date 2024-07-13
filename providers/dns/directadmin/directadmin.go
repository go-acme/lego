package directadmin

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/log"
)

// DNSProvider implements the challenge.Provider interface
type DNSProvider struct {
	apiURL   string
	username string
	password string
}

// NewDNSProvider creates a new DNSProvider instance
func NewDNSProvider() (*DNSProvider, error) {
	apiURL := getEnv("DIRECTADMIN_API_URL", "https://api.directadmin.com")
	username := getEnv("DIRECTADMIN_USERNAME", "default_username")
	password := getEnv("DIRECTADMIN_PASSWORD", "default_password")

	return &DNSProvider{
		apiURL:   apiURL,
		username: username,
		password: password,
	}, nil
}

// getEnv reads an environment variable and returns the value or a default value if the environment variable is not set.
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// Present creates a TXT record to fulfill the DNS-01 challenge
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	url := fmt.Sprintf("%s/CMD_API_DNS_CONTROL?domain=%s&json=yes", d.apiURL, domain)
	data := fmt.Sprintf("action=add&type=TXT&name=_acme-challenge&value=%s", info.Value)

	req, err := http.NewRequest("POST", url, strings.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(d.username, d.password)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
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
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	url := fmt.Sprintf("%s/CMD_API_DNS_CONTROL?domain=%s&json=yes", d.apiURL, domain)
	data := fmt.Sprintf("action=delete&type=TXT&name=_acme-challenge&value=%s", info.Value)

	req, err := http.NewRequest("POST", url, strings.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(d.username, d.password)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
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
