package mythicbeasts

import "fmt"

// Logs into mythic beasts and acquires a bearer token for use in future
// API calls
func (d *DNSProvider) login() error {
	return fmt.Errorf("mythicbeasts: login() not implemented")
}

func (d *DNSProvider) createTXTRecord(zone string, leaf string, value string) error {
	return fmt.Errorf("mythicbeasts: createTXTRecord() not implemented")
}

func (d *DNSProvider) removeTXTRecord(zone string, leaf string, value string) error {
	return fmt.Errorf("mythicbeasts: removeTXTRecord() not implemented")
}
