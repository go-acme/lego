package exec

import (
	"errors"
	"os"
	"os/exec"
	"strconv"

	"github.com/xenolf/lego/acme"
)

// DNSProvider adds and removes the record for the DNS challenge by calling a
// program with command-line parameters.
type DNSProvider struct {
	program string
}

// NewDNSProvider returns a new DNS provider which runs the program in the
// environment variable EXEC_PATH for adding and removing the DNS record.
func NewDNSProvider() (*DNSProvider, error) {
	s := os.Getenv("EXEC_PATH")
	if s == "" {
		return nil, errors.New("environment variable EXEC_PATH not set")
	}

	return NewDNSProviderProgram(s)
}

// NewDNSProviderProgram returns a new DNS provider which runs the given program
// for adding and removing the DNS record.
func NewDNSProviderProgram(program string) (*DNSProvider, error) {
	return &DNSProvider{program: program}, nil
}

// Present creates a TXT record to fulfil the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	var cmd *exec.Cmd
	m := os.Getenv("EXEC_MODE")

	if m == "RAW" {
		cmd = exec.Command(d.program, "present", "--", domain, token, keyAuth)
	} else {
		fqdn, value, ttl := acme.DNS01Record(domain, keyAuth)
		cmd = exec.Command(d.program, "present", fqdn, value, strconv.Itoa(ttl))
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// CleanUp removes the TXT record matching the specified parameters
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, value, ttl := acme.DNS01Record(domain, keyAuth)
	cmd := exec.Command(d.program, "cleanup", fqdn, value, strconv.Itoa(ttl))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
