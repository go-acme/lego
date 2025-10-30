package resolver

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/cenkalti/backoff/v5"
	"github.com/go-acme/lego/v4/acme"
	"github.com/go-acme/lego/v4/acme/api"
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/challenge/http01"
	"github.com/go-acme/lego/v4/challenge/tlsalpn01"
	"github.com/go-acme/lego/v4/log"
	"github.com/go-acme/lego/v4/platform/wait"
)

type byType []acme.Challenge

func (a byType) Len() int           { return len(a) }
func (a byType) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byType) Less(i, j int) bool { return a[i].Type > a[j].Type }

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
			log.Infof("[%s] acme: use %s solver", domain, chlg.Type)
			return solvr
		}

		log.Infof("[%s] acme: Could not find solver for: %s", domain, chlg.Type)
	}

	return nil
}

func validate(core *api.Core, domain string, chlg acme.Challenge) error {
	chlng, err := core.Challenges.New(chlg.URL)
	if err != nil {
		return fmt.Errorf("failed to initiate challenge: %w", err)
	}

	valid, err := checkChallengeStatus(chlng)
	if err != nil {
		return err
	}

	if valid {
		log.Infof("[%s] The server validated our request", domain)
		return nil
	}

	ra, err := strconv.Atoi(chlng.RetryAfter)
	if err != nil {
		// The ACME server MUST return a Retry-After.
		// If it doesn't, we'll just poll hard.
		// Boulder does not implement the ability to retry challenges or the Retry-After header.
		// https://github.com/letsencrypt/boulder/blob/master/docs/acme-divergences.md#section-82
		ra = 5
	}

	initialInterval := time.Duration(ra) * time.Second

	ctx := context.Background()

	bo := backoff.NewExponentialBackOff()
	bo.InitialInterval = initialInterval
	bo.MaxInterval = 10 * initialInterval

	// After the path is sent, the ACME server will access our server.
	// Repeatedly check the server for an updated status on our request.
	operation := func() error {
		authz, err := core.Authorizations.Get(chlng.AuthorizationURL)
		if err != nil {
			return backoff.Permanent(err)
		}

		valid, err := checkAuthorizationStatus(authz)
		if err != nil {
			return backoff.Permanent(err)
		}

		if valid {
			log.Infof("[%s] The server validated our request", domain)
			return nil
		}

		return fmt.Errorf("the server didn't respond to our request (status=%s)", authz.Status)
	}

	return wait.Retry(ctx, operation,
		backoff.WithBackOff(bo),
		backoff.WithMaxElapsedTime(100*initialInterval))
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
