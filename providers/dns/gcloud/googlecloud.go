// Package gcloud implements a DNS provider for solving the DNS-01 challenge using Google Cloud DNS.
package gcloud

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"cloud.google.com/go/compute/metadata"
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/log"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/go-acme/lego/v4/platform/wait"
	"github.com/miekg/dns"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	gdns "google.golang.org/api/dns/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/impersonate"
	"google.golang.org/api/option"
)

// Environment variables names.
const (
	envNamespace = "GCE_"

	EnvServiceAccount            = envNamespace + "SERVICE_ACCOUNT"
	EnvProject                   = envNamespace + "PROJECT"
	EnvZoneID                    = envNamespace + "ZONE_ID"
	EnvAllowPrivateZone          = envNamespace + "ALLOW_PRIVATE_ZONE"
	EnvDebug                     = envNamespace + "DEBUG"
	EnvImpersonateServiceAccount = envNamespace + "IMPERSONATE_SERVICE_ACCOUNT"

	EnvTTL                = envNamespace + "TTL"
	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
)

const changeStatusDone = "done"

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	Debug                     bool
	Project                   string
	ZoneID                    string
	AllowPrivateZone          bool
	ImpersonateServiceAccount string
	PropagationTimeout        time.Duration
	PollingInterval           time.Duration
	TTL                       int
	HTTPClient                *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		Debug:                     env.GetOrDefaultBool(EnvDebug, false),
		ZoneID:                    env.GetOrDefaultString(EnvZoneID, ""),
		AllowPrivateZone:          env.GetOrDefaultBool(EnvAllowPrivateZone, false),
		ImpersonateServiceAccount: env.GetOrDefaultString(EnvImpersonateServiceAccount, ""),
		TTL:                       env.GetOrDefaultInt(EnvTTL, dns01.DefaultTTL),
		PropagationTimeout:        env.GetOrDefaultSecond(EnvPropagationTimeout, 180*time.Second),
		PollingInterval:           env.GetOrDefaultSecond(EnvPollingInterval, 5*time.Second),
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
	client *gdns.Service
}

// NewDNSProvider returns a DNSProvider instance configured for Google Cloud DNS.
// By default, the project name is auto-detected by using the metadata service,
// it can be overridden using the GCE_PROJECT environment variable.
// A Service Account can be passed in the environment variable: GCE_SERVICE_ACCOUNT
// or by specifying the keyfile location: GCE_SERVICE_ACCOUNT_FILE.
func NewDNSProvider() (*DNSProvider, error) {
	// Use a service account file if specified via environment variable.
	if saKey := env.GetOrFile(EnvServiceAccount); saKey != "" {
		return NewDNSProviderServiceAccountKey([]byte(saKey))
	}

	// Use default credentials.
	project := env.GetOrDefaultString(EnvProject, autodetectProjectID(context.Background()))
	return NewDNSProviderCredentials(project)
}

// NewDNSProviderCredentials uses the supplied credentials
// to return a DNSProvider instance configured for Google Cloud DNS.
func NewDNSProviderCredentials(project string) (*DNSProvider, error) {
	if project == "" {
		return nil, errors.New("googlecloud: project name missing")
	}

	config := NewDefaultConfig()
	config.Project = project

	var err error
	config.HTTPClient, err = newClientFromCredentials(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("googlecloud: %w", err)
	}

	return NewDNSProviderConfig(config)
}

// NewDNSProviderServiceAccountKey uses the supplied service account JSON
// to return a DNSProvider instance configured for Google Cloud DNS.
func NewDNSProviderServiceAccountKey(saKey []byte) (*DNSProvider, error) {
	if len(saKey) == 0 {
		return nil, errors.New("googlecloud: Service Account is missing")
	}

	// If GCE_PROJECT is non-empty it overrides the project in the service
	// account file.
	project := env.GetOrDefaultString(EnvProject, "")
	if project == "" {
		// read project id from service account file
		var datJSON struct {
			ProjectID string `json:"project_id"`
		}
		err := json.Unmarshal(saKey, &datJSON)
		if err != nil || datJSON.ProjectID == "" {
			return nil, errors.New("googlecloud: project ID not found in Google Cloud Service Account file")
		}
		project = datJSON.ProjectID
	}

	config := NewDefaultConfig()
	config.Project = project

	var err error
	config.HTTPClient, err = newClientFromServiceAccountKey(context.Background(), config, saKey)
	if err != nil {
		return nil, fmt.Errorf("googlecloud: %w", err)
	}

	return NewDNSProviderConfig(config)
}

// NewDNSProviderServiceAccount uses the supplied service account JSON file
// to return a DNSProvider instance configured for Google Cloud DNS.
func NewDNSProviderServiceAccount(saFile string) (*DNSProvider, error) {
	if saFile == "" {
		return nil, errors.New("googlecloud: Service Account file missing")
	}

	saKey, err := os.ReadFile(saFile)
	if err != nil {
		return nil, fmt.Errorf("googlecloud: unable to read Service Account file: %w", err)
	}

	return NewDNSProviderServiceAccountKey(saKey)
}

// NewDNSProviderConfig return a DNSProvider instance configured for Google Cloud DNS.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("googlecloud: the configuration of the DNS provider is nil")
	}
	if config.HTTPClient == nil {
		return nil, errors.New("googlecloud: unable to create Google Cloud DNS service: client is nil")
	}

	svc, err := gdns.NewService(context.Background(), option.WithHTTPClient(config.HTTPClient))
	if err != nil {
		return nil, fmt.Errorf("googlecloud: unable to create Google Cloud DNS service: %w", err)
	}

	return &DNSProvider{config: config, client: svc}, nil
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	zone, err := d.getHostedZone(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("googlecloud: %w", err)
	}

	// Look for existing records.
	existingRrSet, err := d.findTxtRecords(zone, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("googlecloud: %w", err)
	}

	for _, rrSet := range existingRrSet {
		var rrd []string
		for _, rr := range rrSet.Rrdatas {
			data := mustUnquote(rr)
			rrd = append(rrd, data)

			if data == info.Value {
				log.Printf("skip: the record already exists: %s", info.Value)
				return nil
			}
		}
		rrSet.Rrdatas = rrd
	}

	// Attempt to delete the existing records before adding the new one.
	if len(existingRrSet) > 0 {
		if err = d.applyChanges(zone, &gdns.Change{Deletions: existingRrSet}); err != nil {
			return fmt.Errorf("googlecloud: %w", err)
		}
	}

	rec := &gdns.ResourceRecordSet{
		Name:    info.EffectiveFQDN,
		Rrdatas: []string{info.Value},
		Ttl:     int64(d.config.TTL),
		Type:    "TXT",
	}

	// Append existing TXT record data to the new TXT record data
	for _, rrSet := range existingRrSet {
		for _, rr := range rrSet.Rrdatas {
			if rr != info.Value {
				rec.Rrdatas = append(rec.Rrdatas, rr)
			}
		}
	}

	change := &gdns.Change{
		Additions: []*gdns.ResourceRecordSet{rec},
	}

	if err = d.applyChanges(zone, change); err != nil {
		return fmt.Errorf("googlecloud: %w", err)
	}

	return nil
}

func (d *DNSProvider) applyChanges(zone string, change *gdns.Change) error {
	if d.config.Debug {
		data, _ := json.Marshal(change)
		log.Printf("change (Create): %s", string(data))
	}

	chg, err := d.client.Changes.Create(d.config.Project, zone, change).Do()
	if err != nil {
		var v *googleapi.Error
		if errors.As(err, &v) && v.Code == http.StatusNotFound {
			return nil
		}

		data, _ := json.Marshal(change)
		return fmt.Errorf("failed to perform changes [zone %s, change %s]: %w", zone, string(data), err)
	}

	if chg.Status == changeStatusDone {
		return nil
	}

	chgID := chg.Id

	// wait for change to be acknowledged
	return wait.For("apply change", 30*time.Second, 3*time.Second, func() (bool, error) {
		if d.config.Debug {
			data, _ := json.Marshal(change)
			log.Printf("change (Get): %s", string(data))
		}

		chg, err = d.client.Changes.Get(d.config.Project, zone, chgID).Do()
		if err != nil {
			data, _ := json.Marshal(change)
			return false, fmt.Errorf("failed to get changes [zone %s, change %s]: %w", zone, string(data), err)
		}

		if chg.Status == changeStatusDone {
			return true, nil
		}

		return false, fmt.Errorf("status: %s", chg.Status)
	})
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	info := dns01.GetChallengeInfo(domain, keyAuth)

	zone, err := d.getHostedZone(info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("googlecloud: %w", err)
	}

	records, err := d.findTxtRecords(zone, info.EffectiveFQDN)
	if err != nil {
		return fmt.Errorf("googlecloud: %w", err)
	}

	if len(records) == 0 {
		return nil
	}

	_, err = d.client.Changes.Create(d.config.Project, zone, &gdns.Change{Deletions: records}).Do()
	if err != nil {
		return fmt.Errorf("googlecloud: %w", err)
	}
	return nil
}

// Timeout customizes the timeout values used by the ACME package for checking
// DNS record validity.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// getHostedZone returns the managed-zone.
func (d *DNSProvider) getHostedZone(domain string) (string, error) {
	authZone, zones, err := d.lookupHostedZoneID(domain)
	if err != nil {
		return "", err
	}

	if len(zones) == 0 {
		return "", fmt.Errorf("no matching domain found for domain %s", authZone)
	}

	for _, z := range zones {
		if z.Visibility == "public" || z.Visibility == "" || (z.Visibility == "private" && d.config.AllowPrivateZone) {
			return z.Name, nil
		}
	}

	if d.config.AllowPrivateZone {
		return "", fmt.Errorf("no public or private zone found for domain %s", authZone)
	}

	return "", fmt.Errorf("no public zone found for domain %s", authZone)
}

// lookupHostedZoneID finds the managed zone ID in Google.
//
// Be careful here.
// An automated system might run in a GCloud Service Account, with access to edit the zone
//
//	(gcloud dns managed-zones get-iam-policy $zone_id) (role roles/dns.admin)
//
// but not with project-wide access to list all zones
//
//	(gcloud projects get-iam-policy $project_id) (a role with permission dns.managedZones.list)
//
// If we force a zone list to succeed, we demand more permissions than needed.
func (d *DNSProvider) lookupHostedZoneID(domain string) (string, []*gdns.ManagedZone, error) {
	// GCE_ZONE_ID override for service accounts to avoid needing zones-list permission
	if d.config.ZoneID != "" {
		zone, err := d.client.ManagedZones.Get(d.config.Project, d.config.ZoneID).Do()
		if err != nil {
			return "", nil, fmt.Errorf("API call ManagedZones.Get for explicit zone ID %q in project %q failed: %w", d.config.ZoneID, d.config.Project, err)
		}

		return zone.DnsName, []*gdns.ManagedZone{zone}, nil
	}

	authZone, err := dns01.FindZoneByFqdn(dns.Fqdn(domain))
	if err != nil {
		return "", nil, fmt.Errorf("could not find zone: %w", err)
	}

	zones, err := d.client.ManagedZones.
		List(d.config.Project).
		DnsName(authZone).
		Do()
	if err != nil {
		return "", nil, fmt.Errorf("API call ManagedZones.List failed: %w", err)
	}

	return authZone, zones.ManagedZones, nil
}

func (d *DNSProvider) findTxtRecords(zone, fqdn string) ([]*gdns.ResourceRecordSet, error) {
	recs, err := d.client.ResourceRecordSets.List(d.config.Project, zone).Name(fqdn).Type("TXT").Do()
	if err != nil {
		return nil, err
	}

	return recs.Rrsets, nil
}

func newClientFromCredentials(ctx context.Context, config *Config) (*http.Client, error) {
	if config.ImpersonateServiceAccount != "" {
		ts, err := google.DefaultTokenSource(ctx, "https://www.googleapis.com/auth/cloud-platform")
		if err != nil {
			return nil, fmt.Errorf("unable to get default token source: %w", err)
		}

		return newImpersonateClient(ctx, config.ImpersonateServiceAccount, ts)
	}

	client, err := google.DefaultClient(ctx, gdns.NdevClouddnsReadwriteScope)
	if err != nil {
		return nil, fmt.Errorf("unable to get Google Cloud client: %w", err)
	}

	return client, nil
}

func newClientFromServiceAccountKey(ctx context.Context, config *Config, saKey []byte) (*http.Client, error) {
	if config.ImpersonateServiceAccount != "" {
		conf, err := google.JWTConfigFromJSON(saKey, "https://www.googleapis.com/auth/cloud-platform")
		if err != nil {
			return nil, fmt.Errorf("unable to acquire config: %w", err)
		}

		return newImpersonateClient(ctx, config.ImpersonateServiceAccount, conf.TokenSource(ctx))
	}

	conf, err := google.JWTConfigFromJSON(saKey, gdns.NdevClouddnsReadwriteScope)
	if err != nil {
		return nil, fmt.Errorf("unable to acquire config: %w", err)
	}

	return conf.Client(ctx), nil
}

func newImpersonateClient(ctx context.Context, impersonateServiceAccount string, ts oauth2.TokenSource) (*http.Client, error) {
	impersonatedTS, err := impersonate.CredentialsTokenSource(ctx, impersonate.CredentialsConfig{
		TargetPrincipal: impersonateServiceAccount,
		Scopes:          []string{gdns.NdevClouddnsReadwriteScope},
	}, option.WithTokenSource(ts))
	if err != nil {
		return nil, fmt.Errorf("unable to create impersonated credentials: %w", err)
	}

	return oauth2.NewClient(ctx, impersonatedTS), nil
}

func mustUnquote(raw string) string {
	clean, err := strconv.Unquote(raw)
	if err != nil {
		return raw
	}
	return clean
}

func autodetectProjectID(ctx context.Context) string {
	if pid, err := metadata.ProjectIDWithContext(ctx); err == nil {
		return pid
	}

	return ""
}
