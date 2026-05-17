package resolver

import (
	"context"
	"errors"
	"fmt"
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
	authz   acme.Authorization
	solvers []solver
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
// If the first solver fails, subsequent matching solvers are tried as fallback.
// For example, when both HTTP-01 and TLS-ALPN-01 are enabled and TLS-ALPN-01 fails
// due to CDN TLS termination, HTTP-01 is automatically tried next.
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

		solvers := p.solverManager.chooseSolvers(authz)

		if len(solvers) > 0 {
			authSolver := &selectedAuthSolver{authz: authz, solvers: solvers}

			// Use the first solver for categorization (sequential vs parallel).
			// All solvers for the same authz should have the same nature,
			// but we only care about the first one for grouping.
			solvr := solvers[0]

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
		domain := challenge.GetTargetedDomain(authSolver.authz)

		// Try each solver in order as fallback
		var lastErr error
		for _, solvr := range authSolver.solvers {
			if lastErr != nil {
				log.Debug("Trying fallback solver.",
					log.DomainAttr(domain),
					slog.String("solver_type", fmt.Sprintf("%T", solvr)),
				)
			}

			chlg, _ := challenge.FindChallenge(challenge.DNS01, authSolver.authz)

			if preSolvr, ok := solvr.(preSolver); ok {
				if _, ok := uniq[authSolver.authz.Identifier.Value+chlg.Token]; ok && chlg.Token != "" {
					log.Debug("Duplicate token; skipping pre-solve.",
						slog.String("identifier", authSolver.authz.Identifier.Value),
						slog.String("type", challenge.DNS01.String()),
					)
					continue
				}

				err := preSolvr.PreSolve(ctx, authSolver.authz)
				if err != nil {
					failures.Add(domain, err)
					cleanUp(ctx, solvr, authSolver.authz)
					lastErr = err
					continue
				}

				uniq[authSolver.authz.Identifier.Value+chlg.Token] = struct{}{}
			}

			// Solve the challenge
			err := solvr.Solve(ctx, authSolver.authz)
			if err == nil {
				lastErr = nil
				break
			}

			lastErr = err
			log.Warn("Solver failed, trying fallback.",
				log.DomainAttr(domain),
				slog.String("solver_type", fmt.Sprintf("%T", solvr)),
				log.ErrorAttr(err),
			)
		}

		if lastErr != nil {
			failures.Add(domain, lastErr)
			continue
		}

		chlg, _ := challenge.FindChallenge(challenge.DNS01, authSolver.authz)

		if _, ok := uniq[authSolver.authz.Identifier.Value+chlg.Token]; ok || chlg.Token == "" {
			// Clean challenge
			for _, solvr := range authSolver.solvers {
				cleanUp(ctx, solvr, authSolver.authz)
			}

			if len(authSolvers)-1 > i {
				solvr := authSolver.solvers[0].(sequential)
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

		// Only pre-solve the first solver; fallback solvers will do their own pre-solve if needed
		if solvr, ok := authSolver.solvers[0].(preSolver); ok {
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

			for _, solvr := range authSolver.solvers {
				cleanUp(ctx, solvr, authSolver.authz)
			}
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

		// Try each solver in order as fallback
		var lastErr error
		for _, solvr := range authSolver.solvers {
			if lastErr != nil {
				log.Debug("Trying fallback solver.",
					log.DomainAttr(domain),
					slog.String("solver_type", fmt.Sprintf("%T", solvr)),
				)
			}

			// Pre-solve for fallback solvers if needed (skip for first, already done above)
			if lastErr != nil {
				if preSolvr, ok := solvr.(preSolver); ok {
					err := preSolvr.PreSolve(ctx, authz)
					if err != nil {
						failures.Add(domain, err)
						cleanUp(ctx, solvr, authz)
						lastErr = err
						continue
					}
				}
			}

			err := solvr.Solve(ctx, authz)
			if err == nil {
				lastErr = nil
				break
			}

			lastErr = err
			log.Warn("Solver failed, trying fallback.",
				log.DomainAttr(domain),
				slog.String("solver_type", fmt.Sprintf("%T", solvr)),
				log.ErrorAttr(err),
			)
		}

		if lastErr != nil {
			failures.Add(domain, lastErr)
		}
	}
}

func cleanUp(ctx context.Context, solvr solver, authz acme.Authorization) {
	if solvr, ok := solvr.(cleanup); ok {
		domain := challenge.GetTargetedDomain(authz)

		err := solvr.CleanUp(ctx, authz)
		if err != nil {
			log.Warn("Cleaning up failed.", log.DomainAttr(domain), log.ErrorAttr(err))
		}
	}
}
