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
	"sync"
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

func createRenew() *cli.Command {
	return &cli.Command{
		Name:   "renew",
		Usage:  "Renew a certificate",
		Action: renew,
		Before: flags.RenewFlagsValidation,
		Flags:  flags.CreateRenewFlags(),
	}
}

func renew(ctx context.Context, cmd *cli.Command) error {
	keyType, err := certcrypto.ToKeyType(cmd.String(flags.FlgKeyType))
	if err != nil {
		return err
	}

	accountsStorage := storage.NewAccountsStorage(cmd.String(flags.FlgPath))

	account, err := accountsStorage.Get(cmd.String(flags.FlgServer), keyType, cmd.String(flags.FlgEmail), cmd.String(flags.FlgAccountID))
	if err != nil {
		return fmt.Errorf("set up account: %w", err)
	}

	lazyClient := sync.OnceValues(func() (*lego.Client, error) {
		client, errC := newClient(cmd, account)
		if errC != nil {
			return nil, fmt.Errorf("new client: %w", errC)
		}

		errC = setupChallenges(cmd, client)
		if errC != nil {
			return nil, fmt.Errorf("setup challenges: %w", errC)
		}

		return client, nil
	})

	err = handleRegistration(ctx, cmd, lazyClient, accountsStorage, account, false)
	if err != nil {
		return fmt.Errorf("registration: %w", err)
	}

	certsStorage := storage.NewCertificatesStorage(cmd.String(flags.FlgPath))

	hookManager := newHookManager(cmd, certsStorage, account)

	// CSR
	if cmd.IsSet(flags.FlgCSR) {
		return renewForCSR(ctx, cmd, lazyClient, certsStorage, hookManager)
	}

	// Domains
	return renewForDomains(ctx, cmd, lazyClient, certsStorage, hookManager)
}

func renewForDomains(ctx context.Context, cmd *cli.Command, lazyClient lzSetUp, certsStorage *storage.CertificatesStorage, hookManager *hook.Manager) error {
	domains := cmd.StringSlice(flags.FlgDomains)

	certID := cmd.String(flags.FlgCertName)

	switch {
	case certID == "" && len(domains) > 0:
		certID = domains[0]

	case certID != "" && len(domains) == 0:
		resource, err := certsStorage.ReadResource(certID)
		if err != nil {
			return fmt.Errorf("error while reading resource for %q: %w", certID, err)
		}

		domains = resource.Domains

	case certID != "" && len(domains) > 0:
		// Nothing to do, certID and domains are already consistent.

	default:
		return errors.New("no domains or certificate ID/name provided")
	}

	// load the cert resource from files.
	// We store the certificate, private key and metadata in different files
	// as web servers would not be able to work with a combined file.
	certificates, err := certsStorage.ReadCertificate(certID)
	if err != nil {
		return fmt.Errorf("error while reading the certificate for %q: %w", certID, err)
	}

	cert := certificates[0]

	if cert.IsCA {
		return fmt.Errorf("certificate bundle for %q starts with a CA certificate", certID)
	}

	ariRenewalTime, replacesCertID, err := getARIInfo(ctx, cmd, lazyClient, certID, cert)
	if err != nil {
		return err
	}

	certDomains := certcrypto.ExtractDomains(cert)

	renewalDomains := slices.Clone(domains)
	if !cmd.Bool(flags.FlgForceCertDomains) {
		renewalDomains = merge(certDomains, domains)
	}

	if ariRenewalTime == nil && !cmd.Bool(flags.FlgRenewForce) && sameDomains(certDomains, renewalDomains) &&
		!isInRenewalPeriod(cert, certID, getFlagRenewDays(cmd), time.Now()) {
		return nil
	}

	// This is just meant to be informal for the user.
	log.Info("Trying renewal.",
		log.CertNameAttr(certID),
		slog.Any("time-remaining", log.FormattableDuration(cert.NotAfter.Sub(time.Now().UTC()))),
	)

	err = hookManager.Pre(ctx, certID, renewalDomains)
	if err != nil {
		return fmt.Errorf("pre-renew hook: %w", err)
	}

	defer func() { _ = hookManager.Post(ctx) }()

	client, err := lazyClient()
	if err != nil {
		return fmt.Errorf("set up client: %w", err)
	}

	randomSleep(cmd)

	request, err := newObtainRequest(cmd, renewalDomains)
	if err != nil {
		return err
	}

	if cmd.Bool(flags.FlgReuseKey) {
		request.PrivateKey, err = certsStorage.ReadPrivateKey(certID)
		if err != nil {
			return err
		}
	}

	if replacesCertID != "" {
		request.ReplacesCertID = replacesCertID
	}

	certRes, err := client.Certificate.Obtain(ctx, request)
	if err != nil {
		return fmt.Errorf("could not obtain the certificate for %q: %w", certID, err)
	}

	certRes.ID = certID

	options := newSaveOptions(cmd)

	err = certsStorage.Save(certRes, options)
	if err != nil {
		return fmt.Errorf("could not save the resource: %w", err)
	}

	return hookManager.Deploy(ctx, certRes, options)
}

func renewForCSR(ctx context.Context, cmd *cli.Command, lazyClient lzSetUp, certsStorage *storage.CertificatesStorage, hookManager *hook.Manager) error {
	csr, err := storage.ReadCSRFile(cmd.String(flags.FlgCSR))
	if err != nil {
		return fmt.Errorf("CSR: could not read file %q: %w", cmd.String(flags.FlgCSR), err)
	}

	certID := cmd.String(flags.FlgCertName)
	if certID == "" {
		certID, err = certcrypto.GetCSRMainDomain(csr)
		if err != nil {
			return fmt.Errorf("CSR: could not get the main domain: %w", err)
		}
	}

	// load the cert resource from files.
	// We store the certificate, private key and metadata in different files
	// as web servers would not be able to work with a combined file.
	certificates, err := certsStorage.ReadCertificate(certID)
	if err != nil {
		return fmt.Errorf("CSR: error while reading the certificate for domains %q: %w",
			strings.Join(certcrypto.ExtractDomainsCSR(csr), ","), err)
	}

	cert := certificates[0]

	if cert.IsCA {
		return fmt.Errorf("CSR: certificate bundle for %q starts with a CA certificate", certID)
	}

	ariRenewalTime, replacesCertID, err := getARIInfo(ctx, cmd, lazyClient, certID, cert)
	if err != nil {
		return fmt.Errorf("CSR: %w", err)
	}

	if ariRenewalTime == nil && !cmd.Bool(flags.FlgRenewForce) && sameDomainsCertificate(cert, csr) &&
		!isInRenewalPeriod(cert, certID, getFlagRenewDays(cmd), time.Now()) {
		return nil
	}

	// This is just meant to be informal for the user.
	log.Info("Trying renewal.",
		log.CertNameAttr(certID),
		slog.Any("time-remaining", log.FormattableDuration(cert.NotAfter.Sub(time.Now().UTC()))),
	)

	err = hookManager.Pre(ctx, certID, certcrypto.ExtractDomainsCSR(csr))
	if err != nil {
		return fmt.Errorf("CSR: pre-renew hook: %w", err)
	}

	defer func() { _ = hookManager.Post(ctx) }()

	client, err := lazyClient()
	if err != nil {
		return fmt.Errorf("CSR: set up client: %w", err)
	}

	request := newObtainForCSRRequest(cmd, csr)

	if replacesCertID != "" {
		request.ReplacesCertID = replacesCertID
	}

	certRes, err := client.Certificate.ObtainForCSR(ctx, request)
	if err != nil {
		return fmt.Errorf("CSR: could not obtain the certificate: %w", err)
	}

	certRes.ID = certID

	options := newSaveOptions(cmd)

	err = certsStorage.Save(certRes, options)
	if err != nil {
		return fmt.Errorf("CSR: could not save the resource: %w", err)
	}

	return hookManager.Deploy(ctx, certRes, options)
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

func getARIInfo(ctx context.Context, cmd *cli.Command, lazyClient lzSetUp, certID string, cert *x509.Certificate) (*time.Time, string, error) {
	if cmd.Bool(flags.FlgARIDisable) {
		return nil, "", nil
	}

	client, err := lazyClient()
	if err != nil {
		return nil, "", fmt.Errorf("set up client: %w", err)
	}

	willingToSleep := cmd.Duration(flags.FlgARIWaitToRenewDuration)

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
		log.Info("RenewalInfo endpoint indicates that renewal is not needed.", log.CertNameAttr(certID))
		return nil
	}

	log.Info("RenewalInfo endpoint indicates that renewal is needed.", log.CertNameAttr(certID))

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
