package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/zerosnake0/directadmin"
)

const (
	defaultPropagationTimeout = 120 // seconds
)

// DirectAdminDNSProvider is a DNS provider for DirectAdmin.
type DirectAdminDNSProvider struct {
	client *directadmin.Client
}

// NewDirectAdminDNSProvider creates a new instance of the DirectAdminDNSProvider.
func NewDirectAdminDNSProvider(client *directadmin.Client) *DirectAdminDNSProvider {
	return &DirectAdminDNSProvider{
		client: client,
	}
}

// Present implements the dns01.ChallengeProvider interface.
func (d *DirectAdminDNSProvider) Present(domain, token, keyAuth string) error {
	err := d.addTXTRecord(domain, keyAuth)
	if err != nil {
		return fmt.Errorf("failed to create TXT record: %w", err)
	}

	// Wait for DNS propagation
	propagationTimeout := defaultPropagationTimeout
	log.Printf("Waiting %d seconds for DNS propagation...", propagationTimeout)
	// Note: You may need to implement a proper wait mechanism here.

	return nil
}

// CleanUp implements the dns01.ChallengeProvider interface.
func (d *DirectAdminDNSProvider) CleanUp(domain, token, keyAuth string) error {
	err := d.removeTXTRecord(domain, keyAuth)
	if err != nil {
		return fmt.Errorf("failed to remove TXT record: %w", err)
	}

	return nil
}

func (d *DirectAdminDNSProvider) addTXTRecord(domain, keyAuth string) error {
	return d.client.DNSCreateRecord(domain, "TXT", keyAuth)
}

func (d *DirectAdminDNSProvider) removeTXTRecord(domain, keyAuth string) error {
	return d.client.DNSRemoveRecord(domain, "TXT", keyAuth)
}

// main function for testing
func main() {
	// Replace with your DirectAdmin credentials
	daClient := directadmin.NewClient("https://your-directadmin-server.com", "username", "password")

	// Replace with your domain
	domain := "example.com"

	// Replace with your token and keyAuth
	token := "your-token"
	keyAuth := "your-key-auth"

	// Create the DirectAdminDNSProvider
	dnsProvider := NewDirectAdminDNSProvider(daClient)

	// Present the challenge
	err := dns01.AddTXTRecord(context.Background(), dnsProvider, "_acme-challenge."+domain, keyAuth)
	if err != nil {
		log.Fatal(err)
	}

	// Clean up after the challenge
	err = dnsProvider.CleanUp(domain, token, keyAuth)
	if err != nil {
		log.Fatal(err)
	}
}
