package resolver

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/xenolf/lego/emca/internal/secure"
	"github.com/xenolf/lego/emca/le"
	"github.com/xenolf/lego/log"
)

const (
	statusInvalid = "invalid"
	statusValid   = "valid"
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

// TODO comments
type Prober struct {
	jws           *secure.JWS
	solverManager *SolverManager
}

func NewProber(jws *secure.JWS, solverManager *SolverManager) *Prober {
	return &Prober{
		solverManager: solverManager,
		jws:           jws,
	}
}

// SolveForAuthz Looks through the challenge combinations to find a solvable match.
// Then solves the challenges in series and returns.
func (c *Prober) SolveForAuthz(authorizations []le.Authorization) error {
	failures := make(obtainError)

	var authSolvers []*selectedAuthSolver

	// loop through the resources, basically through the domains. First pass just selects a solver for each authz.
	for _, authz := range authorizations {
		if authz.Status == statusValid {
			// Boulder might recycle recent validated authz (see issue #267)
			log.Infof("[%s] acme: Authorization already valid; skipping challenge", authz.Identifier.Value)
			continue
		}

		if i, solvr := c.solverManager.chooseSolver(authz, authz.Identifier.Value); solvr != nil {
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
	// even an empty obtainError is a non-nil error value
	if len(failures) > 0 {
		return failures
	}
	return nil
}

func validate(j *secure.JWS, domain, uri string, _ le.Challenge) error {
	var chlng le.Challenge

	// Challenge initiation is done by sending a JWS payload containing the trivial JSON object `{}`.
	// We use an empty struct instance as the postJSON payload here to achieve this result.
	hdr, err := j.PostJSON(uri, struct{}{}, &chlng)
	if err != nil {
		return err
	}

	// After the path is sent, the ACME server will access our server.
	// Repeatedly check the server for an updated status on our request.
	for {
		switch chlng.Status {
		case statusValid:
			log.Infof("[%s] The server validated our request", domain)
			return nil
		case "pending":
		case "processing":
		case statusInvalid:
			return handleChallengeError(chlng)
		default:
			return errors.New("the server returned an unexpected state")
		}

		ra, err := strconv.Atoi(hdr.Get("Retry-After"))
		if err != nil {
			// The ACME server MUST return a Retry-After.
			// If it doesn't, we'll just poll hard.
			ra = 5
		}

		time.Sleep(time.Duration(ra) * time.Second)

		resp, err := j.PostAsGet(uri, &chlng)
		if resp != nil {
			hdr = resp.Header
		}
		if err != nil {
			return err
		}

	}
}

func handleChallengeError(chlng le.Challenge) error {
	return chlng.Error
}
