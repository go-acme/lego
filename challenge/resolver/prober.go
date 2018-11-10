package resolver

import (
	"fmt"

	"github.com/xenolf/lego/challenge"
	"github.com/xenolf/lego/le"
	"github.com/xenolf/lego/log"
)

// Interface for all challenge solvers to implement.
type solver interface {
	Solve(authorization le.Authorization) error
}

// Interface for challenges like dns, where we can set a record in advance for ALL challenges.
// This saves quite a bit of time vs creating the records and solving them serially.
type preSolver interface {
	PreSolve(authorization le.Authorization) error
}

// Interface for challenges like dns, where we can solve all the challenges before to delete them.
type cleanup interface {
	CleanUp(authorization le.Authorization) error
}

// an authz with the solver we have chosen and the index of the challenge associated with it
type selectedAuthSolver struct {
	authz  le.Authorization
	solver solver
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
func (p *Prober) Solve(authorizations []le.Authorization) error {
	failures := make(obtainError)

	var authSolvers []*selectedAuthSolver

	// Loop through the resources, basically through the domains.
	// First pass just selects a solver for each authz.
	for _, authz := range authorizations {
		domain := challenge.GetTargetedDomain(authz)
		if authz.Status == le.StatusValid {
			// Boulder might recycle recent validated authz (see issue #267)
			log.Infof("[%s] acme: authorization already valid; skipping challenge", domain)
			continue
		}

		if solvr := p.solverManager.chooseSolver(authz); solvr != nil {
			authSolvers = append(authSolvers, &selectedAuthSolver{
				authz:  authz,
				solver: solvr,
			})
		} else {
			failures[domain] = fmt.Errorf("[%s] acme: could not determine solvers", domain)
		}
	}

	// For all valid preSolvers, first submit the challenges so they have max time to propagate
	for _, authSolver := range authSolvers {
		authz := authSolver.authz
		if solvr, ok := authSolver.solver.(preSolver); ok {
			err := solvr.PreSolve(authz)
			if err != nil {
				failures[challenge.GetTargetedDomain(authz)] = err
			}
		}
	}

	defer func() {
		// Clean all created TXT records
		for _, authSolver := range authSolvers {
			if solvr, ok := authSolver.solver.(cleanup); ok {
				domain := challenge.GetTargetedDomain(authSolver.authz)
				err := solvr.CleanUp(authSolver.authz)
				if err != nil {
					log.Warnf("[%s] acme: error cleaning up: %v ", domain, err)
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

		err := authSolver.solver.Solve(authz)
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
