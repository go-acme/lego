package resolver

import (
	"context"
	"time"

	"github.com/go-acme/lego/v5/acme"
	"github.com/go-acme/lego/v5/challenge"
)

type preSolverMock struct {
	preSolve map[string]error
	solve    map[string]error
	cleanUp  map[string]error
}

func (s *preSolverMock) PreSolve(ctx context.Context, authorization acme.Authorization) error {
	return s.preSolve[authorization.Identifier.Value]
}

func (s *preSolverMock) Solve(ctx context.Context, authorization acme.Authorization) error {
	return s.solve[authorization.Identifier.Value]
}

func (s *preSolverMock) CleanUp(authorization acme.Authorization) error {
	return s.cleanUp[authorization.Identifier.Value]
}

func createStubAuthorizationHTTP01(domain, status string) acme.Authorization {
	return acme.Authorization{
		Status:  status,
		Expires: time.Now(),
		Identifier: acme.Identifier{
			Type:  challenge.HTTP01.String(),
			Value: domain,
		},
		Challenges: []acme.Challenge{
			{
				Type:      challenge.HTTP01.String(),
				Validated: time.Now(),
				Error:     nil,
			},
		},
	}
}
