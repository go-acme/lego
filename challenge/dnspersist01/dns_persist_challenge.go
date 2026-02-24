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
	"github.com/go-acme/lego/v5/log"
	"github.com/go-acme/lego/v5/platform/wait"
	"github.com/miekg/dns"
)

const (
	validationLabel = "_validation-persist"

	// DefaultPropagationTimeout default propagation timeout.
	DefaultPropagationTimeout = 60 * time.Second
	// DefaultPollingInterval default polling interval.
	DefaultPollingInterval = 2 * time.Second
)

// ValidateFunc validates a challenge with the ACME server.
type ValidateFunc func(ctx context.Context, core *api.Core, domain string, chlng acme.Challenge) error

// ChallengeOption configures the dns-persist-01 challenge.
type ChallengeOption func(*Challenge) error

// ChallengeInfo contains the information used to create a dns-persist-01 TXT
// record.
type ChallengeInfo struct {
	// FQDN is the full-qualified challenge domain (i.e.
	// `_validation-persist.[domain].`).
	FQDN string

	// Value contains the TXT record issue-value.
	Value string

	// IssuerDomainName is the normalized issuer-domain-name used in Value.
	IssuerDomainName string
}

// Challenge implements the dns-persist-01 challenge exclusively with manual
// instructions for TXT record creation.
type Challenge struct {
	core     *api.Core
	validate ValidateFunc
	resolver *Resolver
	preCheck preCheck

	accountURI                   string
	userSuppliedIssuerDomainName string
	persistUntil                 *time.Time
	recursiveNameservers         []string
	authoritativeNSPort          string

	propagationTimeout  time.Duration
	propagationInterval time.Duration
}

// NewChallenge creates a dns-persist-01 challenge.
func NewChallenge(core *api.Core, validate ValidateFunc, opts ...ChallengeOption) *Challenge {
	chlg := &Challenge{
		core:                 core,
		validate:             validate,
		resolver:             NewResolver(nil),
		preCheck:             newPreCheck(),
		recursiveNameservers: DefaultNameservers(),
		authoritativeNSPort:  defaultAuthoritativeNSPort,

		propagationTimeout:  DefaultPropagationTimeout,
		propagationInterval: DefaultPollingInterval,
	}

	for _, opt := range opts {
		err := opt(chlg)
		if err != nil {
			log.Warn("Challenge option skipped.", log.ErrorAttr(err))
		}
	}

	if chlg.accountURI == "" {
		log.Fatal("Account is required for DNS-PERSIST-01 challenge type")
	}

	return chlg
}

// WithResolver overrides the resolver used for DNS lookups.
func WithResolver(resolver *Resolver) ChallengeOption {
	return func(chlg *Challenge) error {
		if resolver == nil {
			return errors.New("dnspersist01: resolver is nil")
		}

		chlg.resolver = resolver

		return nil
	}
}

// WithNameservers overrides resolver nameservers using the default timeout.
func WithNameservers(nameservers []string) ChallengeOption {
	return func(chlg *Challenge) error {
		chlg.resolver = NewResolver(nameservers)
		return nil
	}
}

// WithDNSTimeout overrides the default DNS resolver timeout.
func WithDNSTimeout(timeout time.Duration) ChallengeOption {
	return func(chlg *Challenge) error {
		if chlg.resolver == nil {
			chlg.resolver = NewResolver(nil)
		}

		chlg.resolver.Timeout = timeout

		return nil
	}
}

// WithAccountURI sets the ACME account URI bound to dns-persist-01 records. It
// is required both to construct the `accounturi=` parameter and to match
// already-provisioned TXT records that should be updated.
func WithAccountURI(accountURI string) ChallengeOption {
	return func(chlg *Challenge) error {
		if accountURI == "" {
			return errors.New("dnspersist01: ACME account URI cannot be empty")
		}

		chlg.accountURI = accountURI

		return nil
	}
}

// WithIssuerDomainName forces the issuer-domain-name used for dns-persist-01.
// When set, it overrides automatic issuer selection and must match one of the
// issuer-domain-names offered in the ACME challenge. User input is normalized
// and validated at configuration time.
func WithIssuerDomainName(issuerDomainName string) ChallengeOption {
	return func(chlg *Challenge) error {
		normalized, err := normalizeUserSuppliedIssuerDomainName(issuerDomainName)
		if err != nil {
			return err
		}

		err = validateIssuerDomainName(normalized)
		if err != nil {
			return fmt.Errorf("dnspersist01: %w", err)
		}

		chlg.userSuppliedIssuerDomainName = normalized

		return nil
	}
}

// WithPersistUntil sets the optional persistUntil value used when constructing
// dns-persist-01 TXT records.
func WithPersistUntil(persistUntil time.Time) ChallengeOption {
	return func(chlg *Challenge) error {
		if persistUntil.IsZero() {
			return errors.New("dnspersist01: persistUntil cannot be zero")
		}

		ts := persistUntil.UTC().Truncate(time.Second)
		chlg.persistUntil = &ts

		return nil
	}
}

// WithPropagationTimeout overrides the propagation timeout duration.
func WithPropagationTimeout(timeout time.Duration) ChallengeOption {
	return func(chlg *Challenge) error {
		if timeout <= 0 {
			return errors.New("dnspersist01: propagation timeout must be positive")
		}

		chlg.propagationTimeout = timeout

		return nil
	}
}

// WithPropagationInterval overrides the propagation polling interval.
func WithPropagationInterval(interval time.Duration) ChallengeOption {
	return func(chlg *Challenge) error {
		if interval <= 0 {
			return errors.New("dnspersist01: propagation interval must be positive")
		}

		chlg.propagationInterval = interval

		return nil
	}
}

// Solve validates the dns-persist-01 challenge by prompting the user to create
// the required TXT record (if necessary) then performing propagation checks (or
// a wait-only delay) before notifying the ACME server.
//
//nolint:gocyclo // challenge flow has several required branches (reuse/manual/wait/propagation/validate).
func (c *Challenge) Solve(ctx context.Context, authz acme.Authorization) error {
	if c.resolver == nil {
		return errors.New("dnspersist01: resolver is nil")
	}

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
		return err
	}

	fqdn := GetAuthorizationDomainName(domain)

	result, err := c.resolver.LookupTXT(fqdn)
	if err != nil {
		return err
	}

	issuerDomainName, err := c.selectIssuerDomainName(chlng.IssuerDomainNames, result.Records, authz.Wildcard)
	if err != nil {
		return err
	}

	matcher := func(records []TXTRecord) bool {
		return c.hasMatchingRecord(records, issuerDomainName, authz.Wildcard)
	}

	if !matcher(result.Records) {
		info, infoErr := GetChallengeInfo(domain, issuerDomainName, c.accountURI, authz.Wildcard, c.persistUntil)
		if infoErr != nil {
			return infoErr
		}

		displayRecordCreationInstructions(info.FQDN, info.Value)

		waitErr := waitForUser()
		if waitErr != nil {
			return waitErr
		}
	} else {
		fmt.Printf("dnspersist01: Found existing matching TXT record for %s, no need to create a new one\n", fqdn)
	}

	timeout := c.propagationTimeout
	interval := c.propagationInterval

	log.Info("acme: Checking DNS-PERSIST-01 record propagation.",
		log.DomainAttr(domain), slog.String("nameservers", strings.Join(c.getRecursiveNameservers(), ",")),
	)

	time.Sleep(interval)

	err = wait.For("propagation", timeout, interval, func() (bool, error) {
		ok, callErr := c.preCheck.call(domain, fqdn, matcher, c.checkDNSPropagation)
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

func (c *Challenge) getRecursiveNameservers() []string {
	if c == nil || len(c.recursiveNameservers) == 0 {
		return DefaultNameservers()
	}

	return append([]string(nil), c.recursiveNameservers...)
}

// GetAuthorizationDomainName returns the fully-qualified DNS label used by the
// dns-persist-01 challenge for the given domain.
func GetAuthorizationDomainName(domain string) string {
	return dns.Fqdn(validationLabel + "." + domain)
}

// GetChallengeInfo returns information used to create a DNS TXT record which
// can fulfill the `dns-persist-01` challenge for a selected issuer. Domain,
// issuerDomainName, and accountURI parameters are required. Wildcard and
// persistUntil parameters are optional.
func GetChallengeInfo(domain, issuerDomainName, accountURI string, wildcard bool, persistUntil *time.Time) (ChallengeInfo, error) {
	if domain == "" {
		return ChallengeInfo{}, errors.New("dnspersist01: domain cannot be empty")
	}

	if accountURI == "" {
		return ChallengeInfo{}, errors.New("dnspersist01: ACME account URI cannot be empty")
	}

	err := validateIssuerDomainName(issuerDomainName)
	if err != nil {
		return ChallengeInfo{}, fmt.Errorf("dnspersist01: %w", err)
	}

	return ChallengeInfo{
		FQDN:             GetAuthorizationDomainName(domain),
		Value:            BuildIssueValues(issuerDomainName, accountURI, wildcard, persistUntil),
		IssuerDomainName: issuerDomainName,
	}, nil
}

// validateIssuerDomainNames validates the ACME challenge "issuer-domain-names"
// array for dns-persist-01.
//
// Rules enforced:
//   - The array is required and must contain at least 1 entry.
//   - The array must not contain more than 10 entries; larger arrays are
//     treated as malformed challenges and rejected.
//
// Each issuer-domain-name must be a normalized domain name:
//   - represented in A-label (Punycode, RFC5890) form
//   - all lowercase
//   - no trailing dot
//   - maximum total length of 253 octets
//
// The returned list is intended for issuer selection when constructing or
// matching dns-persist-01 TXT records. The challenge can be satisfied by using
// any one valid issuer-domain-name from this list.
func validateIssuerDomainNames(chlng acme.Challenge) error {
	if len(chlng.IssuerDomainNames) == 0 {
		return errors.New("dnspersist01: issuer-domain-names missing from challenge")
	}

	if len(chlng.IssuerDomainNames) > 10 {
		return errors.New("dnspersist01: issuer-domain-names exceeds maximum length of 10")
	}

	for _, issuerDomainName := range chlng.IssuerDomainNames {
		err := validateIssuerDomainName(issuerDomainName)
		if err != nil {
			return fmt.Errorf("dnspersist01: %w", err)
		}
	}

	return nil
}

// selectIssuerDomainName selects the issuer-domain-name to use for a
// dns-persist-01 challenge. If the user has supplied an issuer-domain-name, it
// is used after verifying that it is offered by the ACME challenge. Otherwise,
// the first issuer-domain-name with a matching TXT record is selected. If no
// issuer-domain-name has a matching TXT record, a deterministic default
// issuer-domain-name is selected using lexicographic ordering.
func (c *Challenge) selectIssuerDomainName(challIssuers []string, records []TXTRecord, wildcard bool) (string, error) {
	if len(challIssuers) == 0 {
		return "", errors.New("dnspersist01: issuer-domain-names missing from challenge")
	}

	sortedIssuers := slices.Clone(challIssuers)
	sort.Strings(sortedIssuers)

	if c.userSuppliedIssuerDomainName != "" {
		if !slices.Contains(sortedIssuers, c.userSuppliedIssuerDomainName) {
			return "", fmt.Errorf("dnspersist01: provided issuer-domain-name %q not offered by challenge", c.userSuppliedIssuerDomainName)
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
	for _, record := range records {
		parsed, err := ParseIssueValues(record.Value)
		if err != nil {
			continue
		}

		if parsed.IssuerDomainName != issuerDomainName {
			continue
		}

		if parsed.AccountURI != c.accountURI {
			continue
		}

		if wildcard && !strings.EqualFold(parsed.Policy, policyWildcard) {
			continue
		}

		if c.persistUntil == nil {
			if parsed.PersistUntil != nil {
				continue
			}
		} else {
			if parsed.PersistUntil == nil || !parsed.PersistUntil.Equal(*c.persistUntil) {
				continue
			}
		}

		return true
	}

	return false
}
