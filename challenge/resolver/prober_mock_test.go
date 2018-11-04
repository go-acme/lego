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

func (s *preSolverMock) PreSolve(challenge le.Challenge, domain string) error {
	return s.preSolve[domain]
}
func (s *preSolverMock) Solve(challenge le.Challenge, domain string) error {
	return s.solve[domain]
}
func (s *preSolverMock) CleanUp(challenge le.Challenge, domain string) error {
	return s.cleanUp[domain]
}

type solverMock struct {
	solve   map[string]error
	cleanUp map[string]error
}

func (s *solverMock) Solve(challenge le.Challenge, domain string) error   { return s.solve[domain] }
func (s *solverMock) CleanUp(challenge le.Challenge, domain string) error { return s.cleanUp[domain] }

func createStubAuthorizationHTTP01(domain, status string) le.Authorization {
	return le.Authorization{
		Status:  status,
		Expires: time.Now(),
		Identifier: le.Identifier{
			Type:  string(challenge.HTTP01),
			Value: domain,
		},
		Challenges: []le.Challenge{
			{
				Type:      string(challenge.HTTP01),
				Validated: time.Now(),
				Error:     nil,
			},
		},
	}
}
