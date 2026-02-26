package dnspersist01

import (
	"context"
	"fmt"

	"github.com/miekg/dns"
)

// RecordMatcher returns true when the expected record is present.
type RecordMatcher func(records []TXTRecord) bool

// PreCheckFunc checks DNS propagation before notifying ACME that the challenge is ready.
type PreCheckFunc func(ctx context.Context, fqdn string, matcher RecordMatcher) (bool, error)

// WrapPreCheckFunc wraps a PreCheckFunc in order to do extra operations before or after
// the main check, put it in a loop, etc.
type WrapPreCheckFunc func(ctx context.Context, domain, fqdn string, matcher RecordMatcher, check PreCheckFunc) (bool, error)

// WrapPreCheck Allow to define checks before notifying ACME that the challenge is ready.
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

func (p preCheck) call(ctx context.Context, domain, fqdn string, matcher RecordMatcher) (bool, error) {
	if p.checkFunc == nil {
		return p.checkDNSPropagation(ctx, fqdn, matcher)
	}

	return p.checkFunc(ctx, domain, fqdn, matcher, p.checkDNSPropagation)
}

func (p preCheck) checkDNSPropagation(ctx context.Context, fqdn string, matcher RecordMatcher) (bool, error) {
	client := DefaultClient()

	// Initial attempt to resolve at the recursive NS (require to get CNAME)
	result, err := client.LookupTXT(ctx, fqdn)
	if err != nil {
		return false, fmt.Errorf("initial recursive nameserver: %w", err)
	}

	effectiveFQDN := dns.Fqdn(fqdn)
	if len(result.CNAMEChain) > 0 {
		effectiveFQDN = result.CNAMEChain[len(result.CNAMEChain)-1]
	}

	if p.requireRecursiveNssPropagation {
		_, err = client.checkRecursiveNameserversPropagation(ctx, effectiveFQDN, matcher)
		if err != nil {
			return false, fmt.Errorf("recursive nameservers: %w", err)
		}
	}

	if !p.requireAuthoritativeNssPropagation {
		return true, nil
	}

	found, err := client.checkAuthoritativeNameserversPropagation(ctx, effectiveFQDN, matcher)
	if err != nil {
		return found, fmt.Errorf("authoritative nameservers: %w", err)
	}

	return found, nil
}
