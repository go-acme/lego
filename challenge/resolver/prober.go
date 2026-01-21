package resolver

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-acme/lego/v5/acme"
	"github.com/go-acme/lego/v5/challenge"
	"github.com/go-acme/lego/v5/log"
)

// Interface for all challenge solvers to implement.
type solver interface {
	Solve(ctx context.Context, authorization acme.Authorization) error
}

// Interface for challenges like dns, where we can set a record in advance for ALL challenges.
// This saves quite a bit of time vs creating the records and solving them serially.
type preSolver interface {
	PreSolve(ctx context.Context, authorization acme.Authorization) error
}

// Interface for challenges like dns, where we can solve all the challenges before to delete them.
type cleanup interface {
	CleanUp(ctx context.Context, authorization acme.Authorization) error
}

type sequential interface {
	Sequential() (bool, time.Duration)
}

// an authz with the solver we have chosen and the index of the challenge associated with it.
type selectedAuthSolver struct {
	authz  acme.Authorization
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
func (p *Prober) Solve(ctx context.Context, authorizations []acme.Authorization) error {
	failures := make(obtainError)

	var (
		authSolvers           []*selectedAuthSolver
		authSolversSequential []*selectedAuthSolver
	)

	// Loop through the resources, basically through the domains.
	// First pass just selects a solver for each authz.

	for _, authz := range authorizations {
		domain := challenge.GetTargetedDomain(authz)
		if authz.Status == acme.StatusValid {
			// Boulder might recycle recent validated authz (see issue #267)
			log.Info("acme: authorization already valid; skipping challenge.", log.DomainAttr(domain))
			continue
		}

		if solvr := p.solverManager.chooseSolver(authz); solvr != nil {
			authSolver := &selectedAuthSolver{authz: authz, solver: solvr}

			switch s := solvr.(type) {
			case sequential:
				if ok, _ := s.Sequential(); ok {
					authSolversSequential = append(authSolversSequential, authSolver)
				} else {
					authSolvers = append(authSolvers, authSolver)
				}
			default:
				authSolvers = append(authSolvers, authSolver)
			}
		} else {
			failures[domain] = fmt.Errorf("[%s] acme: could not determine solvers", domain)
		}
	}

	parallelSolve(ctx, authSolvers, failures)

	sequentialSolve(ctx, authSolversSequential, failures)

	// Be careful not to return an empty failures map,
	// for even an empty obtainError is a non-nil error value
	if len(failures) > 0 {
		return failures
	}

	return nil
}

func sequentialSolve(ctx context.Context, authSolvers []*selectedAuthSolver, failures obtainError) {
	for i, authSolver := range authSolvers {
		// Submit the challenge
		domain := challenge.GetTargetedDomain(authSolver.authz)

		if solvr, ok := authSolver.solver.(preSolver); ok {
			err := solvr.PreSolve(ctx, authSolver.authz)
			if err != nil {
				failures[domain] = err

				cleanUp(ctx, authSolver.solver, authSolver.authz)

				continue
			}
		}

		// Solve the challenge
		err := authSolver.solver.Solve(ctx, authSolver.authz)
		if err != nil {
			failures[domain] = err

			cleanUp(ctx, authSolver.solver, authSolver.authz)

			continue
		}

		// Clean challenge
		cleanUp(ctx, authSolver.solver, authSolver.authz)

		if len(authSolvers)-1 > i {
			solvr := authSolver.solver.(sequential)
			_, interval := solvr.Sequential()
			log.Info("sequence: wait.", slog.Duration("interval", interval))
			time.Sleep(interval)
		}
	}
}

func parallelSolve(ctx context.Context, authSolvers []*selectedAuthSolver, failures obtainError) {
	// For all valid preSolvers, first submit the challenges, so they have max time to propagate
	for _, authSolver := range authSolvers {
		authz := authSolver.authz
		if solvr, ok := authSolver.solver.(preSolver); ok {
			err := solvr.PreSolve(ctx, authz)
			if err != nil {
				failures[challenge.GetTargetedDomain(authz)] = err
			}
		}
	}

	defer func() {
		// Clean all created TXT records
		for _, authSolver := range authSolvers {
			cleanUp(ctx, authSolver.solver, authSolver.authz)
		}
	}()

	// Finally solve all challenges for real
	for _, authSolver := range authSolvers {
		authz := authSolver.authz

		domain := challenge.GetTargetedDomain(authz)
		if failures[domain] != nil {
			// already failed in previous loop
			continue
		}

		err := authSolver.solver.Solve(ctx, authz)
		if err != nil {
			failures[domain] = err
		}
	}
}

func cleanUp(ctx context.Context, solvr solver, authz acme.Authorization) {
	if solvr, ok := solvr.(cleanup); ok {
		domain := challenge.GetTargetedDomain(authz)

		err := solvr.CleanUp(ctx, authz)
		if err != nil {
			log.Warn("acme: cleaning up failed.", log.DomainAttr(domain), log.ErrorAttr(err))
		}
	}
}
