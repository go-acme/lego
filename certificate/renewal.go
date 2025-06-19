package certificate

import (
	"crypto/x509"
	"encoding/asn1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
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

	// RetryAfter header indicating the polling interval that the ACME server recommends.
	// Conforming clients SHOULD query the renewalInfo URL again after the RetryAfter period has passed,
	// as the server may provide a different suggestedWindow.
	// https://www.rfc-editor.org/rfc/rfc9773.html#section-4.2
	RetryAfter time.Duration
}

// ShouldRenewAt determines the optimal renewal time based on the current time (UTC),renewal window suggest by ARI, and the client's willingness to sleep.
// It returns a pointer to a time.Time value indicating when the renewal should be attempted or nil if deferred until the next normal wake time.
// This method implements the RECOMMENDED algorithm described in RFC 9773.
//
// - (4.1-11. Getting Renewal Information) https://www.rfc-editor.org/rfc/rfc9773.html
func (r *RenewalInfoResponse) ShouldRenewAt(now time.Time, willingToSleep time.Duration) *time.Time {
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
// The caller should attempt to renew the certificate at the time indicated by the ShouldRenewAt method of the returned RenewalInfoResponse object.
//
// Note: this endpoint is part of a draft specification, not all ACME servers will implement it.
// This method will return api.ErrNoARI if the server does not advertise a renewal info endpoint.
//
// https://www.rfc-editor.org/rfc/rfc9773.html
func (c *Certifier) GetRenewalInfo(req RenewalInfoRequest) (*RenewalInfoResponse, error) {
	certID, err := MakeARICertID(req.Cert)
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

	if retry := resp.Header.Get("Retry-After"); retry != "" {
		info.RetryAfter, err = time.ParseDuration(retry + "s")
		if err != nil {
			return nil, err
		}
	}

	return &info, nil
}

// MakeARICertID constructs a certificate identifier as described in RFC 9773, section 4.1.
func MakeARICertID(leaf *x509.Certificate) (string, error) {
	if leaf == nil {
		return "", errors.New("leaf certificate is nil")
	}

	// Marshal the Serial Number into DER.
	der, err := asn1.Marshal(leaf.SerialNumber)
	if err != nil {
		return "", err
	}

	// Check if the DER encoded bytes are sufficient (at least 3 bytes: tag,
	// length, and value).
	if len(der) < 3 {
		return "", errors.New("invalid DER encoding of serial number")
	}

	// Extract only the integer bytes from the DER encoded Serial Number
	// Skipping the first 2 bytes (tag and length).
	serial := base64.RawURLEncoding.EncodeToString(der[2:])

	// Convert the Authority Key Identifier to base64url encoding without
	// padding.
	aki := base64.RawURLEncoding.EncodeToString(leaf.AuthorityKeyId)

	// Construct the final identifier by concatenating AKI and Serial Number.
	return fmt.Sprintf("%s.%s", aki, serial), nil
}
