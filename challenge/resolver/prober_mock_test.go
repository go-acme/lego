package resolver

import (
	"context"
	"fmt"
	"time"

	"github.com/go-acme/lego/v5/acme"
	"github.com/go-acme/lego/v5/challenge"
)

type preSolverMock struct {
	preSolve map[string]error
	solve    map[string]error
	cleanUp  map[string]error

	preSolveCounter int
	solveCounter    int
	cleanUpCounter  int
}

func (s *preSolverMock) PreSolve(ctx context.Context, authorization acme.Authorization) error {
	s.preSolveCounter++

	return s.preSolve[authorization.Identifier.Value]
}

func (s *preSolverMock) Solve(ctx context.Context, authorization acme.Authorization) error {
	s.solveCounter++

	return s.solve[authorization.Identifier.Value]
}

func (s *preSolverMock) CleanUp(ctx context.Context, authorization acme.Authorization) error {
	s.cleanUpCounter++

	return s.cleanUp[authorization.Identifier.Value]
}

func (s *preSolverMock) String() string {
	return fmt.Sprintf("PreSolve: %d, Solve: %d, CleanUp: %d", s.preSolveCounter, s.solveCounter, s.cleanUpCounter)
}

func createStubAuthorizationHTTP01(domain, status string) acme.Authorization {
	return createStubAuthorization(domain, status, false, acme.Challenge{
		Type:      challenge.HTTP01.String(),
		Validated: time.Now(),
	})
}

func createStubAuthorizationDNS01(domain string, wildcard bool) acme.Authorization {
	var chlgs []acme.Challenge

	if wildcard {
		chlgs = append(chlgs, acme.Challenge{
			Type:      challenge.HTTP01.String(),
			Validated: time.Now(),
		})
	}

	chlgs = append(chlgs, acme.Challenge{
		Type:      challenge.DNS01.String(),
		Validated: time.Now(),
	})

	return createStubAuthorization(domain, acme.StatusProcessing, wildcard, chlgs...)
}

func createStubAuthorization(domain, status string, wildcard bool, chlgs ...acme.Challenge) acme.Authorization {
	return acme.Authorization{
		Wildcard: wildcard,
		Status:   status,
		Expires:  time.Now(),
		Identifier: acme.Identifier{
			Type:  "dns",
			Value: domain,
		},
		Challenges: chlgs,
	}
}
