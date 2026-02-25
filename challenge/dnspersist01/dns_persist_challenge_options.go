package dnspersist01

import (
	"errors"
	"time"
)

// ChallengeOption configures the dns-persist-01 challenge.
type ChallengeOption func(*Challenge) error

// CondOptions Conditional challenge options.
func CondOptions(condition bool, opt ...ChallengeOption) ChallengeOption {
	if !condition {
		// NoOp options
		return func(*Challenge) error {
			return nil
		}
	}

	return func(chlg *Challenge) error {
		for _, opt := range opt {
			err := opt(chlg)
			if err != nil {
				return err
			}
		}

		return nil
	}
}

// WithResolver overrides the resolver used for DNS lookups.
func WithResolver(resolver *Resolver) ChallengeOption {
	return func(chlg *Challenge) error {
		if resolver == nil {
			return errors.New("resolver is nil")
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
			return errors.New("ACME account URI cannot be empty")
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
		if issuerDomainName == "" {
			return nil
		}

		normalized, err := normalizeUserSuppliedIssuerDomainName(issuerDomainName)
		if err != nil {
			return err
		}

		err = validateIssuerDomainName(normalized)
		if err != nil {
			return err
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
			return errors.New("persistUntil cannot be zero")
		}

		chlg.persistUntil = persistUntil.UTC().Truncate(time.Second)

		return nil
	}
}

// WithPropagationTimeout overrides the propagation timeout duration.
func WithPropagationTimeout(timeout time.Duration) ChallengeOption {
	return func(chlg *Challenge) error {
		if timeout <= 0 {
			return errors.New("propagation timeout must be positive")
		}

		chlg.propagationTimeout = timeout

		return nil
	}
}

// WithPropagationInterval overrides the propagation polling interval.
func WithPropagationInterval(interval time.Duration) ChallengeOption {
	return func(chlg *Challenge) error {
		if interval <= 0 {
			return errors.New("propagation interval must be positive")
		}

		chlg.propagationInterval = interval

		return nil
	}
}
