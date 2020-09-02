package resolver

import (
	"time"

	"github.com/go-acme/lego/v4/acme"
	"github.com/go-acme/lego/v4/challenge"
)

type preSolverMock struct {
	preSolve map[string]error
	solve    map[string]error
	cleanUp  map[string]error
}

func (s *preSolverMock) PreSolve(authorization acme.Authorization) error {
	return s.preSolve[authorization.Identifier.Value]
}

func (s *preSolverMock) Solve(authorization acme.Authorization) error {
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
