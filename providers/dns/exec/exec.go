// Package exec implements a DNS provider which runs a program for adding/removing the DNS record.
package exec

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/xenolf/lego/challenge/dns01"
	"github.com/xenolf/lego/log"
	"github.com/xenolf/lego/platform/config/env"
)

type timeoutConfig struct {
	Timeout, Interval int
}

// Config Provider configuration.
type Config struct {
	Program string
	Mode    string
}

// DNSProvider adds and removes the record for the DNS challenge by calling a
// program with command-line parameters.
type DNSProvider struct {
	config *Config
}

// NewDNSProvider returns a new DNS provider which runs the program in the
// environment variable EXEC_PATH for adding and removing the DNS record.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get("EXEC_PATH")
	if err != nil {
		return nil, fmt.Errorf("exec: %v", err)
	}

	return NewDNSProviderConfig(&Config{
		Program: values["EXEC_PATH"],
		Mode:    os.Getenv("EXEC_MODE"),
	})
}

// NewDNSProviderConfig returns a new DNS provider which runs the given configuration
// for adding and removing the DNS record.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("the configuration is nil")
	}

	return &DNSProvider{config: config}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	var args []string
	if d.config.Mode == "RAW" {
		args = []string{"present", "--", domain, token, keyAuth}
	} else {
		fqdn, value := dns01.GetRecord(domain, keyAuth)
		args = []string{"present", fqdn, value}
	}

	cmd := exec.Command(d.config.Program, args...)

	output, err := cmd.CombinedOutput()
	if len(output) > 0 {
		log.Println(string(output))
	}

	return err
}

// CleanUp removes the TXT record matching the specified parameters
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	var args []string
	if d.config.Mode == "RAW" {
		args = []string{"cleanup", "--", domain, token, keyAuth}
	} else {
		fqdn, value := dns01.GetRecord(domain, keyAuth)
		args = []string{"cleanup", fqdn, value}
	}

	cmd := exec.Command(d.config.Program, args...)

	output, err := cmd.CombinedOutput()
	if len(output) > 0 {
		log.Println(string(output))
	}

	return err
}

func (d *DNSProvider) Timeout() (time.Duration, time.Duration) {
	cmd := exec.Command(d.config.Program, "timeout")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return handleTimeoutError(err, output)
	}

	var tc timeoutConfig
	err = json.Unmarshal(output, &tc)
	if err != nil {
		return handleTimeoutError(err, output)
	}

	return time.Duration(tc.Timeout) * time.Second, time.Duration(tc.Interval) * time.Second
}

func handleTimeoutError(err error, output []byte) (time.Duration, time.Duration) {
	log.Infof("fallback to default timeout and interval because command 'timeout' failed: %v", err)
	if len(output) > 0 {
		log.Println(string(output))
	}

	return dns01.DefaultPropagationTimeout, dns01.DefaultPollingInterval
}
