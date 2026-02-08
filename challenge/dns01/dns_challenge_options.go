package dns01

import (
	"context"
	"log/slog"
	"time"

	"github.com/go-acme/lego/v5/log"
)

type ChallengeOption func(*Challenge) error

// CondOption Conditional challenge option.
func CondOption(condition bool, opt ChallengeOption) ChallengeOption {
	if !condition {
		// NoOp options
		return func(*Challenge) error {
			return nil
		}
	}

	return opt
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
