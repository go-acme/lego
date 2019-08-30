package autodns

import (
	"fmt"
	"github.com/go-acme/lego/v3/challenge/dns01"
	"github.com/go-acme/lego/v3/platform/config/env"
	"net/http"
	"net/url"
)

const (
	defaultEndpoint string = `https://api.autodns.com/v1`
	demoEndpoint    string = `https://api.demo.autodns.com/v1`

	defaultEndpointContext int = 1
	demoEndpointContext    int = 4
)

type Config struct {
	Endpoint   *url.URL
	Username   string `json:"username"`
	Password   string `json:"password"`
	Context    int    `json:"-"`
	HTTPClient *http.Client
}

func NewDefaultConfig() *Config {
	endpoint, _ := url.Parse(defaultEndpoint)

	return &Config{
		Endpoint: endpoint,
		Context:  defaultEndpointContext,
	}
}

type DNSProvider struct {
	config *Config
}

func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("AUTODNS_API_USER", "AUTODNS_API_PASSWORD")
	if err != nil {
		return nil, fmt.Errorf("autodns: %v", err)
	}

	rawEndpoint := env.GetOrDefaultString("AUTODNS_ENDPOINT", defaultEndpoint)
	endpoint, err := url.Parse(rawEndpoint)
	if err != nil {
		return nil, fmt.Errorf("autodns: %v", err)
	}

	config := NewDefaultConfig()
	config.Username = values["AUTODNS_API_USER"]
	config.Password = values["AUTODNS_API_PASSWORD"]
	config.Endpoint = endpoint

	return &DNSProvider{config: config}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)
	_, err := d.addTxtRecord(domain, fqdn, value)
	if err != nil {
		return fmt.Errorf("autodns: %v", err)
	}
	return nil
}

// CleanUp removes the TXT record previously created
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	panic("implement me")
}
