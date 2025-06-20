// Package route53 implements a DNS provider for solving the DNS-01 challenge using AWS Route 53 DNS.
package route53

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/platform/wait"
	"github.com/go-acme/lego/v4/providers/dns/internal/ptr"
)

// Environment variables names.
const (
	envNamespace = "AWS_"

	EnvAccessKeyID     = envNamespace + "ACCESS_KEY_ID"
	EnvSecretAccessKey = envNamespace + "SECRET_ACCESS_KEY"
	EnvRegion          = envNamespace + "REGION"
	EnvHostedZoneID    = envNamespace + "HOSTED_ZONE_ID"
	EnvMaxRetries      = envNamespace + "MAX_RETRIES"
	EnvAssumeRoleArn   = envNamespace + "ASSUME_ROLE_ARN"
	EnvExternalID      = envNamespace + "EXTERNAL_ID"
	EnvPrivateZone     = envNamespace + "PRIVATE_ZONE"

	EnvWaitForRecordSetsChanged = envNamespace + "WAIT_FOR_RECORD_SETS_CHANGED"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
)

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	// Static credential chain.
	// These are not set via environment for the time being and are only used if they are explicitly provided.
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	Region          string

	HostedZoneID  string
	MaxRetries    int
	AssumeRoleArn string
	ExternalID    string
	PrivateZone   bool

	WaitForRecordSetsChanged bool

	TTL                int
	PropagationTimeout time.Duration
	PollingInterval    time.Duration

	Client *route53.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		HostedZoneID:  env.GetOrFile(EnvHostedZoneID),
		MaxRetries:    env.GetOrDefaultInt(EnvMaxRetries, 5),
		AssumeRoleArn: env.GetOrDefaultString(EnvAssumeRoleArn, ""),
		ExternalID:    env.GetOrDefaultString(EnvExternalID, ""),
		PrivateZone:   env.GetOrDefaultBool(EnvPrivateZone, false),

		WaitForRecordSetsChanged: env.GetOrDefaultBool(EnvWaitForRecordSetsChanged, true),

		TTL:                env.GetOrDefaultInt(EnvTTL, 10),
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, 2*time.Minute),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, 4*time.Second),
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	client *route53.Client
	config *Config
}

// NewDNSProvider returns a DNSProvider instance configured for the AWS Route 53 service.
//
// AWS Credentials are automatically detected in the following locations and prioritized in the following order:
//  1. Environment variables: AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY,
//     AWS_REGION, [AWS_SESSION_TOKEN]
//  2. Shared credentials file (defaults to ~/.aws/credentials)
//  3. Amazon EC2 IAM role
//
// If AWS_HOSTED_ZONE_ID is not set, Lego tries to determine the correct public hosted zone via the FQDN.
//
// See also: https://github.com/aws/aws-sdk-go/wiki/configuring-sdk
func NewDNSProvider() (*DNSProvider, error) {
	return NewDNSProviderConfig(NewDefaultConfig())
}

// NewDNSProviderConfig takes a given config and returns a custom configured DNSProvider instance.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("route53: the configuration of the Route53 DNS provider is nil")
	}

	if config.Client != nil {
		return &DNSProvider{client: config.Client, config: config}, nil
	}

	ctx := context.Background()

	cfg, err := createAWSConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	return &DNSProvider{
		client: route53.NewFromConfig(cfg),
		config: config,
	}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	ctx := context.Background()
	info := dns01.GetChallengeInfo(domain, keyAuth)

	hostedZoneID, err := d.getHostedZoneID(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("route53: failed to determine hosted zone ID: %w", err)
	}

	records, err := d.getExistingRecordSets(ctx, hostedZoneID, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("route53: %w", err)
	}

	realValue := `"` + info.Value + `"`

	var found bool
	for _, record := range records {
		if ptr.Deref(record.Value) == realValue {
			found = true
		}
	}

	if !found {
		records = append(records, awstypes.ResourceRecord{Value: aws.String(realValue)})
	}

	recordSet := &awstypes.ResourceRecordSet{
		Name:            aws.String(info.EffectiveFQDN),
		Type:            "TXT",
		TTL:             aws.Int64(int64(d.config.TTL)),
		ResourceRecords: records,
	}

	err = d.changeRecord(ctx, awstypes.ChangeActionUpsert, hostedZoneID, recordSet)
	if err != nil {
		return fmt.Errorf("route53: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	ctx := context.Background()
	info := dns01.GetChallengeInfo(domain, keyAuth)

	hostedZoneID, err := d.getHostedZoneID(ctx, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("failed to determine Route 53 hosted zone ID: %w", err)
	}

	existingRecords, err := d.getExistingRecordSets(ctx, hostedZoneID, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("route53: %w", err)
	}

	if len(existingRecords) == 0 {
		return nil
	}

	var nonLegoRecords []awstypes.ResourceRecord
	for _, record := range existingRecords {
		if ptr.Deref(record.Value) != `"`+info.Value+`"` {
			nonLegoRecords = append(nonLegoRecords, record)
		}
	}

	action := awstypes.ChangeActionUpsert

	recordSet := &awstypes.ResourceRecordSet{
		Name:            aws.String(info.EffectiveFQDN),
		Type:            "TXT",
		TTL:             aws.Int64(int64(d.config.TTL)),
		ResourceRecords: nonLegoRecords,
	}

	// If the records are only records created by lego.
	if len(nonLegoRecords) == 0 {
		action = awstypes.ChangeActionDelete

		recordSet.ResourceRecords = existingRecords
	}

	err = d.changeRecord(ctx, action, hostedZoneID, recordSet)
	if err != nil {
		return fmt.Errorf("route53: %w", err)
	}

	return nil
}

func (d *DNSProvider) changeRecord(ctx context.Context, action awstypes.ChangeAction, hostedZoneID string, recordSet *awstypes.ResourceRecordSet) error {
	recordSetInput := &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: aws.String(hostedZoneID),
		ChangeBatch: &awstypes.ChangeBatch{
			Comment: aws.String("Managed by Lego"),
			Changes: []awstypes.Change{{
				Action:            action,
				ResourceRecordSet: recordSet,
			}},
		},
	}

	resp, err := d.client.ChangeResourceRecordSets(ctx, recordSetInput)
	if err != nil {
		return fmt.Errorf("failed to change record set: %w", err)
	}

	changeID := resp.ChangeInfo.Id

	if d.config.WaitForRecordSetsChanged {
		return wait.For("route53", d.config.PropagationTimeout, d.config.PollingInterval, func() (bool, error) {
			resp, err := d.client.GetChange(ctx, &route53.GetChangeInput{Id: changeID})
			if err != nil {
				return false, fmt.Errorf("failed to query change status: %w", err)
			}

			if resp.ChangeInfo.Status == awstypes.ChangeStatusInsync {
				return true, nil
			}

			return false, fmt.Errorf("unable to retrieve change: ID=%s", ptr.Deref(changeID))
		})
	}

	return nil
}

func (d *DNSProvider) getExistingRecordSets(ctx context.Context, hostedZoneID, fqdn string) ([]awstypes.ResourceRecord, error) {
	listInput := &route53.ListResourceRecordSetsInput{
		HostedZoneId:    aws.String(hostedZoneID),
		StartRecordName: aws.String(fqdn),
		StartRecordType: "TXT",
	}

	recordSetsOutput, err := d.client.ListResourceRecordSets(ctx, listInput)
	if err != nil {
		return nil, err
	}

	if recordSetsOutput == nil {
		return nil, nil
	}

	var records []awstypes.ResourceRecord

	for _, recordSet := range recordSetsOutput.ResourceRecordSets {
		if ptr.Deref(recordSet.Name) == fqdn {
			records = append(records, recordSet.ResourceRecords...)
		}
	}

	return records, nil
}

func (d *DNSProvider) getHostedZoneID(ctx context.Context, fqdn string) (string, error) {
	if d.config.HostedZoneID != "" {
		return d.config.HostedZoneID, nil
	}

	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return "", fmt.Errorf("could not find zone for FQDN %q: %w", fqdn, err)
	}

	// .DNSName should not have a trailing dot
	reqParams := &route53.ListHostedZonesByNameInput{
		DNSName: aws.String(dns01.UnFqdn(authZone)),
	}
	resp, err := d.client.ListHostedZonesByName(ctx, reqParams)
	if err != nil {
		return "", err
	}

	var hostedZoneID string
	for _, hostedZone := range resp.HostedZones {
		// .Name has a trailing dot
		if ptr.Deref(hostedZone.Name) == authZone && d.config.PrivateZone == hostedZone.Config.PrivateZone {
			hostedZoneID = ptr.Deref(hostedZone.Id)
			break
		}
	}

	if hostedZoneID == "" {
		return "", fmt.Errorf("zone %s not found for domain %s", authZone, fqdn)
	}

	hostedZoneID = strings.TrimPrefix(hostedZoneID, "/hostedzone/")

	return hostedZoneID, nil
}

func createAWSConfig(ctx context.Context, config *Config) (aws.Config, error) {
	if err := createAWSConfigCheckParams(config); err != nil {
		return aws.Config{}, err
	}

	optFns := []func(options *awsconfig.LoadOptions) error{
		awsconfig.WithRetryer(func() aws.Retryer {
			return retry.NewStandard(func(options *retry.StandardOptions) {
				options.MaxAttempts = config.MaxRetries

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
	}

	if config.AccessKeyID != "" && config.SecretAccessKey != "" {
		optFns = append(optFns,
			awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(config.AccessKeyID, config.SecretAccessKey, config.SessionToken)),
		)
	}

	if config.Region != "" {
		optFns = append(optFns, awsconfig.WithRegion(config.Region))
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx, optFns...)
	if err != nil {
		return aws.Config{}, err
	}

	if config.AssumeRoleArn != "" {
		cfg.Credentials = stscreds.NewAssumeRoleProvider(sts.NewFromConfig(cfg), config.AssumeRoleArn, func(options *stscreds.AssumeRoleOptions) {
			if config.ExternalID != "" {
				options.ExternalID = &config.ExternalID
			}
		})
	}

	return cfg, nil
}

func createAWSConfigCheckParams(config *Config) error {
	if config == nil {
		return errors.New("config is nil")
	}

	switch {
	case config.SessionToken != "" && config.AccessKeyID == "" && config.SecretAccessKey == "":
		return errors.New("SessionToken must be supplied with AccessKeyID and SecretAccessKey")

	case config.AccessKeyID == "" && config.SecretAccessKey != "" || config.AccessKeyID != "" && config.SecretAccessKey == "":
		return errors.New("AccessKeyID and SecretAccessKey must be supplied together")
	}

	return nil
}
