package autodns

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/go-acme/lego/v3/challenge/dns01"
	"github.com/go-acme/lego/v3/platform/config/env"
)

const (
	envAPIUser            = `AUTODNS_API_USER`
	envAPIPassword        = `AUTODNS_API_PASSWORD`
	envAPIEndpoint        = `AUTODNS_ENDPOINT`
	envAPIEndpointContext = `AUTODNS_CONTEXT`
	envTTL                = `AUTODNS_TTL`
	envPropagationTimeout = `AUTODNS_PROPAGATION_TIMEOUT`
	envPollingInterval    = `AUTODNS_POLLING_INTERVAL`

	defaultEndpoint = `https://api.autodns.com/v1/`
	demoEndpoint    = `https://api.demo.autodns.com/v1/`

	defaultEndpointContext int = 4
	defaultTTL             int = 600
)

type Config struct {
	Endpoint           *url.URL
	Username           string        `json:"username"`
	Password           string        `json:"password"`
	Context            int           `json:"-"`
	TTL                int           `json:"-"`
	PropagationTimeout time.Duration `json:"-"`
	PollingInterval    time.Duration `json:"-"`
	HTTPClient         *http.Client
}

func NewDefaultConfig() *Config {
	endpoint, _ := url.Parse(env.GetOrDefaultString(envAPIEndpoint, defaultEndpoint))

	return &Config{
		Endpoint:           endpoint,
		Context:            env.GetOrDefaultInt(envAPIEndpointContext, defaultEndpointContext),
		TTL:                env.GetOrDefaultInt(envTTL, defaultTTL),
		PropagationTimeout: env.GetOrDefaultSecond(envPropagationTimeout, 2*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond(envPollingInterval, 2*time.Second),
		HTTPClient:         &http.Client{},
	}
}

type DNSProvider struct {
	config *Config
	//zoneNameservers map[string]string
	//currentRecords  []*ResourceRecord
}

func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(envAPIUser, envAPIPassword)
	if err != nil {
		return nil, fmt.Errorf("autodns: %v", err)
	}

	config := NewDefaultConfig()
	config.Username = values[envAPIUser]
	config.Password = values[envAPIPassword]

	provider := &DNSProvider{config: config}

	// Because autodns needs the nameservers for each request, we query them all here and put them
	// in our state to avoid making a lot of requests later.
	// FIXME: This should become obsolete once I figure out how the _stream endpoint works.
	/*req, err := provider.makeRequest(http.MethodPost, path.Join("zone", "_search"), nil)
	if err != nil {
		return nil, fmt.Errorf("autodns: %v", err)
	}

	var resp *DataZoneResponse
	if err := provider.sendRequest(req, &resp); err != nil {
		return nil, fmt.Errorf("autodns: %v", err)
	}

	provider.zoneNameservers = make(map[string]string, len(resp.Data))

	for _, zone := range resp.Data {
		provider.zoneNameservers[zone.Name] = zone.VirtualNameServer
	}*/

	return provider, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge
func (d *DNSProvider) Present(domain, token, keyAuth string) (err error) {

	fqdn, value := dns01.GetRecord(domain, keyAuth)
	_, err = d.addTxtRecord(domain, fqdn, value)
	if err != nil {
		return fmt.Errorf("autodns: %v", err)
	}
	return nil
}

// CleanUp removes the TXT record previously created
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	if err := d.removeTXTRecord(domain, "_acme-challenge"); err != nil {
		return fmt.Errorf("autodns: removeTXTRecord: %v", err)
	}

	return nil
}
