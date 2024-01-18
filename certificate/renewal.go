package certificate

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/acme"
)

// RenewalInfoRequest contains the necessary renewal information.
type RenewalInfoRequest struct {
	Cert *x509.Certificate
}

// RenewalInfoResponse is a wrapper around acme.RenewalInfoResponse that provides a method for determining when to renew a certificate.
type RenewalInfoResponse struct {
	acme.RenewalInfoResponse
}

// ShouldRenewAt determines the optimal renewal time based on the current time (UTC),renewal window suggest by ARI, and the client's willingness to sleep.
// It returns a pointer to a time.Time value indicating when the renewal should be attempted or nil if deferred until the next normal wake time.
// This method implements the RECOMMENDED algorithm described in draft-ietf-acme-ari.
//
// - (4.1-11. Getting Renewal Information) https://datatracker.ietf.org/doc/draft-ietf-acme-ari/
func (r *RenewalInfoResponse) ShouldRenewAt(now time.Time, willingToSleep time.Duration) *time.Time {
	// Explicitly convert all times to UTC.
	now = now.UTC()
	start := r.SuggestedWindow.Start.UTC()
	end := r.SuggestedWindow.End.UTC()

	// Select a uniform random time within the suggested window.
	window := end.Sub(start)
	randomDuration := time.Duration(rand.Int63n(int64(window)))
	rt := start.Add(randomDuration)

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
// The caller should attempt to renew the certificate at the time indicated by the ShouldRenewAt method of the returned RenewalInfoResponse object.
//
// Note: this endpoint is part of a draft specification, not all ACME servers will implement it.
// This method will return api.ErrNoARI if the server does not advertise a renewal info endpoint.
//
// https://datatracker.ietf.org/doc/draft-ietf-acme-ari
func (c *Certifier) GetRenewalInfo(req RenewalInfoRequest) (*RenewalInfoResponse, error) {
	certID, err := makeARICertID(req.Cert)
	if err != nil {
		return nil, fmt.Errorf("error making certID: %w", err)
	}

	resp, err := c.core.Certificates.GetRenewalInfo(certID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var info RenewalInfoResponse
	err = json.NewDecoder(resp.Body).Decode(&info)
	if err != nil {
		return nil, err
	}
	return &info, nil
}

// UpdateRenewalInfo sends an update to the ACME server's renewal info endpoint to indicate that the client has successfully replaced a certificate.
// A certificate is considered replaced when its revocation would not disrupt any ongoing services,
// for instance because it has been renewed and the new certificate is in use, or because it is no longer in use.
//
// Note: this endpoint is part of a draft specification, not all ACME servers will implement it.
// This method will return api.ErrNoARI if the server does not advertise a renewal info endpoint.
//
// https://datatracker.ietf.org/doc/draft-ietf-acme-ari
func (c *Certifier) UpdateRenewalInfo(req RenewalInfoRequest) error {
	certID, err := makeARICertID(req.Cert)
	if err != nil {
		return fmt.Errorf("error making certID: %w", err)
	}

	_, err = c.core.Certificates.UpdateRenewalInfo(acme.RenewalInfoUpdateRequest{
		CertID:   certID,
		Replaced: true,
	})
	if err != nil {
		return err
	}

	return nil
}

// makeARICertID constructs a certificate identifier as described in draft-ietf-acme-ari-02, section 4.1.
func makeARICertID(leaf *x509.Certificate) (string, error) {
	if leaf == nil {
		return "", errors.New("leaf certificate is nil")
	}

	return fmt.Sprintf("%s.%s",
		strings.TrimRight(base64.URLEncoding.EncodeToString(leaf.AuthorityKeyId), "="),
		strings.TrimRight(base64.URLEncoding.EncodeToString(leaf.SerialNumber.Bytes()), "="),
	), nil
}
