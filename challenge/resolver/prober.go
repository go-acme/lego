package resolver

import (
	"fmt"

	"github.com/xenolf/lego/le"
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

type Prober struct {
	solverManager *SolverManager
}

func NewProber(solverManager *SolverManager) *Prober {
	return &Prober{
		solverManager: solverManager,
	}
}

// Solve Looks through the challenge combinations to find a solvable match.
// Then solves the challenges in series and returns.
func (c *Prober) Solve(authorizations []le.Authorization) error {
	failures := make(obtainError)

	var authSolvers []*selectedAuthSolver

	// Loop through the resources, basically through the domains.
	// First pass just selects a solver for each authz.
	for _, authz := range authorizations {
		if authz.Status == le.StatusValid {
			// Boulder might recycle recent validated authz (see issue #267)
			log.Infof("[%s] acme: authorization already valid; skipping challenge", authz.Identifier.Value)
			continue
		}

		if i, solvr := c.solverManager.chooseSolver(authz); solvr != nil {
			authSolvers = append(authSolvers, &selectedAuthSolver{
				authz:          authz,
				challengeIndex: i,
				solver:         solvr,
			})
		} else {
			failures[authz.Identifier.Value] = fmt.Errorf("[%s] acme: could not determine solvers", authz.Identifier.Value)
		}
	}

	// For all valid presolvers, first submit the challenges so they have max time to propagate
	for _, authSolver := range authSolvers {
		authz := authSolver.authz
		if solvr, ok := authSolver.solver.(preSolver); ok {
			err := solvr.PreSolve(authz.Challenges[authSolver.challengeIndex], authz.Identifier.Value)
			if err != nil {
				failures[authz.Identifier.Value] = err
			}
		}
	}

	defer func() {
		// Clean all created TXT records
		for _, authSolver := range authSolvers {
			if solvr, ok := authSolver.solver.(cleanup); ok {
				if failures[authSolver.authz.Identifier.Value] != nil {
					// already failed in previous loop
					continue
				}

				err := solvr.CleanUp(authSolver.authz.Challenges[authSolver.challengeIndex], authSolver.authz.Identifier.Value)
				if err != nil {
					log.Warnf("Error cleaning up %s: %v ", authSolver.authz.Identifier.Value, err)
				}
			}
		}
	}()

	// Finally solve all challenges for real
	for _, authSolver := range authSolvers {
		authz := authSolver.authz
		if failures[authz.Identifier.Value] != nil {
			// already failed in previous loop
			continue
		}

		err := authSolver.solver.Solve(authz.Challenges[authSolver.challengeIndex], authz.Identifier.Value)
		if err != nil {
			failures[authz.Identifier.Value] = err
		}
	}

	// Be careful not to return an empty failures map,
	// for even an empty obtainError is a non-nil error value
	if len(failures) > 0 {
		return failures
	}
	return nil
}
