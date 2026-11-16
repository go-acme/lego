package clock

import (
	"crypto/sha256"
	"crypto/x509"
	"encoding/binary"
	"log/slog"
	"math/big"
	"math/rand/v2"
	"os"
	"time"

	"github.com/go-acme/lego/v5/log"
	"github.com/mattn/go-isatty"
)

type RenewClock struct {
	now    func() time.Time
	noDays int
	jitter func(n *big.Int) float64
}

func NewRenewClock(noDays int) *RenewClock {
	return &RenewClock{
		now:    time.Now,
		noDays: noDays,
		jitter: deterministicJitter,
	}
}

func (c *RenewClock) IsInRenewalPeriod(cert *x509.Certificate, certID string, days int) bool {
	now := c.now()

	dueDate := c.getDueDate(cert, days, now)

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

func (c *RenewClock) getDueDate(cert *x509.Certificate, days int, now time.Time) time.Time {
	if days == c.noDays {
		lifetime := cert.NotAfter.Sub(cert.NotBefore)

		var divisor int64 = 3
		if lifetime.Round(24*time.Hour).Hours()/24.0 <= 10 {
			divisor = 2
		}

		variance := time.Duration(c.jitter(cert.SerialNumber) * float64(lifetime))

		return cert.NotAfter.Add(-1 * (time.Duration(lifetime.Nanoseconds()/divisor) + variance))
	}

	if days < 0 {
		// if the number of days is negative: always renew the certificate.
		return now
	}

	return cert.NotAfter.Add(-1 * time.Duration(days) * 24 * time.Hour)
}

func deterministicJitter(n *big.Int) float64 {
	hash := sha256.Sum256([]byte(n.String()))

	r := rand.New(
		rand.NewPCG(
			binary.LittleEndian.Uint64(hash[0:8]),
			binary.LittleEndian.Uint64(hash[8:16]),
		),
	)

	// between -5% and 5%
	return r.Float64()*0.1 - 0.05
}

func RandomSleep() {
	// https://github.com/go-acme/lego/issues/1656
	// https://github.com/certbot/certbot/blob/284023a1b7672be2bd4018dd7623b3b92197d4b0/certbot/certbot/_internal/renewal.py#L435-L440
	if !isatty.IsTerminal(os.Stdout.Fd()) {
		// https://github.com/certbot/certbot/blob/284023a1b7672be2bd4018dd7623b3b92197d4b0/certbot/certbot/_internal/renewal.py#L472
		const jitter = 8 * time.Minute

		sleepTime := time.Duration(rand.Int64N(int64(jitter)))

		log.Info("renewal: random delay.", slog.Duration("sleep", sleepTime))
		time.Sleep(sleepTime)
	}
}

func ARISleep(certID string, ariRenewalTime *time.Time) {
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
