package dns01

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-acme/lego/v5/acme"
	"github.com/go-acme/lego/v5/acme/api"
	"github.com/go-acme/lego/v5/challenge"
	"github.com/go-acme/lego/v5/log"
	"github.com/go-acme/lego/v5/platform/wait"
)

const (
	// DefaultPropagationTimeout default propagation timeout.
	DefaultPropagationTimeout = 60 * time.Second

	// DefaultPollingInterval default polling interval.
	DefaultPollingInterval = 2 * time.Second

	// DefaultTTL default TTL.
	DefaultTTL = 120
)

type ValidateFunc func(ctx context.Context, core *api.Core, domain string, chlng acme.Challenge) error

// Challenge implements the dns-01 challenge.
type Challenge struct {
	core     *api.Core
	validate ValidateFunc
	provider challenge.Provider
	preCheck preCheck
}

func NewChallenge(core *api.Core, validate ValidateFunc, provider challenge.Provider, opts ...ChallengeOption) *Challenge {
	chlg := &Challenge{
		core:     core,
		validate: validate,
		provider: provider,
		preCheck: newPreCheck(),
	}

	for _, opt := range opts {
		err := opt(chlg)
		if err != nil {
			log.Warn("Challenge option skipped.", "error", err)
		}
	}

	return chlg
}

// PreSolve just submits the txt record to the dns provider.
// It does not validate record propagation or do anything at all with the ACME server.
func (c *Challenge) PreSolve(ctx context.Context, authz acme.Authorization) error {
	domain := challenge.GetTargetedDomain(authz)
	log.Info("acme: Preparing to solve DNS-01.", "domain", domain)

	chlng, err := challenge.FindChallenge(challenge.DNS01, authz)
	if err != nil {
		return err
	}

	if c.provider == nil {
		return fmt.Errorf("[%s] acme: no DNS Provider configured", domain)
	}

	// Generate the Key Authorization for the challenge
	keyAuth, err := c.core.GetKeyAuthorization(chlng.Token)
	if err != nil {
		return err
	}

	err = c.provider.Present(ctx, authz.Identifier.Value, chlng.Token, keyAuth)
	if err != nil {
		return fmt.Errorf("[%s] acme: error presenting token: %w", domain, err)
	}

	return nil
}

func (c *Challenge) Solve(ctx context.Context, authz acme.Authorization) error {
	domain := challenge.GetTargetedDomain(authz)
	log.Info("acme: Trying to solve DNS-01.", "domain", domain)

	chlng, err := challenge.FindChallenge(challenge.DNS01, authz)
	if err != nil {
		return err
	}

	// Generate the Key Authorization for the challenge
	keyAuth, err := c.core.GetKeyAuthorization(chlng.Token)
	if err != nil {
		return err
	}

	info := GetChallengeInfo(ctx, authz.Identifier.Value, keyAuth)

	var timeout, interval time.Duration

	switch provider := c.provider.(type) {
	case challenge.ProviderTimeout:
		timeout, interval = provider.Timeout()
	default:
		timeout, interval = DefaultPropagationTimeout, DefaultPollingInterval
	}

	log.Info("acme: Checking DNS record propagation.",
		"domain", domain, "nameservers", strings.Join(DefaultClient().recursiveNameservers, ","))

	time.Sleep(interval)

	err = wait.For("propagation", timeout, interval, func() (bool, error) {
		stop, errP := c.preCheck.call(ctx, domain, info.EffectiveFQDN, info.Value)
		if !stop || errP != nil {
			log.Info("acme: Waiting for DNS record propagation.", "domain", domain)
		}

		return stop, errP
	})
	if err != nil {
		return err
	}

	chlng.KeyAuthorization = keyAuth

	return c.validate(ctx, c.core, domain, chlng)
}

// CleanUp cleans the challenge.
func (c *Challenge) CleanUp(ctx context.Context, authz acme.Authorization) error {
	log.Info("acme: Cleaning DNS-01 challenge.", "domain", challenge.GetTargetedDomain(authz))

	chlng, err := challenge.FindChallenge(challenge.DNS01, authz)
	if err != nil {
		return err
	}

	keyAuth, err := c.core.GetKeyAuthorization(chlng.Token)
	if err != nil {
		return err
	}

	return c.provider.CleanUp(ctx, authz.Identifier.Value, chlng.Token, keyAuth)
}

func (c *Challenge) Sequential() (bool, time.Duration) {
	if p, ok := c.provider.(sequential); ok {
		return ok, p.Sequential()
	}

	return false, 0
}

type sequential interface {
	Sequential() time.Duration
}

// ChallengeInfo contains the information use to create the TXT record.
type ChallengeInfo struct {
	// FQDN is the full-qualified challenge domain (i.e. `_acme-challenge.[domain].`)
	FQDN string

	// EffectiveFQDN contains the resulting FQDN after the CNAMEs resolutions.
	EffectiveFQDN string

	// Value contains the value for the TXT record.
	Value string
}

// GetChallengeInfo returns information used to create a DNS record which will fulfill the `dns-01` challenge.
func GetChallengeInfo(ctx context.Context, domain, keyAuth string) ChallengeInfo {
	keyAuthShaBytes := sha256.Sum256([]byte(keyAuth))
	// base64URL encoding without padding
	value := base64.RawURLEncoding.EncodeToString(keyAuthShaBytes[:sha256.Size])

	ok, _ := strconv.ParseBool(os.Getenv("LEGO_DISABLE_CNAME_SUPPORT"))

	return ChallengeInfo{
		Value:         value,
		FQDN:          getChallengeFQDN(ctx, domain, false),
		EffectiveFQDN: getChallengeFQDN(ctx, domain, !ok),
	}
}

func getChallengeFQDN(ctx context.Context, domain string, followCNAME bool) string {
	fqdn := fmt.Sprintf("_acme-challenge.%s.", domain)

	if !followCNAME {
		return fqdn
	}

	return DefaultClient().lookupCNAME(ctx, fqdn)
}
