package certificate

import (
	"context"
	"crypto"
	"crypto/x509"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/go-acme/lego/v5/acme"
	"github.com/go-acme/lego/v5/acme/api"
	"github.com/go-acme/lego/v5/certcrypto"
	"github.com/go-acme/lego/v5/log"
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
		randomDuration := time.Duration(rand.Int64N(int64(window)))
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

// RenewOptions options used by [Certifier.Renew].
type RenewOptions struct {
	NotBefore time.Time
	NotAfter  time.Time
	// If true, the []byte contains both the issuer certificate and your issued certificate as a bundle.
	Bundle           bool
	PreferredChain   string
	EnableCommonName bool

	Profile string

	UseARICertID bool

	AlwaysDeactivateAuthorizations bool
	// Not supported for CSR request.
	MustStaple     bool
	EmailAddresses []string
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

// Renew takes a Resource and tries to renew the certificate.
//
// If the renewal process succeeds, the new certificate will be returned in a new CertResource.
// Please be aware that this function will return a new certificate in ANY case that is not an error.
// If the server does not provide us with a new cert on a GET request to the CertURL
// this function will start a new-cert flow where a new certificate gets generated.
//
// If bundle is true, the []byte contains both the issuer certificate and your issued certificate as a bundle.
//
// For private key reuse the PrivateKey property of the passed in Resource should be non-nil.
func (c *Certifier) Renew(ctx context.Context, certRes Resource, options *RenewOptions) (*Resource, error) {
	// Input certificate is PEM encoded.
	// Decode it here as we may need the decoded cert later on in the renewal process.
	// The input may be a bundle or a single certificate.
	certificates, err := certcrypto.ParsePEMBundle(certRes.Certificate)
	if err != nil {
		return nil, err
	}

	x509Cert := certificates[0]
	if x509Cert.IsCA {
		return nil, fmt.Errorf("certificate bundle starts with a CA certificate (%s)", strings.Join(certRes.Domains, ", "))
	}

	// This is just meant to be informal for the user.
	timeLeft := x509Cert.NotAfter.Sub(time.Now().UTC())
	log.Info("Trying renewal.",
		log.DomainsAttr(certRes.Domains),
		slog.Int("hoursRemaining", int(timeLeft.Hours())),
	)

	// We always need to request a new certificate to renew.
	// Start by checking to see if the certificate was based off a CSR,
	// and use that if it's defined.
	if len(certRes.CSR) > 0 {
		request, errP := newRenewRequestForCSR(certRes, x509Cert, options)
		if errP != nil {
			return nil, errP
		}

		return c.ObtainForCSR(ctx, request)
	}

	request, err := newRenewRequest(certRes, x509Cert, options)
	if err != nil {
		return nil, err
	}

	return c.Obtain(ctx, request)
}

func newRenewRequestForCSR(certRes Resource, x509Cert *x509.Certificate, options *RenewOptions) (ObtainForCSRRequest, error) {
	csr, err := certcrypto.PemDecodeTox509CSR(certRes.CSR)
	if err != nil {
		return ObtainForCSRRequest{}, err
	}

	request := ObtainForCSRRequest{CSR: csr}

	if options == nil {
		return request, nil
	}

	request.NotBefore = options.NotBefore
	request.NotAfter = options.NotAfter
	request.Bundle = options.Bundle
	request.PreferredChain = options.PreferredChain
	request.Profile = options.Profile
	request.AlwaysDeactivateAuthorizations = options.AlwaysDeactivateAuthorizations

	if options.UseARICertID {
		request.ReplacesCertID, err = api.MakeARICertID(x509Cert)
		if err != nil {
			return ObtainForCSRRequest{}, fmt.Errorf("making ARI certificate ID: %w", err)
		}
	}

	return request, nil
}

func newRenewRequest(certRes Resource, x509Cert *x509.Certificate, options *RenewOptions) (ObtainRequest, error) {
	var privateKey crypto.Signer

	if certRes.PrivateKey != nil {
		var err error

		privateKey, err = certcrypto.ParsePEMPrivateKey(certRes.PrivateKey)
		if err != nil {
			return ObtainRequest{}, err
		}
	}

	request := ObtainRequest{
		Domains:    certcrypto.ExtractDomains(x509Cert),
		PrivateKey: privateKey,
	}

	if options == nil {
		return request, nil
	}

	request.MustStaple = options.MustStaple
	request.NotBefore = options.NotBefore
	request.NotAfter = options.NotAfter
	request.Bundle = options.Bundle
	request.PreferredChain = options.PreferredChain
	request.EnableCommonName = options.EnableCommonName
	request.EmailAddresses = options.EmailAddresses
	request.Profile = options.Profile
	request.AlwaysDeactivateAuthorizations = options.AlwaysDeactivateAuthorizations

	if options.UseARICertID {
		var err error

		request.ReplacesCertID, err = api.MakeARICertID(x509Cert)
		if err != nil {
			return ObtainRequest{}, fmt.Errorf("making ARI certificate ID: %w", err)
		}
	}

	return request, nil
}
