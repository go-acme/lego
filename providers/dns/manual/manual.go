package manual

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/go-acme/lego/v5/challenge/dnsnew"
)

const (
	dnsTemplate = `%s %d IN TXT %q`
)

// DNSProvider is an implementation of the ChallengeProvider interface.
type DNSProvider struct{}

// NewDNSProvider returns a DNSProvider instance.
func NewDNSProvider() (*DNSProvider, error) {
	return &DNSProvider{}, nil
}

// Present prints instructions for manually creating the TXT record.
func (*DNSProvider) Present(domain, token, keyAuth string) error {
	ctx := context.Background()
	info := dnsnew.GetChallengeInfo(ctx, domain, keyAuth)

	authZone, err := dnsnew.DefaultClient().FindZoneByFqdn(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("manual: could not find zone: %w", err)
	}

	fmt.Printf("lego: Please create the following TXT record in your %s zone:\n", authZone)
	fmt.Printf(dnsTemplate+"\n", info.EffectiveFQDN, dnsnew.DefaultTTL, info.Value)
	fmt.Printf("lego: Press 'Enter' when you are done\n")

	_, err = bufio.NewReader(os.Stdin).ReadBytes('\n')
	if err != nil {
		return fmt.Errorf("manual: %w", err)
	}

	return nil
}

// CleanUp prints instructions for manually removing the TXT record.
func (*DNSProvider) CleanUp(domain, token, keyAuth string) error {
	ctx := context.Background()
	info := dnsnew.GetChallengeInfo(ctx, domain, keyAuth)

	authZone, err := dnsnew.DefaultClient().FindZoneByFqdn(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("manual: could not find zone: %w", err)
	}

	fmt.Printf("lego: You can now remove this TXT record from your %s zone:\n", authZone)
	fmt.Printf(dnsTemplate+"\n", info.EffectiveFQDN, dnsnew.DefaultTTL, "...")

	return nil
}

// Sequential All DNS challenges for this provider will be resolved sequentially.
// Returns the interval between each iteration.
func (d *DNSProvider) Sequential() time.Duration {
	return dnsnew.DefaultPropagationTimeout
}
