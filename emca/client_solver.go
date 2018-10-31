package emca

import (
	"fmt"
	"net"

	"github.com/xenolf/lego/emca/challenge/tlsalpn01"

	"github.com/xenolf/lego/emca/challenge/dns01"
	"github.com/xenolf/lego/emca/challenge/http01"

	"github.com/xenolf/lego/emca/challenge"
	"github.com/xenolf/lego/emca/le"
	"github.com/xenolf/lego/log"
)

// Interface for all challenge solvers to implement.
type solver interface {
	Solve(challenge le.Challenge, domain string) error
}

// Interface for challenges like dns, where we can set a record in advance for ALL challenges.
// This saves quite a bit of time vs creating the records and solving them serially.
type preSolver interface {
	PreSolve(challenge le.Challenge, domain string) error
}

// Interface for challenges like dns, where we can solve all the challenges before to delete them.
type cleanup interface {
	CleanUp(challenge le.Challenge, domain string) error
}

// an authz with the solver we have chosen and the index of the challenge associated with it
type selectedAuthSolver struct {
	authz          le.Authorization
	challengeIndex int
	solver         solver
}

// SetHTTPAddress specifies a custom interface:port to be used for HTTP based challenges.
// If this option is not used, the default port 80 and all interfaces will be used.
// To only specify a port and no interface use the ":port" notation.
//
// NOTE: This REPLACES any custom HTTP provider previously set by calling
// c.SetChallengeProvider with the default HTTP challenge provider.
func (c *Client) SetHTTPAddress(iface string) error {
	host, port, err := net.SplitHostPort(iface)
	if err != nil {
		return err
	}

	if chlng, ok := c.solvers[challenge.HTTP01]; ok {
		chlng.(*http01.Challenge).SetProvider(http01.NewProviderServer(host, port))
	}

	return nil
}

// SetTLSAddress specifies a custom interface:port to be used for TLS based challenges.
// If this option is not used, the default port 443 and all interfaces will be used.
// To only specify a port and no interface use the ":port" notation.
//
// NOTE: This REPLACES any custom TLS-ALPN provider previously set by calling
// c.SetChallengeProvider with the default TLS-ALPN challenge provider.
func (c *Client) SetTLSAddress(iface string) error {
	host, port, err := net.SplitHostPort(iface)
	if err != nil {
		return err
	}

	if chlng, ok := c.solvers[challenge.TLSALPN01]; ok {
		chlng.(*tlsalpn01.Challenge).SetProvider(tlsalpn01.NewProviderServer(host, port))
	}

	return nil
}

// SetChallengeProvider specifies a custom provider p that can solve the given challenge type.
func (c *Client) SetChallengeProvider(chlg challenge.Type, p challenge.ChallengeProvider) error {
	switch chlg {
	case challenge.HTTP01:
		c.solvers[chlg] = http01.NewChallenge(c.jws, validate, p)
	case challenge.DNS01:
		c.solvers[chlg] = dns01.NewChallenge(c.jws, validate, p)
	case challenge.TLSALPN01:
		c.solvers[chlg] = tlsalpn01.NewChallenge(c.jws, validate, p)
	default:
		return fmt.Errorf("unknown challenge %v", chlg)
	}
	return nil
}

// ExcludeChallenges explicitly removes challenges from the pool for solving.
func (c *Client) ExcludeChallenges(challenges []challenge.Type) {
	// Loop through all challenges and delete the requested one if found.
	for _, chlg := range challenges {
		delete(c.solvers, chlg)
	}
}

// Looks through the challenge combinations to find a solvable match.
// Then solves the challenges in series and returns.
func (c *Client) solveChallengeForAuthz(authorizations []le.Authorization) error {
	failures := make(ObtainError)

	var authSolvers []*selectedAuthSolver

	// loop through the resources, basically through the domains. First pass just selects a solver for each authz.
	for _, authz := range authorizations {
		if authz.Status == statusValid {
			// Boulder might recycle recent validated authz (see issue #267)
			log.Infof("[%s] acme: Authorization already valid; skipping challenge", authz.Identifier.Value)
			continue
		}
		if i, solvr := c.chooseSolver(authz, authz.Identifier.Value); solvr != nil {
			authSolvers = append(authSolvers, &selectedAuthSolver{
				authz:          authz,
				challengeIndex: i,
				solver:         solvr,
			})
		} else {
			failures[authz.Identifier.Value] = fmt.Errorf("[%s] acme: Could not determine solvers", authz.Identifier.Value)
		}
	}

	// for all valid presolvers, first submit the challenges so they have max time to propagate
	for _, item := range authSolvers {
		authz := item.authz
		i := item.challengeIndex
		if presolver, ok := item.solver.(preSolver); ok {
			if err := presolver.PreSolve(authz.Challenges[i], authz.Identifier.Value); err != nil {
				failures[authz.Identifier.Value] = err
			}
		}
	}

	defer func() {
		// clean all created TXT records
		for _, item := range authSolvers {
			if clean, ok := item.solver.(cleanup); ok {
				if failures[item.authz.Identifier.Value] != nil {
					// already failed in previous loop
					continue
				}
				err := clean.CleanUp(item.authz.Challenges[item.challengeIndex], item.authz.Identifier.Value)
				if err != nil {
					log.Warnf("Error cleaning up %s: %v ", item.authz.Identifier.Value, err)
				}
			}
		}
	}()

	// finally solve all challenges for real
	for _, item := range authSolvers {
		authz := item.authz
		i := item.challengeIndex
		if failures[authz.Identifier.Value] != nil {
			// already failed in previous loop
			continue
		}
		if err := item.solver.Solve(authz.Challenges[i], authz.Identifier.Value); err != nil {
			failures[authz.Identifier.Value] = err
		}
	}

	// be careful not to return an empty failures map, for
	// even an empty ObtainError is a non-nil error value
	if len(failures) > 0 {
		return failures
	}
	return nil
}

// Checks all challenges from the server in order and returns the first matching solver.
func (c *Client) chooseSolver(auth le.Authorization, domain string) (int, solver) {
	for i, chlg := range auth.Challenges {
		if solver, ok := c.solvers[challenge.Type(chlg.Type)]; ok {
			return i, solver
		}
		log.Infof("[%s] acme: Could not find solver for: %s", domain, chlg.Type)
	}
	return 0, nil
}
