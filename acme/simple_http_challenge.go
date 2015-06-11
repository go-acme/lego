package acme

type simpleHTTPChallenge struct{}

func (s *simpleHTTPChallenge) CanSolve() bool {
	return true
}

func (s *simpleHTTPChallenge) Solve(challenge challenge) {

}
