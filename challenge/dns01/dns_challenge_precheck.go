package dns01

import (
	"context"
	"fmt"

	"github.com/miekg/dns"
)

// PreCheckFunc checks DNS propagation before notifying ACME that the DNS challenge is ready.
type PreCheckFunc func(ctx context.Context, fqdn, value string) (bool, error)

// WrapPreCheckFunc wraps a PreCheckFunc in order to do extra operations before or after
// the main check, put it in a loop, etc.
type WrapPreCheckFunc func(ctx context.Context, domain, fqdn, value string, check PreCheckFunc) (bool, error)

// WrapPreCheck Allow to define checks before notifying ACME that the DNS challenge is ready.
func WrapPreCheck(wrap WrapPreCheckFunc) ChallengeOption {
	return func(chlg *Challenge) error {
		chlg.preCheck.checkFunc = wrap
		return nil
	}
}

type preCheck struct {
	// checks DNS propagation before notifying ACME that the DNS challenge is ready.
	checkFunc WrapPreCheckFunc

	// require the TXT record to be propagated to all authoritative name servers
	requireAuthoritativeNssPropagation bool

	// require the TXT record to be propagated to all recursive name servers
	requireRecursiveNssPropagation bool
}

func newPreCheck() preCheck {
	return preCheck{
		requireAuthoritativeNssPropagation: true,
		requireRecursiveNssPropagation:     true,
	}
}

func (p preCheck) call(ctx context.Context, domain, fqdn, value string) (bool, error) {
	if p.checkFunc == nil {
		return p.checkDNSPropagation(ctx, fqdn, value)
	}

	return p.checkFunc(ctx, domain, fqdn, value, p.checkDNSPropagation)
}

// checkDNSPropagation checks if the expected TXT record has been propagated to all authoritative nameservers.
func (p preCheck) checkDNSPropagation(ctx context.Context, fqdn, value string) (bool, error) {
	client := DefaultClient()

	// Initial attempt to resolve at the recursive NS (require getting CNAME)
	r, err := client.sendQuery(ctx, fqdn, dns.TypeTXT, true)
	if err != nil {
		return false, fmt.Errorf("initial recursive nameserver: %w", err)
	}

	if r.Rcode == dns.RcodeSuccess {
		fqdn = updateDomainWithCName(r, fqdn)
	}

	if p.requireRecursiveNssPropagation {
		_, err = client.checkNameserversPropagation(ctx, fqdn, value, false)
		if err != nil {
			return false, fmt.Errorf("recursive nameservers: %w", err)
		}
	}

	if !p.requireAuthoritativeNssPropagation {
		return true, nil
	}

	authoritativeNss, err := client.lookupAuthoritativeNameservers(ctx, fqdn)
	if err != nil {
		return false, err
	}

	found, err := client.checkNameserversPropagationCustom(ctx, fqdn, value, authoritativeNss, true)
	if err != nil {
		return found, fmt.Errorf("authoritative nameservers: %w", err)
	}

	return found, nil
}
