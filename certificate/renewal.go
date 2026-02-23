package certificate

import (
	"context"
	"crypto/x509"
	"fmt"
	"math/rand"
	"time"

	"github.com/go-acme/lego/v5/acme"
	"github.com/go-acme/lego/v5/acme/api"
)

// RenewalInfo is a wrapper around acme.ExtendedRenewalInfo that provides a method for determining when to renew a certificate.
type RenewalInfo struct {
	*acme.ExtendedRenewalInfo
}

// ShouldRenewAt determines the optimal renewal time based on the current time (UTC),
// renewal window suggest by ARI, and the client's willingness to sleep.
// It returns a pointer to a time.Time value indicating when the renewal should be attempted or nil if deferred until the next normal wake time.
// This method implements the RECOMMENDED algorithm described in RFC 9773.
//
// - (4.1-11. Getting Renewal Information) https://www.rfc-editor.org/rfc/rfc9773.html
func (r *RenewalInfo) ShouldRenewAt(now time.Time, willingToSleep time.Duration) *time.Time {
	// Explicitly convert all times to UTC.
	now = now.UTC()
	start := r.SuggestedWindow.Start.UTC()
	end := r.SuggestedWindow.End.UTC()

	// Select a uniform random time within the suggested window.
	rt := start
	if window := end.Sub(start); window > 0 {
		randomDuration := time.Duration(rand.Int63n(int64(window)))
		rt = rt.Add(randomDuration)
	}

	// If the selected time is in the past, attempt renewal immediately.
	if rt.Before(now) {
		return &now
	}

	// Otherwise, if the client can schedule itself to attempt renewal at exactly the selected time, do so.
	willingToSleepUntil := now.Add(willingToSleep)
	if willingToSleepUntil.After(rt) || willingToSleepUntil.Equal(rt) {
		return &rt
	}

	// TODO: Otherwise, if the selected time is before the next time that the client would wake up normally, attempt renewal immediately.

	// Otherwise, sleep until the next normal wake time, re-check ARI, and return to Step 1.
	return nil
}

// GetRenewalInfo sends a request to the ACME server's renewalInfo endpoint to obtain a suggested renewal window.
// The caller MUST provide the certificate and issuer certificate for the certificate they wish to renew.
// The caller should attempt to renew the certificate at the time indicated by the RenewalInfo.ShouldRenewAt method.
//
// Note: this endpoint is part of a draft specification, not all ACME servers will implement it.
// This method will return api.ErrNoARI if the server does not advertise a renewal info endpoint.
//
// https://www.rfc-editor.org/rfc/rfc9773.html
func (c *Certifier) GetRenewalInfo(ctx context.Context, cert *x509.Certificate) (*RenewalInfo, error) {
	certID, err := api.MakeARICertID(cert)
	if err != nil {
		return nil, fmt.Errorf("error making certID: %w", err)
	}

	info, err := c.core.Certificates.GetRenewalInfo(ctx, certID)
	if err != nil {
		return nil, err
	}

	return &RenewalInfo{ExtendedRenewalInfo: info}, err
}
