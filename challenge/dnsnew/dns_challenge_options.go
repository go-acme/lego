package dnsnew

import (
	"context"
	"time"
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

func RecursiveNSsPropagationRequirement() ChallengeOption {
	return func(chlg *Challenge) error {
		chlg.preCheck.requireRecursiveNssPropagation = true
		return nil
	}
}

func PropagationWait(wait time.Duration, skipCheck bool) ChallengeOption {
	return WrapPreCheck(func(ctx context.Context, domain, fqdn, value string, check PreCheckFunc) (bool, error) {
		time.Sleep(wait)

		if skipCheck {
			return true, nil
		}

		return check(ctx, fqdn, value)
	})
}
