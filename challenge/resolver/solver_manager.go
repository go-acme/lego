package resolver

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"time"

	"github.com/cenkalti/backoff/v5"
	"github.com/go-acme/lego/v5/acme"
	"github.com/go-acme/lego/v5/acme/api"
	"github.com/go-acme/lego/v5/challenge"
	"github.com/go-acme/lego/v5/challenge/dns01"
	"github.com/go-acme/lego/v5/challenge/dnspersist01"
	"github.com/go-acme/lego/v5/challenge/http01"
	"github.com/go-acme/lego/v5/challenge/tlsalpn01"
	"github.com/go-acme/lego/v5/log"
	"github.com/go-acme/lego/v5/platform/wait"
)

type byType []acme.Challenge

func (a byType) Len() int      { return len(a) }
func (a byType) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a byType) Less(i, j int) bool {
	// When users configure both DNS and DNS-PERSIST-01, prefer DNS-01 to avoid
	// unexpectedly selecting the manual-only DNS-PERSIST-01 workflow.
	if a[i].Type == string(challenge.DNS01) && a[j].Type == string(challenge.DNSPersist01) {
		return true
	}
	if a[i].Type == string(challenge.DNSPersist01) && a[j].Type == string(challenge.DNS01) {
		return false
	}
	return a[i].Type > a[j].Type
}

type SolverManager struct {
	core    *api.Core
	solvers map[challenge.Type]solver
}

func NewSolversManager(core *api.Core) *SolverManager {
	return &SolverManager{
		solvers: map[challenge.Type]solver{},
		core:    core,
	}
}

// SetHTTP01Provider specifies a custom provider p that can solve the given HTTP-01 challenge.
func (c *SolverManager) SetHTTP01Provider(p challenge.Provider, opts ...http01.ChallengeOption) error {
	c.solvers[challenge.HTTP01] = http01.NewChallenge(c.core, validate, p, opts...)
	return nil
}

// SetTLSALPN01Provider specifies a custom provider p that can solve the given TLS-ALPN-01 challenge.
func (c *SolverManager) SetTLSALPN01Provider(p challenge.Provider, opts ...tlsalpn01.ChallengeOption) error {
	c.solvers[challenge.TLSALPN01] = tlsalpn01.NewChallenge(c.core, validate, p, opts...)
	return nil
}

// SetDNS01Provider specifies a custom provider p that can solve the given DNS-01 challenge.
func (c *SolverManager) SetDNS01Provider(p challenge.Provider, opts ...dns01.ChallengeOption) error {
	c.solvers[challenge.DNS01] = dns01.NewChallenge(c.core, validate, p, opts...)
	return nil
}

// SetDNSPersist01 configures the dns-persist-01 challenge solver.
func (c *SolverManager) SetDNSPersist01(opts ...dnspersist01.ChallengeOption) error {
	c.solvers[challenge.DNSPersist01] = dnspersist01.NewChallenge(c.core, validate, opts...)
	return nil
}

// Remove removes a challenge type from the available solvers.
func (c *SolverManager) Remove(chlgType challenge.Type) {
	delete(c.solvers, chlgType)
}

// Checks all challenges from the server in order and returns the first matching solver.
func (c *SolverManager) chooseSolver(authz acme.Authorization) solver {
	// Allow to have a deterministic challenge order
	sort.Sort(byType(authz.Challenges))

	domain := challenge.GetTargetedDomain(authz)
	for _, chlg := range authz.Challenges {
		if solvr, ok := c.solvers[challenge.Type(chlg.Type)]; ok {
			log.Info("acme: use solver.", log.DomainAttr(domain), slog.String("type", chlg.Type))
			return solvr
		}

		log.Info("acme: Could not find the solver.", log.DomainAttr(domain), slog.String("type", chlg.Type))
	}

	return nil
}

func validate(ctx context.Context, core *api.Core, domain string, chlg acme.Challenge) error {
	chlng, err := core.Challenges.New(ctx, chlg.URL)
	if err != nil {
		return fmt.Errorf("failed to initiate challenge: %w", err)
	}

	valid, err := checkChallengeStatus(chlng)
	if err != nil {
		return err
	}

	if valid {
		log.Info("The server validated our request.", log.DomainAttr(domain))
		return nil
	}

	retryAfter, err := api.ParseRetryAfter(chlng.RetryAfter)
	if err != nil || retryAfter == 0 {
		// The ACME server MUST return a Retry-After.
		// If it doesn't, or if it's invalid, we'll just poll hard.
		// Boulder does not implement the ability to retry challenges or the Retry-After header.
		// https://github.com/letsencrypt/boulder/blob/master/docs/acme-divergences.md#section-82
		retryAfter = 5 * time.Second
	}

	bo := backoff.NewExponentialBackOff()
	bo.InitialInterval = retryAfter
	bo.MaxInterval = 10 * retryAfter

	// After the path is sent, the ACME server will access our server.
	// Repeatedly check the server for an updated status on our request.
	operation := func() error {
		authz, err := core.Authorizations.Get(ctx, chlng.AuthorizationURL)
		if err != nil {
			return backoff.Permanent(err)
		}

		valid, err := checkAuthorizationStatus(authz)
		if err != nil {
			return backoff.Permanent(err)
		}

		if valid {
			log.Info("The server validated our request.", log.DomainAttr(domain))
			return nil
		}

		return fmt.Errorf("the server didn't respond to our request (status=%s)", authz.Status)
	}

	return wait.Retry(ctx, operation,
		backoff.WithBackOff(bo),
		backoff.WithMaxElapsedTime(100*retryAfter))
}

func checkChallengeStatus(chlng acme.ExtendedChallenge) (bool, error) {
	switch chlng.Status {
	case acme.StatusValid:
		return true, nil
	case acme.StatusPending, acme.StatusProcessing:
		return false, nil
	case acme.StatusInvalid:
		return false, fmt.Errorf("invalid challenge: %w", chlng.Err())
	default:
		return false, fmt.Errorf("the server returned an unexpected challenge status: %s", chlng.Status)
	}
}

func checkAuthorizationStatus(authz acme.Authorization) (bool, error) {
	switch authz.Status {
	case acme.StatusValid:
		return true, nil
	case acme.StatusPending, acme.StatusProcessing:
		return false, nil
	case acme.StatusDeactivated, acme.StatusExpired, acme.StatusRevoked:
		return false, fmt.Errorf("the authorization state %s", authz.Status)
	case acme.StatusInvalid:
		for _, chlg := range authz.Challenges {
			if chlg.Status == acme.StatusInvalid && chlg.Error != nil {
				return false, fmt.Errorf("invalid authorization: %w", chlg.Err())
			}
		}

		return false, errors.New("invalid authorization")
	default:
		return false, fmt.Errorf("the server returned an unexpected authorization status: %s", authz.Status)
	}
}
