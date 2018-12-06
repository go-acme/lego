package http01

import (
	"fmt"

	"github.com/xenolf/lego/emca/challenge"
	"github.com/xenolf/lego/emca/le"
	"github.com/xenolf/lego/emca/le/api"
	"github.com/xenolf/lego/log"
)

// ChallengePath returns the URL path for the `http-01` challenge
func ChallengePath(token string) string {
	return "/.well-known/acme-challenge/" + token
}

// FIXME refactor
type validateFunc func(core *api.Core, domain, uri string, chlng le.Challenge) error

type Challenge struct {
	core     *api.Core
	validate validateFunc
	provider challenge.Provider
}

func NewChallenge(core *api.Core, validate validateFunc, provider challenge.Provider) *Challenge {
	return &Challenge{
		core:     core,
		validate: validate,
		provider: provider,
	}
}

func (c *Challenge) SetProvider(provider challenge.Provider) {
	c.provider = provider
}

func (c *Challenge) Solve(chlng le.Challenge, domain string) error {
	log.Infof("[%s] acme: Trying to solve HTTP-01", domain)

	// Generate the Key Authorization for the challenge
	keyAuth, err := c.core.GetKeyAuthorization(chlng.Token)
	if err != nil {
		return err
	}

	err = c.provider.Present(domain, chlng.Token, keyAuth)
	if err != nil {
		return fmt.Errorf("[%s] error presenting token: %v", domain, err)
	}
	defer func() {
		err := c.provider.CleanUp(domain, chlng.Token, keyAuth)
		if err != nil {
			log.Warnf("[%s] error cleaning up: %v", domain, err)
		}
	}()

	return c.validate(c.core, domain, chlng.URL, le.Challenge{Type: chlng.Type, Token: chlng.Token, KeyAuthorization: keyAuth})
}
