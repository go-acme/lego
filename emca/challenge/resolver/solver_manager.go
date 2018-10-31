package resolver

import (
	"fmt"
	"net"

	"github.com/xenolf/lego/emca/le"
	"github.com/xenolf/lego/log"

	"github.com/xenolf/lego/emca/challenge"
	"github.com/xenolf/lego/emca/challenge/dns01"
	"github.com/xenolf/lego/emca/challenge/http01"
	"github.com/xenolf/lego/emca/challenge/tlsalpn01"
	"github.com/xenolf/lego/emca/internal/secure"
)

type SolverManager struct {
	jws     *secure.JWS
	solvers map[challenge.Type]solver
}

func NewSolversManager(jws *secure.JWS) *SolverManager {
	// REVIEW: best possibility?
	// Add all available solvers with the right index as per ACME spec to this map.
	// Otherwise they won't be found.
	solvers := map[challenge.Type]solver{
		challenge.HTTP01:    http01.NewChallenge(jws, validate, &http01.ProviderServer{}),
		challenge.TLSALPN01: tlsalpn01.NewChallenge(jws, validate, &tlsalpn01.ProviderServer{}),
	}

	return &SolverManager{
		solvers: solvers,
		jws:     jws,
	}
}

// SetHTTPAddress specifies a custom interface:port to be used for HTTP based challenges.
// If this option is not used, the default port 80 and all interfaces will be used.
// To only specify a port and no interface use the ":port" notation.
//
// NOTE: This REPLACES any custom HTTP provider previously set by calling
// c.SetChallengeProvider with the default HTTP challenge provider.
func (c *SolverManager) SetHTTPAddress(iface string) error {
	host, port, err := net.SplitHostPort(iface)
	if err != nil {
		return err
	}

	if chlng, ok := c.solvers[challenge.HTTP01]; ok {
		chlng.(*http01.Challenge).SetProvider(http01.NewProviderServer(host, port))
	}

	return nil
}

// SetTLSAddress specifies a custom interface:port to be used for TLS based challenges.
// If this option is not used, the default port 443 and all interfaces will be used.
// To only specify a port and no interface use the ":port" notation.
//
// NOTE: This REPLACES any custom TLS-ALPN provider previously set by calling
// c.SetChallengeProvider with the default TLS-ALPN challenge provider.
func (c *SolverManager) SetTLSAddress(iface string) error {
	host, port, err := net.SplitHostPort(iface)
	if err != nil {
		return err
	}

	if chlng, ok := c.solvers[challenge.TLSALPN01]; ok {
		chlng.(*tlsalpn01.Challenge).SetProvider(tlsalpn01.NewProviderServer(host, port))
	}

	return nil
}

// SetChallengeProvider specifies a custom provider p that can solve the given challenge type.
func (c *SolverManager) SetChallengeProvider(chlg challenge.Type, p challenge.Provider) error {
	switch chlg {
	case challenge.HTTP01:
		c.solvers[chlg] = http01.NewChallenge(c.jws, validate, p)
	case challenge.DNS01:
		c.solvers[chlg] = dns01.NewChallenge(c.jws, validate, p)
	case challenge.TLSALPN01:
		c.solvers[chlg] = tlsalpn01.NewChallenge(c.jws, validate, p)
	default:
		return fmt.Errorf("unknown challenge %v", chlg)
	}
	return nil
}

// ExcludeChallenges explicitly removes challenges from the pool for solving.
func (c *SolverManager) ExcludeChallenges(challenges []challenge.Type) {
	// Loop through all challenges and delete the requested one if found.
	for _, chlg := range challenges {
		delete(c.solvers, chlg)
	}
}

// Checks all challenges from the server in order and returns the first matching solver.
func (c *SolverManager) chooseSolver(auth le.Authorization, domain string) (int, solver) {
	for i, chlg := range auth.Challenges {
		if solvr, ok := c.solvers[challenge.Type(chlg.Type)]; ok {
			return i, solvr
		}
		log.Infof("[%s] acme: Could not find solvr for: %s", domain, chlg.Type)
	}
	return 0, nil
}
