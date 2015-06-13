package acme

type dvsniChallenge struct{}

func (s *dvsniChallenge) CanSolve(domain string) bool {
	return false
}

func (s *dvsniChallenge) Solve(challenge challenge, domain string) error {
	return nil
}
