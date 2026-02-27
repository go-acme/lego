package dns01

import (
	"context"
	"log/slog"
	"time"

	"github.com/go-acme/lego/v5/challenge/internal"
	"github.com/go-acme/lego/v5/log"
)

// ChallengeOption configures the dns-01 challenge.
type ChallengeOption = internal.ChallengeOption[*Challenge]

// CondOptions Conditional challenge options.
func CondOptions(condition bool, opt ...ChallengeOption) ChallengeOption {
	return internal.CondOptions(condition, opt...)
}

// LazyCondOption Lazy conditional challenge option.
func LazyCondOption(condition bool, fn func() ChallengeOption) ChallengeOption {
	return internal.LazyCondOption(condition, fn)
}

// CombineOptions Combines multiple challenge options into one.
func CombineOptions(opts ...ChallengeOption) ChallengeOption {
	return internal.CombineOptions(opts...)
}

func DisableAuthoritativeNssPropagationRequirement() ChallengeOption {
	return func(chlg *Challenge) error {
		chlg.preCheck.requireAuthoritativeNssPropagation = false
		return nil
	}
}

func DisableRecursiveNSsPropagationRequirement() ChallengeOption {
	return func(chlg *Challenge) error {
		chlg.preCheck.requireRecursiveNssPropagation = false
		return nil
	}
}

func PropagationWait(wait time.Duration, skipCheck bool) ChallengeOption {
	return WrapPreCheck(func(ctx context.Context, domain, fqdn, value string, check PreCheckFunc) (bool, error) {
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

		return check(ctx, fqdn, value)
	})
}
