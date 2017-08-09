// Package pipe implements a DNS provider for solving the DNS-01 challenge using
// a named pipe.
package pipe

import (
	"bufio"
	"fmt"
	"os"

	"github.com/xenolf/lego/acme"
)

// DNSProvider is an implementation of the acme.ChallengeProvider interface.
type DNSProvider struct {
	pipe *os.File
}

// NewDNSProvider returns a DNSProvider instance with an opened pipe.
func NewDNSProvider() (*DNSProvider, error) {
	pipePath := os.Getenv("PIPE_PATH")
	if pipeFd, err := os.OpenFile(pipePath, os.O_RDWR, os.ModeNamedPipe); err != nil {
		return nil, err
	} else {
		c := &DNSProvider{
			pipe: pipeFd,
		}
		return c, nil
	}
}

// Present creates a TXT record to fulfil the DNS-01 challenge.
func (c *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value, ttl := acme.DNS01Record(domain, keyAuth)

	sendMsg := fmt.Sprintf("PRESENT %s %s %s %d\n", domain, fqdn, value, ttl)
	if _, err := c.pipe.WriteString(sendMsg); err != nil {
		return fmt.Errorf("Writing 'present' operation failed: %s", err)
	}

	return c.pipeResponse()
}

// CleanUp removes the TXT record matching the specified parameters.
func (c *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _, _ := acme.DNS01Record(domain, keyAuth)

	sendMsg := fmt.Sprintf("CLEANUP %s %s\n", domain, fqdn)
	if _, err := c.pipe.WriteString(sendMsg); err != nil {
		return fmt.Errorf("Writing 'cleanup' operation failed: %s", err)
	}

	return c.pipeResponse()
}

// Handle response from pipe
func (c *DNSProvider) pipeResponse() error {
	reader := bufio.NewReader(c.pipe)
	if recvMsg, err := reader.ReadString('\n'); err != nil {
		return fmt.Errorf("Failed to read result of 'present' operation: %s", err)
	} else if recvMsg[:3] == "ERR" {
		return fmt.Errorf("Operation 'present' failed: %s", recvMsg[4:len(recvMsg)])
	} else if recvMsg[:3] != "OKY" {
		return fmt.Errorf("Operation 'present' returned strange result")
	}
	return nil
}
