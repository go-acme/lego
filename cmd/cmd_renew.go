package cmd

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"math/rand"
	"os"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/go-acme/lego/v5/acme/api"
	"github.com/go-acme/lego/v5/certcrypto"
	"github.com/go-acme/lego/v5/certificate"
	"github.com/go-acme/lego/v5/cmd/internal/hook"
	"github.com/go-acme/lego/v5/cmd/internal/storage"
	"github.com/go-acme/lego/v5/lego"
	"github.com/go-acme/lego/v5/log"
	"github.com/mattn/go-isatty"
	"github.com/urfave/cli/v3"
)

const noDays = -math.MaxInt

type lzSetUp func() (*lego.Client, error)

func createRenew() *cli.Command {
	return &cli.Command{
		Name:   "renew",
		Usage:  "Renew a certificate",
		Action: renew,
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			hasDomains := len(cmd.StringSlice(flgDomains)) > 0
			hasCsr := cmd.String(flgCSR) != ""
			hasCertID := cmd.String(flgCertName) != ""

			if hasDomains && hasCsr {
				log.Fatal(fmt.Sprintf("Please specify either --%s/-d or --%s, but not both", flgDomains, flgCSR))
			}

			if !hasCertID && !hasDomains && !hasCsr {
				log.Fatal(fmt.Sprintf("Please specify --%s or --%s/-d (or --%s if you already have a CSR)", flgCertName, flgDomains, flgCSR))
			}

			if cmd.Bool(flgForceCertDomains) && hasCsr {
				log.Fatal(fmt.Sprintf("--%s only works with --%s/-d, --%s doesn't support this option.", flgForceCertDomains, flgDomains, flgCSR))
			}

			return ctx, validateNetworkStack(cmd)
		},
		Flags: createRenewFlags(),
	}
}

func renew(ctx context.Context, cmd *cli.Command) error {
	keyType, err := certcrypto.GetKeyType(cmd.String(flgKeyType))
	if err != nil {
		return fmt.Errorf("get the key type: %w", err)
	}

	accountsStorage, err := storage.NewAccountsStorage(newAccountsStorageConfig(cmd))
	if err != nil {
		return fmt.Errorf("accounts storage initialization: %w", err)
	}

	account, err := accountsStorage.Get(ctx, keyType, cmd.String(flgEmail), cmd.String(flgAccountID))
	if err != nil {
		return fmt.Errorf("set up account: %w", err)
	}

	if account.Registration == nil {
		return fmt.Errorf("the account %s is not registered", account.GetID())
	}

	certsStorage := storage.NewCertificatesStorage(cmd.String(flgPath))

	meta := map[string]string{
		// TODO(ldez) add account ID.
		hook.EnvAccountEmail: account.Email,
	}

	lazyClient := sync.OnceValues(func() (*lego.Client, error) {
		client, err := newClient(cmd, account, keyType)
		if err != nil {
			return nil, fmt.Errorf("new client: %w", err)
		}

		setupChallenges(cmd, client)

		return client, nil
	})

	// CSR
	if cmd.IsSet(flgCSR) {
		return renewForCSR(ctx, cmd, lazyClient, certsStorage, meta)
	}

	// Domains
	return renewForDomains(ctx, cmd, lazyClient, certsStorage, meta)
}

func renewForDomains(ctx context.Context, cmd *cli.Command, lazyClient lzSetUp, certsStorage *storage.CertificatesStorage, meta map[string]string) error {
	domains := cmd.StringSlice(flgDomains)

	certID := cmd.String(flgCertName)

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
	if !cmd.Bool(flgForceCertDomains) {
		renewalDomains = merge(certDomains, domains)
	}

	if ariRenewalTime == nil && !cmd.Bool(flgRenewForce) && sameDomains(certDomains, renewalDomains) &&
		!isInRenewalPeriod(cert, certID, getFlagRenewDays(cmd), time.Now()) {
		return nil
	}

	// This is just meant to be informal for the user.
	log.Info("acme: Trying renewal.",
		log.CertNameAttr(certID),
		slog.Any("time-remaining", FormattableDuration(cert.NotAfter.Sub(time.Now().UTC()))),
	)

	client, err := lazyClient()
	if err != nil {
		return fmt.Errorf("set up client: %w", err)
	}

	randomSleep(cmd)

	request := newObtainRequest(cmd, renewalDomains)

	if cmd.Bool(flgReuseKey) {
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

	hook.AddPathToMetadata(meta, certRes, certsStorage, options)

	return hook.Launch(ctx, cmd.String(flgDeployHook), cmd.Duration(flgDeployHookTimeout), meta)
}

func renewForCSR(ctx context.Context, cmd *cli.Command, lazyClient lzSetUp, certsStorage *storage.CertificatesStorage, meta map[string]string) error {
	csr, err := readCSRFile(cmd.String(flgCSR))
	if err != nil {
		return fmt.Errorf("could not read CSR file %q: %w", cmd.String(flgCSR), err)
	}

	certID := cmd.String(flgCertName)
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
		return fmt.Errorf("certificate bundle for %q starts with a CA certificate", certID)
	}

	ariRenewalTime, replacesCertID, err := getARIInfo(ctx, cmd, lazyClient, certID, cert)
	if err != nil {
		return fmt.Errorf("CSR: %w", err)
	}

	if ariRenewalTime == nil && !cmd.Bool(flgRenewForce) && sameDomainsCertificate(cert, csr) &&
		!isInRenewalPeriod(cert, certID, getFlagRenewDays(cmd), time.Now()) {
		return nil
	}

	// This is just meant to be informal for the user.
	log.Info("acme: Trying renewal.",
		log.CertNameAttr(certID),
		slog.Any("time-remaining", FormattableDuration(cert.NotAfter.Sub(time.Now().UTC()))),
	)

	client, err := lazyClient()
	if err != nil {
		return fmt.Errorf("set up client: %w", err)
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

	hook.AddPathToMetadata(meta, certRes, certsStorage, options)

	return hook.Launch(ctx, cmd.String(flgDeployHook), cmd.Duration(flgDeployHookTimeout), meta)
}

func getFlagRenewDays(cmd *cli.Command) int {
	if cmd.IsSet(flgRenewDays) {
		return cmd.Int(flgRenewDays)
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
			FormattableDuration(dueDate.Sub(now)),
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
	if cmd.Bool(flgARIDisable) {
		return nil, "", nil
	}

	client, err := lazyClient()
	if err != nil {
		return nil, "", fmt.Errorf("set up client: %w", err)
	}

	willingToSleep := cmd.Duration(flgARIWaitToRenewDuration)

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

	replacesCertID, err := certificate.MakeARICertID(cert)
	if err != nil {
		return nil, "", fmt.Errorf("error while constructing the ARI CertID for domain %q: %w", certID, err)
	}

	return ariRenewalTime, replacesCertID, nil
}

// getARIRenewalTime checks if the certificate needs to be renewed using the renewalInfo endpoint.
func getARIRenewalTime(ctx context.Context, willingToSleep time.Duration, cert *x509.Certificate, certID string, client *lego.Client) *time.Time {
	if cert.IsCA {
		log.Fatal("Certificate bundle starts with a CA certificate.", log.CertNameAttr(certID))
	}

	renewalInfo, err := client.Certificate.GetRenewalInfo(ctx, certificate.RenewalInfoRequest{Cert: cert})
	if err != nil {
		if errors.Is(err, api.ErrNoARI) {
			log.Warn("acme: the server does not advertise a renewal info endpoint.",
				log.CertNameAttr(certID),
				log.ErrorAttr(err),
			)

			return nil
		}

		log.Warn("acme: calling renewal info endpoint",
			log.CertNameAttr(certID),
			log.ErrorAttr(err),
		)

		return nil
	}

	now := time.Now().UTC()

	renewalTime := renewalInfo.ShouldRenewAt(now, willingToSleep)
	if renewalTime == nil {
		log.Info("acme: renewalInfo endpoint indicates that renewal is not needed.", log.CertNameAttr(certID))
		return nil
	}

	log.Info("acme: renewalInfo endpoint indicates that renewal is needed.", log.CertNameAttr(certID))

	if renewalInfo.ExplanationURL != "" {
		log.Info("acme: renewalInfo endpoint provided an explanation.",
			log.CertNameAttr(certID),
			slog.String("explanationURL", renewalInfo.ExplanationURL),
		)
	}

	return renewalTime
}

func randomSleep(cmd *cli.Command) {
	// https://github.com/go-acme/lego/issues/1656
	// https://github.com/certbot/certbot/blob/284023a1b7672be2bd4018dd7623b3b92197d4b0/certbot/certbot/_internal/renewal.py#L435-L440
	if !isatty.IsTerminal(os.Stdout.Fd()) && !cmd.Bool(flgNoRandomSleep) {
		// https://github.com/certbot/certbot/blob/284023a1b7672be2bd4018dd7623b3b92197d4b0/certbot/certbot/_internal/renewal.py#L472
		const jitter = 8 * time.Minute

		rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
		sleepTime := time.Duration(rnd.Int63n(int64(jitter)))

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

type FormattableDuration time.Duration

func (f FormattableDuration) String() string {
	d := time.Duration(f)

	days := int(math.Trunc(d.Hours() / 24))
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	ns := int(d.Nanoseconds()) % int(time.Second)

	s := new(strings.Builder)

	if days > 0 {
		_, _ = fmt.Fprintf(s, "%dd", days)
	}

	if hours > 0 {
		_, _ = fmt.Fprintf(s, "%dh", hours)
	}

	if minutes > 0 {
		_, _ = fmt.Fprintf(s, "%dm", minutes)
	}

	if seconds > 0 {
		_, _ = fmt.Fprintf(s, "%ds", seconds)
	}

	if ns > 0 {
		_, _ = fmt.Fprintf(s, "%dns", ns)
	}

	return s.String()
}

func (f FormattableDuration) LogValue() slog.Value {
	return slog.StringValue(f.String())
}
