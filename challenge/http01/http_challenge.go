package http01

import (
	"context"
	"fmt"
	"path"
	"time"

	"github.com/go-acme/lego/v5/acme"
	"github.com/go-acme/lego/v5/acme/api"
	"github.com/go-acme/lego/v5/challenge"
	"github.com/go-acme/lego/v5/log"
)

const PathPrefix = "/.well-known/acme-challenge/"

type ValidateFunc func(ctx context.Context, core *api.Core, domain string, chlng acme.Challenge) error

type ChallengeOption func(*Challenge) error

// SetDelay sets a delay between the start of the HTTP server and the challenge validation.
func SetDelay(delay time.Duration) ChallengeOption {
	return func(chlg *Challenge) error {
		chlg.delay = delay
		return nil
	}
}

// ChallengePath returns the URL path for the `http-01` challenge.
func ChallengePath(token string) string {
	if isBase64url(token) {
		return PathPrefix + token
	}

	return path.Join(PathPrefix, "invalid")
}

type Challenge struct {
	core     *api.Core
	validate ValidateFunc
	provider challenge.Provider
	delay    time.Duration
}

func NewChallenge(core *api.Core, validate ValidateFunc, provider challenge.Provider, opts ...ChallengeOption) *Challenge {
	chlg := &Challenge{
		core:     core,
		validate: validate,
		provider: provider,
	}

	for _, opt := range opts {
		err := opt(chlg)
		if err != nil {
			log.Warn("http01: Challenge option skipped.", log.ErrorAttr(err))
		}
	}

	return chlg
}

func (c *Challenge) Solve(ctx context.Context, authz acme.Authorization) error {
	domain := challenge.GetTargetedDomain(authz)
	log.Info("http01: Trying to solve HTTP-01.", log.DomainAttr(domain))

	chlng, err := challenge.FindChallenge(challenge.HTTP01, authz)
	if err != nil {
		return err
	}

	// Generate the Key Authorization for the challenge
	keyAuth, err := c.core.GetKeyAuthorization(chlng.Token)
	if err != nil {
		return err
	}

	err = c.provider.Present(ctx, authz.Identifier.Value, chlng.Token, keyAuth)
	if err != nil {
		return fmt.Errorf("[%s] http01: error presenting token: %w", domain, err)
	}

	defer func() {
		err := c.provider.CleanUp(ctx, authz.Identifier.Value, chlng.Token, keyAuth)
		if err != nil {
			log.Warn("http01: cleaning up failed.", log.DomainAttr(domain), log.ErrorAttr(err))
		}
	}()

	if c.delay > 0 {
		time.Sleep(c.delay)
	}

	chlng.KeyAuthorization = keyAuth

	return c.validate(ctx, c.core, domain, chlng)
}

// https://en.wikipedia.org/wiki/Base64#Alphabet
// ^[A-Za-z0-9_-]+$
// It MUST NOT contain any characters outside the base64url alphabet
// and MUST NOT include base64 padding characters ("=")
// https://www.rfc-editor.org/rfc/rfc8555.html#section-8.3
func isBase64url(s string) bool {
	for _, c := range s {
		isUpper := c >= 'A' && c <= 'Z'
		isLower := c >= 'a' && c <= 'z'
		isDigit := c >= '0' && c <= '9'
		isSpecial := c == '-' || c == '_'

		if !isUpper && !isLower && !isDigit && !isSpecial {
			return false
		}
	}

	return true
}
