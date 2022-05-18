// Package iijdpf implements a DNS provider for solving the DNS-01 challenge using IIJ DNS Platform Service.
package iijdpf

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/miekg/dns"
	dpfapi "github.com/mimuret/golang-iij-dpf/pkg/api"
	dpfzones "github.com/mimuret/golang-iij-dpf/pkg/apis/dpf/v1/zones"
	dpfapiutils "github.com/mimuret/golang-iij-dpf/pkg/apiutils"
	dpftypes "github.com/mimuret/golang-iij-dpf/pkg/types"
)

// Environment variables names.
const (
	envNamespace = "IIJ_DPF_"

	EnvAPIToken = envNamespace + "API_TOKEN"
	EnvServiceCode        = envNamespace + "DPM_SERVICE_CODE"

	EnvAPIEndpoint        = envNamespace + "API_ENDPOINT"
	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	Token       string
	ServiceCode string

	Endpoint           string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	TTL                int
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		Endpoint:           env.GetOrDefaultString(EnvAPIEndpoint, dpfapi.DefaultEndpoint),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 660 * time.Second),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, 5*time.Second),
		TTL:                env.GetOrDefaultInt(EnvTTL, 300),
	}
}

var _ challenge.Provider = &DNSProvider{}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	client dpfapi.ClientInterface
	config *Config
}

// NewDNSProvider returns a DNSProvider instance configured for IIJ DNS.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIToken, EnvServiceCode)
	if err != nil {
		return nil, fmt.Errorf("iijdpf: %w", err)
	}

	config := NewDefaultConfig()
	config.Token = values[EnvAPIToken]
	config.ServiceCode = values[EnvServiceCode]

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig takes a given config
// and returns a custom configured DNSProvider instance.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config.Token == "" {
		return nil, errors.New("iijdpf: API token missing")
	}
	if config.ServiceCode == "" {
		return nil, errors.New("iijdpf: Servicecode missing")
	}

	return &DNSProvider{
		client: dpfapi.NewClient(config.Token, config.Endpoint, nil),
		config: config,
	}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

func (d *DNSProvider)setup(domain, keyAuth string) (string,string,string, error) {
	fqdn, value := dns01.GetRecord(domain, keyAuth)
	fqdn = dns.CanonicalName(fqdn)
	rdata := `"` + value + `"`
	zoneID, err := dpfapiutils.GetZoneIdFromServiceCode(context.Background(),d.client,d.config.ServiceCode)
	return fqdn,rdata,zoneID,err
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn,rdata,zoneID,err := d.setup(domain,keyAuth)
	if err != nil {
		return fmt.Errorf("iijdpf: failed to get zone id: %w", err)
	}
	err = d.addTxtRecord(zoneID, fqdn, rdata)
	if err != nil {
		return fmt.Errorf("iijdpf: %w", err)
	}
	err = d.commit(zoneID)
	if err != nil {
		return fmt.Errorf("iijdpf: %w", err)
	}
	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn,rdata,zoneID,err := d.setup(domain,keyAuth)
	if err != nil {
		return fmt.Errorf("iijdpf: failed to get zone id: %w", err)
	}
	err = d.deleteTxtRecord(zoneID, fqdn, rdata)
	if err != nil {
		return fmt.Errorf("iijdpf: %w", err)
	}
	err = d.commit(zoneID)
	if err != nil {
		return fmt.Errorf("iijdpf: %w", err)
	}
	return nil
}

func (d *DNSProvider) addTxtRecord(zoneID, fqdn, rdata string) error {
	r,err := dpfapiutils.GetRecordFromZoneID(context.Background(),d.client,zoneID,fqdn,dpfzones.TypeTXT)
	if err != nil && !errors.Is(err, dpfapiutils.ErrRecordNotFound) {
		return err
	}
	if r != nil {
		r.RData = append(r.RData, dpfzones.RecordRDATA{
			Value: rdata,
		})
		if _, _, err := dpfapiutils.SyncUpdate(context.Background(), d.client, r, nil); err != nil {
			return fmt.Errorf("failed to update record: %w", err)
		}
	} else {
		txt := &dpfzones.Record{
			AttributeMeta: dpfzones.AttributeMeta{
				ZoneID: zoneID,
			},
			Name:          fqdn,
			TTL:           dpftypes.NullablePositiveInt32(d.config.TTL),
			RRType:        dpfzones.TypeTXT,
			RData: dpfzones.RecordRDATASlice{
				dpfzones.RecordRDATA{Value: rdata},
			},
			Description: "ACME",
		}
		if _, _, err := dpfapiutils.SyncCreate(context.Background(), d.client, txt, nil); err != nil {
			return fmt.Errorf("failed to create record: %w", err)
		}
	}
	return nil
}

func (d *DNSProvider) deleteTxtRecord(zoneID, fqdn, rdata string) error {
	r,err := dpfapiutils.GetRecordFromZoneID(context.Background(),d.client,zoneID,fqdn,dpfzones.TypeTXT)
	if err != nil {
		if errors.Is(err, dpfapiutils.ErrRecordNotFound) {
			// empty target rrset
			return nil
		}
		// api error
		return err
	}
	if len(r.RData) == 1 {
		// delete rrset
		if _, _, err := dpfapiutils.SyncDelete(context.Background(), d.client, r); err != nil {
			return fmt.Errorf("failed to delete record: %w", err)
		}
	} else {
		// delete rdata
		rdataSlice := dpfzones.RecordRDATASlice{}
		for _, v := range r.RData {
			if v.Value != rdata {
				rdataSlice = append(rdataSlice, v)
			}
		}
		r.RData = rdataSlice
		if _, _, err := dpfapiutils.SyncUpdate(context.Background(), d.client, r, nil); err != nil {
			return fmt.Errorf("failed to update record: %w", err)
		}
	}

	return nil
}

func (d *DNSProvider) commit(zoneID string) error {
	apply := &dpfzones.ZoneApply{
		AttributeMeta: dpfzones.AttributeMeta{
			ZoneID: zoneID,
		},
		Description: "ACME Processing",
	}
	if _, _, err := dpfapiutils.SyncApply(context.Background(), d.client, apply, nil); err != nil {
		return fmt.Errorf("failed to apply zone: %w", err)
	}
	return nil
}
