package dnspersist01

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"sort"
	"strings"
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

// ChallengeInfo contains the information used to create a dns-persist-01 TXT record.
type ChallengeInfo struct {
	// FQDN is the full-qualified challenge domain (i.e. `_validation-persist.[domain].`).
	FQDN string

	// Value contains the TXT record value, an RFC 8659 issue-value.
	Value string

	// IssuerDomainName is the normalized issuer-domain-name used in Value.
	IssuerDomainName string
}

// Challenge implements the dns-persist-01 challenge.
type Challenge struct {
	core     *api.Core
	validate ValidateFunc
	provider challenge.PersistentProvider
	preCheck preCheck

	accountURI                   string
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
			return nil, fmt.Errorf("dnspersist01: %w", err)
		}
	}

	if chlg.accountURI == "" {
		return nil, errors.New("dnspersist01: account URI cannot be empty")
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

	chlng, err := challenge.FindChallenge(challenge.DNSPersist01, authz)
	if err != nil {
		return err
	}

	err = validateIssuerDomainNames(chlng)
	if err != nil {
		return fmt.Errorf("dnspersist01: %w", err)
	}

	fqdn := GetAuthorizationDomainName(domain)

	result, err := DefaultClient().LookupTXT(ctx, fqdn)
	if err != nil {
		return err
	}

	issuerDomainName, err := c.selectIssuerDomainName(chlng.IssuerDomainNames, result.Records, authz.Wildcard)
	if err != nil {
		return fmt.Errorf("dnspersist01: %w", err)
	}

	matcher := func(records []TXTRecord) bool {
		return c.hasMatchingRecord(records, issuerDomainName, authz.Wildcard)
	}

	if !matcher(result.Records) {
		err = c.provider.Persist(ctx, authz, issuerDomainName, c.accountURI, c.persistUntil)
		if err != nil {
			return err
		}
	} else {
		fmt.Printf("dnspersist01: Found existing matching TXT record for %s, no need to create a new one\n", fqdn)
	}

	timeout, interval := c.provider.Timeout()

	log.Info("acme: waiting for DNS-PERSIST-01 record propagation.",
		log.DomainAttr(domain),
		slog.String("nameservers", strings.Join(DefaultClient().recursiveNameservers, ",")),
	)

	time.Sleep(interval)

	err = wait.For("propagation", timeout, interval, func() (bool, error) {
		ok, callErr := c.preCheck.call(ctx, domain, fqdn, matcher)
		if !ok || callErr != nil {
			log.Info("acme: Waiting for DNS-PERSIST-01 record propagation.", log.DomainAttr(domain))
		}

		return ok, callErr
	})
	if err != nil {
		return err
	}

	return c.validate(ctx, c.core, domain, chlng)
}

// selectIssuerDomainName selects the issuer-domain-name to use for a dns-persist-01 challenge.
// If the user has supplied an issuer-domain-name,
// it is used after verifying that it is offered by the ACME challenge.
// Otherwise,
// the first issuer-domain-name with a matching TXT record is selected.
// If no issuer-domain-name has a matching TXT record,
// a deterministic default issuer-domain-name is selected using lexicographic ordering.
func (c *Challenge) selectIssuerDomainName(challIssuers []string, records []TXTRecord, wildcard bool) (string, error) {
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
		if c.hasMatchingRecord(records, issuerDomainName, wildcard) {
			return issuerDomainName, nil
		}
	}

	return sortedIssuers[0], nil
}

func (c *Challenge) hasMatchingRecord(records []TXTRecord, issuerDomainName string, wildcard bool) bool {
	iv := IssueValue{
		IssuerDomainName: issuerDomainName,
		AccountURI:       c.accountURI,
		PersistUntil:     c.persistUntil,
	}

	if wildcard {
		iv.Policy = policyWildcard
	}

	return slices.ContainsFunc(records, func(record TXTRecord) bool {
		parsed, err := parseIssueValue(record.Value)
		if err != nil {
			return false
		}

		return parsed.match(iv)
	})
}

// GetAuthorizationDomainName returns the fully qualified DNS label
// used by the dns-persist-01 challenge for the given domain.
func GetAuthorizationDomainName(domain string) string {
	return dns.Fqdn(validationLabel + "." + domain)
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
		return ChallengeInfo{}, err
	}

	return ChallengeInfo{
		FQDN:             GetAuthorizationDomainName(authz.Identifier.Value),
		Value:            value,
		IssuerDomainName: issuerDomainName,
	}, nil
}
