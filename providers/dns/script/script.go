// Package script implements a DNS provider for solving the DNS-01 challenge using
// a user-provided script.
package script

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/xenolf/lego/acme"
)

// DNSProvider is an implementation of the acme.ChallengeProvider interface.
type DNSProvider struct {
	script string
}

// NewDNSProvider returns a DNSProvider instance with a script defined.
func NewDNSProvider() (*DNSProvider, error) {
	scriptCmd := os.Getenv("SCRIPT_PATH")
	if scriptPath, err := exec.LookPath(scriptCmd); err != nil {
		return nil, fmt.Errorf(err.Error())
	} else {
		c := &DNSProvider{
			script: scriptPath,
		}
		return c, nil
	}
}

// Present creates a TXT record to fulfil the DNS-01 challenge.
func (c *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, ttl := acme.DNS01Record(domain, keyAuth)

	cmd := exec.Command(c.script, "present", domain, fqdn, value, strconv.Itoa(ttl))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Script failed at 'present' phase: %s", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (c *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _, _ := acme.DNS01Record(domain, keyAuth)

	cmd := exec.Command(c.script, "cleanup", domain, fqdn)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Script failed at 'cleanup' phase: %s", err)
	}
	return nil
}
