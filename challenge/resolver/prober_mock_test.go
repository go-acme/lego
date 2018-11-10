package resolver

import (
	"time"

	"github.com/xenolf/lego/challenge"
	"github.com/xenolf/lego/le"
)

type preSolverMock struct {
	preSolve map[string]error
	solve    map[string]error
	cleanUp  map[string]error
}

func (s *preSolverMock) PreSolve(authorization le.Authorization) error {
	return s.preSolve[authorization.Identifier.Value]
}
func (s *preSolverMock) Solve(authorization le.Authorization) error {
	return s.solve[authorization.Identifier.Value]
}
func (s *preSolverMock) CleanUp(authorization le.Authorization) error {
	return s.cleanUp[authorization.Identifier.Value]
}

func createStubAuthorizationHTTP01(domain, status string) le.Authorization {
	return le.Authorization{
		Status:  status,
		Expires: time.Now(),
		Identifier: le.Identifier{
			Type:  challenge.HTTP01.String(),
			Value: domain,
		},
		Challenges: []le.Challenge{
			{
				Type:      challenge.HTTP01.String(),
				Validated: time.Now(),
				Error:     nil,
			},
		},
	}
}
