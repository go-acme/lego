package clock

import (
	"crypto/sha256"
	"crypto/x509"
	"encoding/binary"
	"math/big"
	"math/rand/v2"
	"time"

	"github.com/go-acme/lego/v5/log"
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

func (c *RenewClock) getDueDate(x509Cert *x509.Certificate, days int, now time.Time) time.Time {
	if days == c.noDays {
		lifetime := x509Cert.NotAfter.Sub(x509Cert.NotBefore)

		var divisor int64 = 3
		if lifetime.Round(24*time.Hour).Hours()/24.0 <= 10 {
			divisor = 2
		}

		variance := time.Duration(c.jitter(x509Cert.SerialNumber) * float64(lifetime))

		return x509Cert.NotAfter.Add(-1 * (time.Duration(lifetime.Nanoseconds()/divisor) + variance))
	}

	if days < 0 {
		// if the number of days is negative: always renew the certificate.
		return now
	}

	return x509Cert.NotAfter.Add(-1 * time.Duration(days) * 24 * time.Hour)
}

func deterministicJitter(n *big.Int) float64 {
	hash := sha256.Sum256([]byte(n.String()))

	r := rand.New(
		rand.NewPCG(
			binary.LittleEndian.Uint64(hash[0:16]),
			binary.LittleEndian.Uint64(hash[16:32]),
		),
	)

	// between -5% and 5%
	return r.Float64()*0.1 - 0.05
}
