package softlayer

import (
	"fmt"

	"github.com/softlayer/softlayer-go/services"
)

func (d *DNSProvider) addTXTRecord(fqdn, domain, value string, ttl int) error {
	domainService := &domainService{
		client: services.GetDnsDomainService(d.client),
	}
	if err := domainService.setDomainID(domain); err != nil {
		return fmt.Errorf("softlayer: %w", err)
	}
	if _, err := domainService.client.CreateTxtRecord(&fqdn, &value, &ttl); err != nil {
		return fmt.Errorf("softlayer: %w", err)
	}
	return nil
}

func (d *DNSProvider) cleanupTXTRecord(fqdn, domain string) error {
	domainService := &domainService{
		client: services.GetDnsDomainService(d.client),
	}
	if err := domainService.setDomainID(domain); err != nil {
		return fmt.Errorf("softlayer: %w", err)
	}

	records, err := domainService.findTxtRecords(fqdn)
	if err != nil {
		return fmt.Errorf("softlayer: %w", err)
	}

	if err := domainService.deleteResourceRecords(records); err != nil {
		return err
	}
	return nil
}
