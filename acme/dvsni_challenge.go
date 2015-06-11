package acme

type dvsniChallenge struct{}

func (s *dvsniChallenge) CanSolve() bool {
	return false
}

func (s *dvsniChallenge) Solve(challenge challenge, domain string) {

}
