package resolver

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/go-acme/lego/v5/acme"
	"github.com/go-acme/lego/v5/challenge"
	"github.com/go-acme/lego/v5/internal/errutils"
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
	failures := errutils.NewDomainsError("resolver")

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
			log.Info("Authorization is already valid; skipping the challenge.", log.DomainAttr(domain))
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
			failures.Add(domain, errors.New("prober: could not determine solvers"))
		}
	}

	parallelSolve(ctx, authSolvers, failures)

	sequentialSolve(ctx, authSolversSequential, failures)

	return failures.Join()
}

func sequentialSolve(ctx context.Context, authSolvers []*selectedAuthSolver, failures *errutils.DomainsError) {
	// Some CA are using the same token,
	// this can be a problem with the DNS01 challenge when the DNS provider doesn't support duplicate TXT records.
	// In the sequential mode, this is not a problem because we can solve the challenges in order.
	// But it can reduce the number of call the DNS provider APIs.
	uniq := make(map[string]struct{})

	for i, authSolver := range authSolvers {
		// Submit the challenge
		domain := challenge.GetTargetedDomain(authSolver.authz)

		chlg, _ := challenge.FindChallenge(challenge.DNS01, authSolver.authz)

		if solvr, ok := authSolver.solver.(preSolver); ok {
			if _, ok := uniq[authSolver.authz.Identifier.Value+chlg.Token]; ok && chlg.Token != "" {
				log.Debug("Duplicate token; skipping pre-solve.",
					slog.String("identifier", authSolver.authz.Identifier.Value),
					slog.String("type", challenge.DNS01.String()),
				)

				continue
			}

			err := solvr.PreSolve(ctx, authSolver.authz)
			if err != nil {
				failures.Add(domain, err)

				cleanUp(ctx, authSolver.solver, authSolver.authz)

				continue
			}

			uniq[authSolver.authz.Identifier.Value+chlg.Token] = struct{}{}
		}

		// Solve the challenge
		err := authSolver.solver.Solve(ctx, authSolver.authz)
		if err != nil {
			failures.Add(domain, err)

			cleanUp(ctx, authSolver.solver, authSolver.authz)

			continue
		}

		if _, ok := uniq[authSolver.authz.Identifier.Value+chlg.Token]; ok || chlg.Token == "" {
			// Clean challenge
			cleanUp(ctx, authSolver.solver, authSolver.authz)

			if len(authSolvers)-1 > i {
				solvr := authSolver.solver.(sequential)
				_, interval := solvr.Sequential()
				log.Info("sequence: wait.", slog.Duration("interval", interval), log.DomainAttr(domain))
				time.Sleep(interval)
			}

			delete(uniq, authSolver.authz.Identifier.Value+chlg.Token)
		} else {
			log.Debug("Duplicate token; skipping cleanup.",
				slog.String("identifier", authSolver.authz.Identifier.Value),
				slog.String("type", challenge.DNS01.String()),
			)
		}
	}
}

func parallelSolve(ctx context.Context, authSolvers []*selectedAuthSolver, failures *errutils.DomainsError) {
	// Some CA are using the same token,
	// this can be a problem with the DNS01 challenge when the DNS provider doesn't support duplicate TXT records.
	uniq := make(map[string]struct{})

	// For all valid preSolvers, first submit the challenges, so they have max time to propagate
	for _, authSolver := range authSolvers {
		authz := authSolver.authz

		chlg, err := challenge.FindChallenge(challenge.DNS01, authz)
		if err == nil {
			if _, ok := uniq[authz.Identifier.Value+chlg.Token]; ok {
				log.Debug("Duplicate token; skipping pre-solve.",
					slog.String("identifier", authSolver.authz.Identifier.Value),
					slog.String("type", challenge.DNS01.String()),
				)

				continue
			}

			uniq[authz.Identifier.Value+chlg.Token] = struct{}{}
		}

		if solvr, ok := authSolver.solver.(preSolver); ok {
			err := solvr.PreSolve(ctx, authz)
			if err != nil {
				failures.Add(challenge.GetTargetedDomain(authz), err)
			}
		}
	}

	defer func() {
		// Clean all created TXT records
		for _, authSolver := range authSolvers {
			chlg, err := challenge.FindChallenge(challenge.DNS01, authSolver.authz)
			if err == nil {
				if _, ok := uniq[authSolver.authz.Identifier.Value+chlg.Token]; ok {
					delete(uniq, authSolver.authz.Identifier.Value+chlg.Token)
				} else {
					log.Debug("Duplicate token; skipping cleanup.",
						slog.String("identifier", authSolver.authz.Identifier.Value),
						slog.String("type", challenge.DNS01.String()),
					)

					continue
				}
			}

			cleanUp(ctx, authSolver.solver, authSolver.authz)
		}
	}()

	// Finally, solve all challenges for real.
	for _, authSolver := range authSolvers {
		authz := authSolver.authz

		domain := challenge.GetTargetedDomain(authz)
		if failures.Has(domain) {
			// already failed in previous loop
			continue
		}

		err := authSolver.solver.Solve(ctx, authz)
		if err != nil {
			failures.Add(domain, err)
		}
	}
}

func cleanUp(ctx context.Context, solvr solver, authz acme.Authorization) {
	s, ok := solvr.(cleanup)
	if !ok {
		return
	}

	err := s.CleanUp(ctx, authz)
	if err != nil {
		log.Warn("Cleaning up failed.",
			log.DomainAttr(challenge.GetTargetedDomain(authz)),
			log.ErrorAttr(err))
	}
}
