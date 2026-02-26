package dnspersist01

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/go-acme/lego/v5/log"
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

// WithAccountURI sets the ACME account URI bound to dns-persist-01 records.
// It is required both to construct the `accounturi=` parameter and
// to match already-provisioned TXT records that should be updated.
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
// When set, it overrides automatic issuer selection and
// must match one of the issuer-domain-names offered in the ACME challenge.
// User input is normalized and validated at configuration time.
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

// WithPersistUntil sets the optional persistUntil value used when constructing dns-persist-01 TXT records.
func WithPersistUntil(persistUntil time.Time) ChallengeOption {
	return func(chlg *Challenge) error {
		if persistUntil.IsZero() {
			return errors.New("persistUntil cannot be zero")
		}

		chlg.persistUntil = persistUntil.UTC().Truncate(time.Second)

		return nil
	}
}

// DisableAuthoritativeNssPropagationRequirement disables authoritative nameserver checks.
func DisableAuthoritativeNssPropagationRequirement() ChallengeOption {
	return func(chlg *Challenge) error {
		chlg.preCheck.requireAuthoritativeNssPropagation = false
		return nil
	}
}

// DisableRecursiveNSsPropagationRequirement disables recursive nameserver checks.
func DisableRecursiveNSsPropagationRequirement() ChallengeOption {
	return func(chlg *Challenge) error {
		chlg.preCheck.requireRecursiveNssPropagation = false
		return nil
	}
}

// PropagationWait sleeps for the specified duration, optionally skipping checks.
func PropagationWait(wait time.Duration, skipCheck bool) ChallengeOption {
	return WrapPreCheck(func(ctx context.Context, domain, fqdn string, matcher RecordMatcher, check PreCheckFunc) (bool, error) {
		if skipCheck {
			log.Info("acme: the active propagation check is disabled, waiting for the propagation instead.",
				slog.Duration("wait", wait),
				log.DomainAttr(domain),
			)
		}

		time.Sleep(wait)

		if skipCheck {
			return true, nil
		}

		return check(ctx, fqdn, matcher)
	})
}
