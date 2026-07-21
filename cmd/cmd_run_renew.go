package cmd

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"math/rand/v2"
	"os"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/go-acme/lego/v5/acme/api"
	"github.com/go-acme/lego/v5/certcrypto"
	"github.com/go-acme/lego/v5/cmd/internal/flags"
	"github.com/go-acme/lego/v5/cmd/internal/hook"
	"github.com/go-acme/lego/v5/cmd/internal/storage"
	"github.com/go-acme/lego/v5/lego"
	"github.com/go-acme/lego/v5/log"
	"github.com/mattn/go-isatty"
	"github.com/urfave/cli/v3"
)

const noDays = -math.MaxInt

type renewProcessor struct {
	cmd *cli.Command

	lazyClient lzSetUp

	certsStorage *storage.CertificatesStorage
	hookManager  *hook.Manager
}

func (p *renewProcessor) renew(ctx context.Context, certID string, resource *storage.Certificate) error {
	changed := hasChanged(resource, p.cmd)

	if p.cmd.IsSet(flags.FlgCSR) {
		return p.renewForCSR(ctx, certID, changed)
	}

	domains := p.cmd.StringSlice(flags.FlgDomains)
	if len(domains) == 0 {
		domains = resource.Domains
	}

	return p.renewForDomains(ctx, certID, domains, changed)
}

func (p *renewProcessor) renewForDomains(ctx context.Context, certID string, domains []string, changed bool) error {
	certificates, err := p.certsStorage.ReadCertificate(certID)
	if err != nil {
		return fmt.Errorf("error while reading the certificate for %q: %w", certID, err)
	}

	cert := certificates[0]

	if cert.IsCA {
		return fmt.Errorf("certificate bundle for %q starts with a CA certificate", certID)
	}

	certDomains := certcrypto.ExtractDomains(cert)

	renewalDomains := slices.Clone(domains)
	if !p.cmd.Bool(flags.FlgForceCertDomains) {
		renewalDomains = merge(certDomains, domains)
	}

	var replacesCertID string

	// If the domains are different, this must create a new certificate, then the renewal constraints must be skipped.
	if !changed && sameDomains(certDomains, renewalDomains) {
		var ariRenewalTime *time.Time

		ariRenewalTime, replacesCertID, err = p.getARIInfo(ctx, certID, cert)
		if err != nil {
			return err
		}

		if ariRenewalTime == nil && !p.cmd.Bool(flags.FlgRenewForce) && !isInRenewalPeriod(cert, certID, getFlagRenewDays(p.cmd), time.Now()) {
			return nil
		}
	} else {
		log.Info("Changes detected in the certificate configuration.")
	}

	// This is just meant to be informal for the user.
	log.Info("Trying renewal.",
		log.CertNameAttr(certID),
		log.DurationAttr("time-remaining", cert.NotAfter.Sub(time.Now().UTC())),
	)

	request, err := newObtainRequest(p.cmd, renewalDomains)
	if err != nil {
		return err
	}

	if p.cmd.Bool(flags.FlgReuseKey) {
		request.PrivateKey, err = p.certsStorage.ReadPrivateKey(certID)
		if err != nil {
			return err
		}
	}

	if replacesCertID != "" {
		request.ReplacesCertID = replacesCertID
	}

	err = p.hookManager.PreForDomains(ctx, certID, request)
	if err != nil {
		return fmt.Errorf("pre-renew hook: %w", err)
	}

	defer func() { _ = p.hookManager.Post(ctx) }()

	client, err := p.lazyClient()
	if err != nil {
		return fmt.Errorf("set up client: %w", err)
	}

	// If the domains are different, this must create a new certificate, then the renewal constraints must be skipped.
	if !changed && sameDomains(certDomains, renewalDomains) {
		randomSleep(p.cmd)
	}

	certRes, err := client.Certificate.Obtain(ctx, request)
	if err != nil {
		return fmt.Errorf("could not obtain the certificate for %q: %w", certID, err)
	}

	certRes.ID = certID

	options := newSaveOptions(p.cmd)

	err = p.certsStorage.Save(
		&storage.Certificate{
			Resource: certRes,
			Origin:   storage.OriginCommand,
		},
		options,
	)
	if err != nil {
		return fmt.Errorf("could not save the resource: %w", err)
	}

	return p.hookManager.Deploy(ctx, certRes, options)
}

func (p *renewProcessor) renewForCSR(ctx context.Context, certID string, changed bool) error {
	csr, err := storage.ReadCSRFile(p.cmd.String(flags.FlgCSR))
	if err != nil {
		return fmt.Errorf("CSR: could not read file %q: %w", p.cmd.String(flags.FlgCSR), err)
	}

	certificates, err := p.certsStorage.ReadCertificate(certID)
	if err != nil {
		return fmt.Errorf("CSR: error while reading the certificate for domains %q: %w",
			strings.Join(certcrypto.ExtractDomainsCSR(csr), ","), err)
	}

	cert := certificates[0]

	if cert.IsCA {
		return fmt.Errorf("CSR: certificate bundle for %q starts with a CA certificate", certID)
	}

	var replacesCertID string

	// If the domains are different, this must create a new certificate, then the renewal constraints must be skipped.
	if !changed && sameDomainsCertificate(cert, csr) {
		var ariRenewalTime *time.Time

		ariRenewalTime, replacesCertID, err = p.getARIInfo(ctx, certID, cert)
		if err != nil {
			return fmt.Errorf("CSR: %w", err)
		}

		if ariRenewalTime == nil && !p.cmd.Bool(flags.FlgRenewForce) && !isInRenewalPeriod(cert, certID, getFlagRenewDays(p.cmd), time.Now()) {
			return nil
		}
	} else {
		log.Info("Changes detected in the certificate configuration.")
	}

	// This is just meant to be informal for the user.
	log.Info("Trying renewal.",
		log.CertNameAttr(certID),
		log.DurationAttr("time-remaining", cert.NotAfter.Sub(time.Now().UTC())),
	)

	request := newObtainForCSRRequest(p.cmd, csr)

	if replacesCertID != "" {
		request.ReplacesCertID = replacesCertID
	}

	err = p.hookManager.PreForCSR(ctx, certID, request)
	if err != nil {
		return fmt.Errorf("CSR: pre-renew hook: %w", err)
	}

	defer func() { _ = p.hookManager.Post(ctx) }()

	client, err := p.lazyClient()
	if err != nil {
		return fmt.Errorf("CSR: set up client: %w", err)
	}

	certRes, err := client.Certificate.ObtainForCSR(ctx, request)
	if err != nil {
		return fmt.Errorf("CSR: could not obtain the certificate: %w", err)
	}

	certRes.ID = certID

	options := newSaveOptions(p.cmd)

	err = p.certsStorage.Save(
		&storage.Certificate{
			Resource: certRes,
			Origin:   storage.OriginCommand,
		},
		options,
	)
	if err != nil {
		return fmt.Errorf("CSR: could not save the resource: %w", err)
	}

	return p.hookManager.Deploy(ctx, certRes, options)
}

func (p *renewProcessor) getARIInfo(ctx context.Context, certID string, cert *x509.Certificate) (*time.Time, string, error) {
	if p.cmd.Bool(flags.FlgARIDisable) {
		return nil, "", nil
	}

	client, err := p.lazyClient()
	if err != nil {
		return nil, "", fmt.Errorf("set up client: %w", err)
	}

	willingToSleep := p.cmd.Duration(flags.FlgARIWaitToRenewDuration)

	ariRenewalTime := getARIRenewalTime(ctx, willingToSleep, cert, certID, client)
	if ariRenewalTime != nil {
		now := time.Now().UTC()

		// Figure out if we need to sleep before renewing.
		if ariRenewalTime.After(now) {
			log.Info("Sleeping until renewal time",
				log.CertNameAttr(certID),
				slog.Duration("sleep", ariRenewalTime.Sub(now)),
				slog.Time("renewalTime", *ariRenewalTime),
			)

			time.Sleep(ariRenewalTime.Sub(now))
		}
	}

	replacesCertID, err := api.MakeARICertID(cert)
	if err != nil {
		return nil, "", fmt.Errorf("error while constructing the ARI CertID for domain %q: %w", certID, err)
	}

	return ariRenewalTime, replacesCertID, nil
}

func getFlagRenewDays(cmd *cli.Command) int {
	if cmd.IsSet(flags.FlgRenewDays) {
		return cmd.Int(flags.FlgRenewDays)
	}

	return noDays
}

func isInRenewalPeriod(cert *x509.Certificate, certID string, days int, now time.Time) bool {
	dueDate := getDueDate(cert, days, now)

	if dueDate.Before(now) || dueDate.Equal(now) {
		return true
	}

	log.Infof(
		log.LazySprintf("Skip renewal: The certificate expires at %s, the renewal can be performed in %s.",
			cert.NotAfter.Format(time.RFC3339),
			log.FormattableDuration(dueDate.Sub(now)),
		),
		log.CertNameAttr(certID),
	)

	return false
}

func getDueDate(x509Cert *x509.Certificate, days int, now time.Time) time.Time {
	if days == noDays {
		lifetime := x509Cert.NotAfter.Sub(x509Cert.NotBefore)

		var divisor int64 = 3
		if lifetime.Round(24*time.Hour).Hours()/24.0 <= 10 {
			divisor = 2
		}

		return x509Cert.NotAfter.Add(-1 * time.Duration(lifetime.Nanoseconds()/divisor))
	}

	if days < 0 {
		// if the number of days is negative: always renew the certificate.
		return now
	}

	return x509Cert.NotAfter.Add(-1 * time.Duration(days) * 24 * time.Hour)
}

// getARIRenewalTime checks if the certificate needs to be renewed using the renewalInfo endpoint.
func getARIRenewalTime(ctx context.Context, willingToSleep time.Duration, cert *x509.Certificate, certID string, client *lego.Client) *time.Time {
	renewalInfo, err := client.Certificate.GetRenewalInfo(ctx, cert)
	if err != nil {
		if errors.Is(err, api.ErrNoARI) {
			log.Warn("The server does not advertise a renewal info endpoint.",
				log.CertNameAttr(certID),
				log.ErrorAttr(err),
			)

			return nil
		}

		log.Warn("Calling renewal info endpoint",
			log.CertNameAttr(certID),
			log.ErrorAttr(err),
		)

		return nil
	}

	now := time.Now().UTC()

	renewalTime := renewalInfo.ShouldRenewAt(now, willingToSleep)
	if renewalTime == nil {
		log.Debug("RenewalInfo endpoint indicates that renewal is not needed.", log.CertNameAttr(certID))
		return nil
	}

	log.Debug("RenewalInfo endpoint indicates that renewal is needed.", log.CertNameAttr(certID))

	if renewalInfo.ExplanationURL != "" {
		log.Info("RenewalInfo endpoint provided an explanation.",
			log.CertNameAttr(certID),
			slog.String("explanationURL", renewalInfo.ExplanationURL),
		)
	}

	return renewalTime
}

func randomSleep(cmd *cli.Command) {
	// https://github.com/go-acme/lego/issues/1656
	// https://github.com/certbot/certbot/blob/284023a1b7672be2bd4018dd7623b3b92197d4b0/certbot/certbot/_internal/renewal.py#L435-L440
	if !isatty.IsTerminal(os.Stdout.Fd()) && !cmd.Bool(flags.FlgNoRandomSleep) {
		// https://github.com/certbot/certbot/blob/284023a1b7672be2bd4018dd7623b3b92197d4b0/certbot/certbot/_internal/renewal.py#L472
		const jitter = 8 * time.Minute

		sleepTime := time.Duration(rand.Int64N(int64(jitter)))

		log.Info("renewal: random delay.", slog.Duration("sleep", sleepTime))
		time.Sleep(sleepTime)
	}
}

func merge(prevDomains, nextDomains []string) []string {
	for _, next := range nextDomains {
		if slices.Contains(prevDomains, next) {
			continue
		}

		prevDomains = append(prevDomains, next)
	}

	return prevDomains
}

func sameDomainsCertificate(cert *x509.Certificate, csr *x509.CertificateRequest) bool {
	return sameDomains(certcrypto.ExtractDomains(cert), certcrypto.ExtractDomainsCSR(csr))
}

func sameDomains(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	aClone := slices.Clone(a)
	sort.Strings(aClone)

	bClone := slices.Clone(b)
	sort.Strings(bClone)

	return slices.Equal(aClone, bClone)
}

func hasChanged(resource *storage.Certificate, cmd *cli.Command) bool {
	return resource.Profile != cmd.String(flags.FlgProfile)
}
