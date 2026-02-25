package dnspersist

/**
 * NOTE(ldez): This implementation is internal because the [challenge.Persistent] interface is experimental.
 * This will evolve in the future.
 */

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-acme/lego/v5/acme"
	"github.com/go-acme/lego/v5/challenge"
	"github.com/go-acme/lego/v5/challenge/dnspersist01"
	"github.com/go-acme/lego/v5/internal/env"
)

// Environment variables names.
const (
	envNamespace = "DNSPERSIST_MANUAL_"

	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
)

const (
	// defaultPropagationTimeout default propagation timeout.
	defaultPropagationTimeout = 60 * time.Second

	// defaultPollingInterval default polling interval.
	defaultPollingInterval = 2 * time.Second
)

const maxTXTStringOctets = 255

var _ challenge.PersistentProvider = (*Provider)(nil)

type Provider struct {
	propagationTimeout  time.Duration
	propagationInterval time.Duration
}

func NewProvider() *Provider {
	return &Provider{
		propagationTimeout:  env.GetOrDefaultSecond(EnvPropagationTimeout, defaultPropagationTimeout),
		propagationInterval: env.GetOrDefaultSecond(EnvPollingInterval, defaultPollingInterval),
	}
}

func (p *Provider) Persist(_ context.Context, authz acme.Authorization, issuerDomainName, accountURI string, persistUntil time.Time) error {
	info, err := dnspersist01.GetChallengeInfo(authz, issuerDomainName, accountURI, persistUntil)
	if err != nil {
		return err
	}

	fmt.Println("lego: Please create a TXT record with the following value:")
	fmt.Printf("%s IN TXT %s\n", info.FQDN, formatTXTValue(info.Value))
	fmt.Println("lego: Press 'Enter' once the record is available.")

	_, err = bufio.NewReader(os.Stdin).ReadBytes('\n')
	if err != nil {
		return fmt.Errorf("dnspersist01-manual: read stdin: %w", err)
	}

	return nil
}

func (p *Provider) Timeout() (timeout, interval time.Duration) {
	return p.propagationTimeout, p.propagationInterval
}

// formatTXTValue formats a TXT record value for display,
// splitting it into multiple quoted strings if it exceeds 255 octets,
// as per RFC 1035.
func formatTXTValue(value string) string {
	chunks := splitTXTValue(value)
	if len(chunks) == 1 {
		return fmt.Sprintf("%q", chunks[0])
	}

	parts := make([]string, 0, len(chunks))
	for _, chunk := range chunks {
		parts = append(parts, fmt.Sprintf("%q", chunk))
	}

	return strings.Join(parts, " ")
}

// splitTXTValue splits a TXT value into RFC 1035 <character-string> chunks of at most 255 octets
// so long TXT values can be represented as multiple strings in one RR.
func splitTXTValue(value string) []string {
	if len(value) <= maxTXTStringOctets {
		return []string{value}
	}

	var chunks []string
	for len(value) > maxTXTStringOctets {
		chunks = append(chunks, value[:maxTXTStringOctets])
		value = value[maxTXTStringOctets:]
	}

	if value != "" {
		chunks = append(chunks, value)
	}

	return chunks
}
