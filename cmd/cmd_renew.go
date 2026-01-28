package cmd

import (
	"context"
	"crypto"
	"crypto/x509"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"math/rand"
	"os"
	"slices"
	"strings"
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

func createRenew() *cli.Command {
	return &cli.Command{
		Name:   "renew",
		Usage:  "Renew a certificate",
		Action: renew,
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			// we require either domains or csr, but not both
			hasDomains := len(cmd.StringSlice(flgDomains)) > 0

			hasCsr := cmd.String(flgCSR) != ""
			if hasDomains && hasCsr {
				log.Fatal(fmt.Sprintf("Please specify either --%s/-d or --%s/-c, but not both", flgDomains, flgCSR))
			}

			if !hasDomains && !hasCsr {
				log.Fatal(fmt.Sprintf("Please specify --%s/-d (or --%s/-c if you already have a CSR)", flgDomains, flgCSR))
			}

			if cmd.Bool(flgForceCertDomains) && hasCsr {
				log.Fatal(fmt.Sprintf("--%s only works with --%s/-d, --%s/-c doesn't support this option.", flgForceCertDomains, flgDomains, flgCSR))
			}

			return ctx, nil
		},
		Flags: createRenewFlags(),
	}
}

func renew(ctx context.Context, cmd *cli.Command) error {
	accountsStorage, err := storage.NewAccountsStorage(newAccountsStorageConfig(cmd))
	if err != nil {
		log.Fatal("Accounts storage initialization", log.ErrorAttr(err))
	}

	keyType := getKeyType(cmd)

	account := setupAccount(ctx, keyType, accountsStorage)

	if account.Registration == nil {
		log.Fatal("The account is not registered. Use 'run' to register a new account.", slog.String("email", account.Email))
	}

	certsStorage, err := storage.NewCertificatesStorage(newCertificatesWriterConfig(cmd))
	if err != nil {
		log.Fatal("Certificates storage", log.ErrorAttr(err))
	}

	meta := map[string]string{
		hook.EnvAccountEmail: account.Email,
	}

	// CSR
	if cmd.IsSet(flgCSR) {
		return renewForCSR(ctx, cmd, account, keyType, certsStorage, meta)
	}

	// Domains
	return renewForDomains(ctx, cmd, account, keyType, certsStorage, meta)
}

func renewForDomains(ctx context.Context, cmd *cli.Command, account *storage.Account, keyType certcrypto.KeyType, certsStorage *storage.CertificatesStorage, meta map[string]string) error {
	domains := cmd.StringSlice(flgDomains)
	domain := domains[0]

	// load the cert resource from files.
	// We store the certificate, private key and metadata in different files
	// as web servers would not be able to work with a combined file.
	certificates, err := certsStorage.ReadCertificate(domain, storage.ExtCert)
	if err != nil {
		log.Fatal("Error while loading the certificate.",
			log.DomainAttr(domain),
			log.ErrorAttr(err),
		)
	}

	cert := certificates[0]

	var (
		ariRenewalTime *time.Time
		replacesCertID string
	)

	var client *lego.Client

	if !cmd.Bool(flgARIDisable) {
		client = setupClient(cmd, account, keyType)

		willingToSleep := cmd.Duration(flgARIWaitToRenewDuration)

		ariRenewalTime = getARIRenewalTime(ctx, willingToSleep, cert, domain, client)
		if ariRenewalTime != nil {
			now := time.Now().UTC()

			// Figure out if we need to sleep before renewing.
			if ariRenewalTime.After(now) {
				log.Info("Sleeping until renewal time",
					log.DomainAttr(domain),
					slog.Duration("sleep", ariRenewalTime.Sub(now)),
					slog.Time("renewalTime", *ariRenewalTime),
				)
				time.Sleep(ariRenewalTime.Sub(now))
			}
		}

		replacesCertID, err = certificate.MakeARICertID(cert)
		if err != nil {
			log.Fatal("Error while construction the ARI CertID.", log.DomainAttr(domain), log.ErrorAttr(err))
		}
	}

	forceDomains := cmd.Bool(flgForceCertDomains)

	certDomains := certcrypto.ExtractDomains(cert)

	if ariRenewalTime == nil && !needRenewal(cert, domain, cmd.Int(flgRenewDays), cmd.Bool(flgRenewDynamic)) &&
		(!forceDomains || slices.Equal(certDomains, domains)) {
		return nil
	}

	if client == nil {
		client = setupClient(cmd, account, keyType)
	}

	// This is just meant to be informal for the user.
	timeLeft := cert.NotAfter.Sub(time.Now().UTC())
	log.Info("acme: Trying renewal.",
		log.DomainAttr(domain),
		slog.Int("hoursRemaining", int(timeLeft.Hours())),
	)

	var privateKey crypto.PrivateKey

	if cmd.Bool(flgReuseKey) {
		keyBytes, errR := certsStorage.ReadFile(domain, storage.ExtKey)
		if errR != nil {
			log.Fatal("Error while loading the private key.",
				log.DomainAttr(domain),
				log.ErrorAttr(errR),
			)
		}

		privateKey, errR = certcrypto.ParsePEMPrivateKey(keyBytes)
		if errR != nil {
			return errR
		}
	}

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

	renewalDomains := slices.Clone(domains)
	if !forceDomains {
		renewalDomains = merge(certDomains, domains)
	}

	request := newObtainRequest(cmd, renewalDomains)

	request.PrivateKey = privateKey

	if replacesCertID != "" {
		request.ReplacesCertID = replacesCertID
	}

	certRes, err := client.Certificate.Obtain(ctx, request)
	if err != nil {
		log.Fatal("Could not obtain the certificate.", log.ErrorAttr(err))
	}

	certRes.Domain = domain

	certsStorage.SaveResource(certRes)

	hook.AddPathToMetadata(meta, certRes.Domain, certRes, certsStorage)

	return hook.Launch(ctx, cmd.String(flgDeployHook), cmd.Duration(flgDeployHookTimeout), meta)
}

func renewForCSR(ctx context.Context, cmd *cli.Command, account *storage.Account, keyType certcrypto.KeyType, certsStorage *storage.CertificatesStorage, meta map[string]string) error {
	csr, err := readCSRFile(cmd.String(flgCSR))
	if err != nil {
		log.Fatal("Could not read CSR file.",
			slog.String(flgCSR, cmd.String(flgCSR)),
			log.ErrorAttr(err),
		)
	}

	domain, err := certcrypto.GetCSRMainDomain(csr)
	if err != nil {
		log.Fatal("Could not get CSR main domain.", log.ErrorAttr(err))
	}

	// load the cert resource from files.
	// We store the certificate, private key and metadata in different files
	// as web servers would not be able to work with a combined file.
	certificates, err := certsStorage.ReadCertificate(domain, storage.ExtCert)
	if err != nil {
		log.Fatal("Error while loading the certificate.",
			log.DomainAttr(domain),
			log.ErrorAttr(err),
		)
	}

	cert := certificates[0]

	var (
		ariRenewalTime *time.Time
		replacesCertID string
	)

	var client *lego.Client

	if !cmd.Bool(flgARIDisable) {
		client = setupClient(cmd, account, keyType)

		willingToSleep := cmd.Duration(flgARIWaitToRenewDuration)

		ariRenewalTime = getARIRenewalTime(ctx, willingToSleep, cert, domain, client)
		if ariRenewalTime != nil {
			now := time.Now().UTC()

			// Figure out if we need to sleep before renewing.
			if ariRenewalTime.After(now) {
				log.Info("Sleeping until renewal time",
					log.DomainAttr(domain),
					slog.Duration("sleep", ariRenewalTime.Sub(now)),
					slog.Time("renewalTime", *ariRenewalTime),
				)
				time.Sleep(ariRenewalTime.Sub(now))
			}
		}

		replacesCertID, err = certificate.MakeARICertID(cert)
		if err != nil {
			log.Fatal("Error while construction the ARI CertID.", log.DomainAttr(domain), log.ErrorAttr(err))
		}
	}

	if ariRenewalTime == nil && !needRenewal(cert, domain, cmd.Int(flgRenewDays), cmd.Bool(flgRenewDynamic)) {
		return nil
	}

	if client == nil {
		client = setupClient(cmd, account, keyType)
	}

	// This is just meant to be informal for the user.
	timeLeft := cert.NotAfter.Sub(time.Now().UTC())
	log.Info("acme: Trying renewal.",
		log.DomainAttr(domain),
		slog.Int("hoursRemaining", int(timeLeft.Hours())),
	)

	request := newObtainForCSRRequest(cmd, csr)

	if replacesCertID != "" {
		request.ReplacesCertID = replacesCertID
	}

	certRes, err := client.Certificate.ObtainForCSR(ctx, request)
	if err != nil {
		log.Fatal("Could not obtain the certificate for CSR.", log.ErrorAttr(err))
	}

	certsStorage.SaveResource(certRes)

	hook.AddPathToMetadata(meta, domain, certRes, certsStorage)

	return hook.Launch(ctx, cmd.String(flgDeployHook), cmd.Duration(flgDeployHookTimeout), meta)
}

func needRenewal(x509Cert *x509.Certificate, domain string, days int, dynamic bool) bool {
	if x509Cert.IsCA {
		log.Fatal("Certificate bundle starts with a CA certificate.", log.DomainAttr(domain))
	}

	if dynamic {
		return needRenewalDynamic(x509Cert, domain, time.Now())
	}

	if days < 0 {
		return true
	}

	notAfter := int(time.Until(x509Cert.NotAfter).Hours() / 24.0)
	if notAfter <= days {
		return true
	}

	log.Infof(
		log.LazySprintf("Skip renewal: the certificate expires in %d days, the number of days defined to perform the renewal is %d.",
			notAfter, days),
		log.DomainAttr(domain),
	)

	return false
}

func needRenewalDynamic(x509Cert *x509.Certificate, domain string, now time.Time) bool {
	lifetime := x509Cert.NotAfter.Sub(x509Cert.NotBefore)

	var divisor int64 = 3
	if lifetime.Round(24*time.Hour).Hours()/24.0 <= 10 {
		divisor = 2
	}

	dueDate := x509Cert.NotAfter.Add(-1 * time.Duration(lifetime.Nanoseconds()/divisor))

	if dueDate.Before(now) {
		return true
	}

	log.Infof(log.LazySprintf("Skip renewal: The certificate expires at %s, the renewal can be performed in %s.",
		x509Cert.NotAfter.Format(time.RFC3339), FormattableDuration(dueDate.Sub(now))), log.DomainAttr(domain))

	return false
}

// getARIRenewalTime checks if the certificate needs to be renewed using the renewalInfo endpoint.
func getARIRenewalTime(ctx context.Context, willingToSleep time.Duration, cert *x509.Certificate, domain string, client *lego.Client) *time.Time {
	if cert.IsCA {
		log.Fatal("Certificate bundle starts with a CA certificate.", log.DomainAttr(domain))
	}

	renewalInfo, err := client.Certificate.GetRenewalInfo(ctx, certificate.RenewalInfoRequest{Cert: cert})
	if err != nil {
		if errors.Is(err, api.ErrNoARI) {
			log.Warn("acme: the server does not advertise a renewal info endpoint.",
				log.DomainAttr(domain),
				log.ErrorAttr(err),
			)

			return nil
		}

		log.Warn("acme: calling renewal info endpoint",
			log.DomainAttr(domain),
			log.ErrorAttr(err),
		)

		return nil
	}

	now := time.Now().UTC()

	renewalTime := renewalInfo.ShouldRenewAt(now, willingToSleep)
	if renewalTime == nil {
		log.Info("acme: renewalInfo endpoint indicates that renewal is not needed.", log.DomainAttr(domain))
		return nil
	}

	log.Info("acme: renewalInfo endpoint indicates that renewal is needed.", log.DomainAttr(domain))

	if renewalInfo.ExplanationURL != "" {
		log.Info("acme: renewalInfo endpoint provided an explanation.",
			log.DomainAttr(domain),
			slog.String("explanationURL", renewalInfo.ExplanationURL),
		)
	}

	return renewalTime
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

type FormattableDuration time.Duration

func (f FormattableDuration) String() string {
	d := time.Duration(f)

	days := int(math.Trunc(d.Hours() / 24))
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	ns := int(d.Nanoseconds()) % int(time.Second)

	var s strings.Builder

	if days > 0 {
		s.WriteString(fmt.Sprintf("%dd", days))
	}

	if hours > 0 {
		s.WriteString(fmt.Sprintf("%dh", hours))
	}

	if minutes > 0 {
		s.WriteString(fmt.Sprintf("%dm", minutes))
	}

	if seconds > 0 {
		s.WriteString(fmt.Sprintf("%ds", seconds))
	}

	if ns > 0 {
		s.WriteString(fmt.Sprintf("%dns", ns))
	}

	return s.String()
}
