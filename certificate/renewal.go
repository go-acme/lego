package certificate

import (
	"crypto"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"math/rand"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/acme"
)

// RenewalInfoRequest contains the necessary renewal information.
type RenewalInfoRequest struct {
	Cert   *x509.Certificate
	Issuer *x509.Certificate
	// HashName must be the string representation of a crypto.Hash constant in the golang.org/x/crypto package (e.g. "SHA-256").
	// The correct value depends on the algorithm expected by the ACME server's ARI implementation.
	HashName string
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
	certID, err := makeCertID(req.Cert, req.Issuer, req.HashName)
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
	certID, err := makeCertID(req.Cert, req.Issuer, req.HashName)
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

// makeCertID returns a base64url-encoded string that uniquely identifies a certificate to endpoints
// that implement the draft-ietf-acme-ari specification: https://datatracker.ietf.org/doc/draft-ietf-acme-ari.
// hashName must be the string representation of a crypto.Hash constant in the golang.org/x/crypto package.
// Supported hash functions are SHA-1, SHA-256, SHA-384, and SHA-512.
func makeCertID(leaf, issuer *x509.Certificate, hashName string) (string, error) {
	if leaf == nil {
		return "", fmt.Errorf("leaf certificate is nil")
	}
	if issuer == nil {
		return "", fmt.Errorf("issuer certificate is nil")
	}

	var hashFunc crypto.Hash
	var oid asn1.ObjectIdentifier

	switch hashName {
	// The following correlation of hashFunc to OID is copied from a private mapping in golang.org/x/crypto/ocsp:
	// https://cs.opensource.google/go/x/crypto/+/refs/tags/v0.8.0:ocsp/ocsp.go;l=156
	case crypto.SHA1.String():
		hashFunc = crypto.SHA1
		oid = asn1.ObjectIdentifier([]int{1, 3, 14, 3, 2, 26})

	case crypto.SHA256.String():
		hashFunc = crypto.SHA256
		oid = asn1.ObjectIdentifier([]int{2, 16, 840, 1, 101, 3, 4, 2, 1})

	case crypto.SHA384.String():
		hashFunc = crypto.SHA384
		oid = asn1.ObjectIdentifier([]int{2, 16, 840, 1, 101, 3, 4, 2, 2})

	case crypto.SHA512.String():
		hashFunc = crypto.SHA512
		oid = asn1.ObjectIdentifier([]int{2, 16, 840, 1, 101, 3, 4, 2, 3})

	default:
		return "", fmt.Errorf("hashName %q is not supported by this package", hashName)
	}

	if !hashFunc.Available() {
		// This should never happen.
		return "", fmt.Errorf("hash function %q is not available on your platform", hashFunc)
	}

	var spki struct {
		Algorithm pkix.AlgorithmIdentifier
		PublicKey asn1.BitString
	}

	_, err := asn1.Unmarshal(issuer.RawSubjectPublicKeyInfo, &spki)
	if err != nil {
		return "", err
	}
	h := hashFunc.New()
	h.Write(spki.PublicKey.RightAlign())
	issuerKeyHash := h.Sum(nil)

	h.Reset()
	h.Write(issuer.RawSubject)
	issuerNameHash := h.Sum(nil)

	type certID struct {
		HashAlgorithm  pkix.AlgorithmIdentifier
		IssuerNameHash []byte
		IssuerKeyHash  []byte
		SerialNumber   *big.Int
	}

	// DER-encode the CertID ASN.1 sequence [RFC6960].
	certIDBytes, err := asn1.Marshal(certID{
		HashAlgorithm: pkix.AlgorithmIdentifier{
			Algorithm: oid,
		},
		IssuerNameHash: issuerNameHash,
		IssuerKeyHash:  issuerKeyHash,
		SerialNumber:   leaf.SerialNumber,
	})
	if err != nil {
		return "", err
	}

	// base64url-encode [RFC4648] the bytes of the DER-encoded CertID ASN.1 sequence [RFC6960].
	encodedBytes := base64.URLEncoding.EncodeToString(certIDBytes)

	// Any trailing '=' characters MUST be stripped.
	return strings.TrimRight(encodedBytes, "="), nil
}
