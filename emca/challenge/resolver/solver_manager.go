package resolver

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/xenolf/lego/emca/api"
	"github.com/xenolf/lego/emca/challenge"
	"github.com/xenolf/lego/emca/challenge/dns01"
	"github.com/xenolf/lego/emca/challenge/http01"
	"github.com/xenolf/lego/emca/challenge/tlsalpn01"
	"github.com/xenolf/lego/emca/le"
	"github.com/xenolf/lego/log"
)

type SolverManager struct {
	core    *api.Core
	solvers map[challenge.Type]solver
}

func NewSolversManager(core *api.Core) *SolverManager {
	// REVIEW: best possibility?
	// Add all available solvers with the right index as per ACME spec to this map.
	// Otherwise they won't be found.
	solvers := map[challenge.Type]solver{
		challenge.HTTP01:    http01.NewChallenge(core, validate, &http01.ProviderServer{}),
		challenge.TLSALPN01: tlsalpn01.NewChallenge(core, validate, &tlsalpn01.ProviderServer{}),
	}

	return &SolverManager{
		solvers: solvers,
		core:    core,
	}
}

// SetHTTPAddress specifies a custom interface:port to be used for HTTP based challenges.
// If this option is not used, the default port 80 and all interfaces will be used.
// To only specify a port and no interface use the ":port" notation.
//
// NOTE: This REPLACES any custom HTTP provider previously set by calling
// c.SetProvider with the default HTTP challenge provider.
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
// c.SetProvider with the default TLS-ALPN challenge provider.
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

// SetProvider specifies a custom provider p that can solve the given challenge type.
func (c *SolverManager) SetProvider(chlg challenge.Type, p challenge.Provider) error {
	switch chlg {
	case challenge.HTTP01:
		c.solvers[chlg] = http01.NewChallenge(c.core, validate, p)
	case challenge.DNS01:
		c.solvers[chlg] = dns01.NewChallenge(c.core, validate, p)
	case challenge.TLSALPN01:
		c.solvers[chlg] = tlsalpn01.NewChallenge(c.core, validate, p)
	default:
		return fmt.Errorf("unknown challenge %v", chlg)
	}
	return nil
}

// Exclude explicitly removes challenges from the pool for solving.
func (c *SolverManager) Exclude(challenges []challenge.Type) {
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

func validate(core *api.Core, domain, uri string, _ le.Challenge) error {
	var chlng le.Challenge

	// Challenge initiation is done by sending a JWS payload containing the trivial JSON object `{}`.
	// We use an empty struct instance as the postJSON payload here to achieve this result.
	resp, err := core.Post(uri, struct{}{}, &chlng)
	if err != nil {
		return err
	}

	// After the path is sent, the ACME server will access our server.
	// Repeatedly check the server for an updated status on our request.
	for {
		switch chlng.Status {
		case le.StatusValid:
			log.Infof("[%s] The server validated our request", domain)
			return nil
		case le.StatusPending:
		case le.StatusProcessing:
		case le.StatusInvalid:
			return chlng.Error
		default:
			return errors.New("the server returned an unexpected state")
		}

		ra, err := strconv.Atoi(resp.Header.Get("Retry-After"))
		if err != nil {
			// The ACME server MUST return a Retry-After.
			// If it doesn't, we'll just poll hard.
			ra = 5
		}

		time.Sleep(time.Duration(ra) * time.Second)

		resp, err = core.PostAsGet(uri, &chlng)
		if err != nil {
			return err
		}
	}
}
