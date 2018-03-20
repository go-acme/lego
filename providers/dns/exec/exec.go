// Package exec implements a DNS provider for solving the DNS-01 challenge using
// a user-provided executable.
package exec

import (
	"fmt"
	"os"
	osexec "os/exec"
	"strconv"

	"github.com/xenolf/lego/acme"
)

// DNSProvider is an implementation of the acme.ChallengeProvider interface.
type DNSProvider struct {
	exec string
}

// NewDNSProvider returns a DNSProvider instance with a script defined.
func NewDNSProvider() (*DNSProvider, error) {
	cmd := os.Getenv("EXEC_PATH")
	if execPath, err := osexec.LookPath(cmd); err != nil {
		return nil, err
	} else {
		c := &DNSProvider{
			exec: execPath,
		}
		return c, nil
	}
}

// Present creates a TXT record to fulfil the DNS-01 challenge.
func (c *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, ttl := acme.DNS01Record(domain, keyAuth)

	cmd := osexec.Command(c.exec, "present", domain, fqdn, value, strconv.Itoa(ttl))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Command %s failed at 'present' phase: %s", c.exec, err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (c *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _, _ := acme.DNS01Record(domain, keyAuth)

	cmd := osexec.Command(c.exec, "cleanup", domain, fqdn)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Command %s failed at 'cleanup' phase: %s", c.exec, err)
	}
	return nil
}
