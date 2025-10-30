// Package lightsail implements a DNS provider for solving the DNS-01 challenge using AWS Lightsail DNS.
package lightsail

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lightsail/types"
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
)

// Environment variables names.
const (
	envNamespace = "LIGHTSAIL_"

	EnvRegion  = envNamespace + "REGION"
	EnvDNSZone = "DNS_ZONE"

	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
)

const maxRetries = 5

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	DNSZone            string
	Region             string
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	client *lightsail.Client
	config *Config
}

// NewDNSProvider returns a DNSProvider instance configured for the AWS Lightsail service.
//
// AWS Credentials are automatically detected in the following locations
// and prioritized in the following order:
//  1. Environment variables: AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY,
//     [AWS_SESSION_TOKEN], [DNS_ZONE], [LIGHTSAIL_REGION]
//  2. Shared credentials file (defaults to ~/.aws/credentials)
//  3. Amazon EC2 IAM role
//
// public hosted zone via the FQDN.
//
// See also: https://github.com/aws/aws-sdk-go/wiki/configuring-sdk
func NewDNSProvider() (*DNSProvider, error) {
	config := NewDefaultConfig()

	config.DNSZone = env.GetOrFile(EnvDNSZone)
	config.Region = env.GetOrDefaultString(EnvRegion, "us-east-1")

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for AWS Lightsail.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("lightsail: the configuration of the DNS provider is nil")
	}

	ctx := context.Background()

	cfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(config.Region),
		awsconfig.WithRetryer(func() aws.Retryer {
			return retry.NewStandard(func(options *retry.StandardOptions) {
				options.MaxAttempts = maxRetries

				// It uses a basic exponential backoff algorithm that returns an initial
				// delay of ~400ms with an upper limit of ~30 seconds which should prevent
				// causing a high number of consecutive throttling errors.
				// For reference: Route 53 enforces an account-wide(!) 5req/s query limit.
				options.Backoff = retry.BackoffDelayerFunc(func(attempt int, err error) (time.Duration, error) {
					retryCount := min(attempt, 7)

					delay := (1 << uint(retryCount)) * (rand.Intn(50) + 200)

					return time.Duration(delay) * time.Millisecond, nil
				})
			})
		}),
	)
	if err != nil {
		return nil, err
	}

	return &DNSProvider{
		config: config,
		client: lightsail.NewFromConfig(cfg),
	}, nil
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, _, keyAuth string) error {
	ctx := context.Background()
	info := dns01.GetChallengeInfo(domain, keyAuth)

	params := &lightsail.CreateDomainEntryInput{
		DomainName: aws.String(d.config.DNSZone),
		DomainEntry: &awstypes.DomainEntry{
			Name:   aws.String(info.EffectiveFQDN),
			Target: aws.String(strconv.Quote(info.Value)),
			Type:   aws.String("TXT"),
		},
	}

	_, err := d.client.CreateDomainEntry(ctx, params)
	if err != nil {
		return fmt.Errorf("lightsail: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, _, keyAuth string) error {
	ctx := context.Background()
	info := dns01.GetChallengeInfo(domain, keyAuth)

	params := &lightsail.DeleteDomainEntryInput{
		DomainName: aws.String(d.config.DNSZone),
		DomainEntry: &awstypes.DomainEntry{
			Name:   aws.String(info.EffectiveFQDN),
			Type:   aws.String("TXT"),
			Target: aws.String(strconv.Quote(info.Value)),
		},
	}

	_, err := d.client.DeleteDomainEntry(ctx, params)
	if err != nil {
		return fmt.Errorf("lightsail: %w", err)
	}

	return nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}
