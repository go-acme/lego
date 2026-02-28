package dnspersist01

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sort"
	"time"

	"github.com/go-acme/lego/v5/acme"
	"github.com/go-acme/lego/v5/acme/api"
	"github.com/go-acme/lego/v5/challenge"
	"github.com/go-acme/lego/v5/internal/wait"
	"github.com/go-acme/lego/v5/log"
	"github.com/miekg/dns"
)

const validationLabel = "_validation-persist"

// ValidateFunc validates a challenge with the ACME server.
type ValidateFunc func(ctx context.Context, core *api.Core, domain string, chlng acme.Challenge) error

// Challenge implements the dns-persist-01 challenge.
type Challenge struct {
	core     *api.Core
	validate ValidateFunc
	provider challenge.PersistentProvider
	preCheck preCheck

	userSuppliedIssuerDomainName string
	persistUntil                 time.Time
}

// NewChallenge creates a dns-persist-01 challenge.
func NewChallenge(core *api.Core, validate ValidateFunc, provider challenge.PersistentProvider, opts ...ChallengeOption) (*Challenge, error) {
	chlg := &Challenge{
		core:     core,
		validate: validate,
		provider: provider,
		preCheck: newPreCheck(),
	}

	for _, opt := range opts {
		err := opt(chlg)
		if err != nil {
			log.Warn("dnspersist01: challenge option skipped.", log.ErrorAttr(err))
		}
	}

	return chlg, nil
}

// Solve validates the dns-persist-01 challenge by prompting the user to create the required TXT record (if necessary)
// then performing propagation checks (or a wait-only delay) before notifying the ACME server.
func (c *Challenge) Solve(ctx context.Context, authz acme.Authorization) error {
	domain := authz.Identifier.Value
	if domain == "" {
		return errors.New("dnspersist01: empty identifier")
	}

	accountURI := c.core.GetKid()
	if accountURI == "" {
		return errors.New("dnspersist01: ACME account URI cannot be empty")
	}

	log.Info("dnspersist01: trying to solve the challenge.", log.DomainAttr(domain))

	chlng, err := challenge.FindChallenge(challenge.DNSPersist01, authz)
	if err != nil {
		return err
	}

	err = validateIssuerDomainNames(chlng)
	if err != nil {
		return fmt.Errorf("dnspersist01: %w", err)
	}

	fqdn := getAuthorizationDomainName(domain)

	result, err := DefaultClient().LookupTXT(ctx, fqdn)
	if err != nil {
		return fmt.Errorf("dnspersist01: %w", err)
	}

	issuerDomainName, err := c.selectIssuerDomainName(chlng.IssuerDomainNames, result.Records, accountURI, authz.Wildcard)
	if err != nil {
		return fmt.Errorf("dnspersist01: %w", err)
	}

	matcher := func(records []TXTRecord) bool {
		return c.hasMatchingRecord(records, issuerDomainName, accountURI, authz.Wildcard)
	}

	if !matcher(result.Records) {
		var info ChallengeInfo

		info, err = GetChallengeInfo(authz, issuerDomainName, accountURI, c.persistUntil)
		if err != nil {
			return err
		}

		err = c.provider.Persist(ctx, info.FQDN, info.Value)
		if err != nil {
			return err
		}
	} else {
		log.Info("dnspersist01: found existing matching TXT record for %s, no need to create a new one", log.DomainAttr(fqdn))
	}

	err = c.waitForPropagation(ctx, domain, fqdn, matcher)
	if err != nil {
		return err
	}

	return c.validate(ctx, c.core, domain, chlng)
}

func (c *Challenge) waitForPropagation(ctx context.Context, domain, fqdn string, matcher RecordMatcher) error {
	timeout, interval := c.provider.Timeout()

	log.Info("dnspersist01: waiting for record propagation.", log.DomainAttr(domain))

	time.Sleep(interval)

	return wait.For("propagation", timeout, interval, func() (bool, error) {
		stop, callErr := c.preCheck.call(ctx, domain, fqdn, matcher)
		if !stop || callErr != nil {
			log.Info("dnspersist01: waiting for record propagation.", log.DomainAttr(domain))
		}

		return stop, callErr
	})
}

// selectIssuerDomainName selects the issuer-domain-name to use for a dns-persist-01 challenge.
// If the user has supplied an issuer-domain-name,
// it is used after verifying that it is offered by the ACME challenge.
// Otherwise,
// the first issuer-domain-name with a matching TXT record is selected.
// If no issuer-domain-name has a matching TXT record,
// a deterministic default issuer-domain-name is selected using lexicographic ordering.
func (c *Challenge) selectIssuerDomainName(challIssuers []string, records []TXTRecord, accountURI string, wildcard bool) (string, error) {
	if len(challIssuers) == 0 {
		return "", errors.New("issuer-domain-names missing from the challenge")
	}

	sortedIssuers := slices.Clone(challIssuers)
	sort.Strings(sortedIssuers)

	if c.userSuppliedIssuerDomainName != "" {
		if !slices.Contains(sortedIssuers, c.userSuppliedIssuerDomainName) {
			return "", fmt.Errorf("provided issuer-domain-name %q not offered by the challenge", c.userSuppliedIssuerDomainName)
		}

		return c.userSuppliedIssuerDomainName, nil
	}

	for _, issuerDomainName := range sortedIssuers {
		if c.hasMatchingRecord(records, issuerDomainName, accountURI, wildcard) {
			return issuerDomainName, nil
		}
	}

	return sortedIssuers[0], nil
}

func (c *Challenge) hasMatchingRecord(records []TXTRecord, issuerDomainName, accountURI string, wildcard bool) bool {
	iv := IssueValue{
		IssuerDomainName: issuerDomainName,
		AccountURI:       accountURI,
		PersistUntil:     c.persistUntil,
	}

	if wildcard {
		iv.Policy = policyWildcard
	}

	return slices.ContainsFunc(records, func(record TXTRecord) bool {
		parsed, err := parseIssueValue(record.Value)
		if err != nil {
			log.Debug("dnspersist01: failed to parse TXT record value", log.ErrorAttr(err))
			return false
		}

		return parsed.match(iv)
	})
}

// ChallengeInfo contains the information used to create a dns-persist-01 TXT record.
type ChallengeInfo struct {
	// FQDN is the full-qualified challenge domain (i.e. `_validation-persist.[domain].`).
	FQDN string

	// Value contains the TXT record value, an RFC 8659 issue-value.
	Value string

	// IssuerDomainName is the normalized issuer-domain-name used in Value.
	IssuerDomainName string
}

// GetChallengeInfo returns information used to create a DNS TXT record
// which can fulfill the `dns-persist-01` challenge.
// Domain, issuerDomainName, and accountURI parameters are required.
// Wildcard and persistUntil parameters are optional.
func GetChallengeInfo(authz acme.Authorization, issuerDomainName, accountURI string, persistUntil time.Time) (ChallengeInfo, error) {
	if authz.Identifier.Value == "" {
		return ChallengeInfo{}, errors.New("dnspersist01: domain cannot be empty")
	}

	value, err := buildIssueValue(issuerDomainName, accountURI, authz.Wildcard, persistUntil)
	if err != nil {
		return ChallengeInfo{}, fmt.Errorf("dnspersist01: %w", err)
	}

	return ChallengeInfo{
		FQDN:             getAuthorizationDomainName(authz.Identifier.Value),
		Value:            value,
		IssuerDomainName: issuerDomainName,
	}, nil
}

// getAuthorizationDomainName returns the fully qualified DNS label
// used by the dns-persist-01 challenge for the given domain.
func getAuthorizationDomainName(domain string) string {
	return dns.Fqdn(validationLabel + "." + domain)
}
