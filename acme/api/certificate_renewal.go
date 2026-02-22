package api

import (
	"context"
	"crypto/x509"
	"encoding/asn1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-acme/lego/v5/acme"
)

// ErrNoARI is returned when the server does not advertise a renewal info endpoint.
var ErrNoARI = errors.New("renewalInfo[get/post]: server does not advertise a renewal info endpoint")

// GetRenewalInfo GETs renewal information for a certificate from the renewalInfo endpoint.
// This is used to determine if a certificate needs to be renewed.
//
// Note: this endpoint is part of a draft specification, not all ACME servers will implement it.
// This method will return api.ErrNoARI if the server does not advertise a renewal info endpoint.
//
// https://www.rfc-editor.org/rfc/rfc9773.html
func (c *CertificateService) GetRenewalInfo(ctx context.Context, certID string) (*acme.ExtendedRenewalInfo, error) {
	if c.core.GetDirectory().RenewalInfo == "" {
		return nil, ErrNoARI
	}

	if certID == "" {
		return nil, errors.New("renewalInfo[get]: 'certID' cannot be empty")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.core.GetDirectory().RenewalInfo+"/"+certID, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.core.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() { _ = resp.Body.Close() }()

	info := new(acme.ExtendedRenewalInfo)

	err = json.NewDecoder(resp.Body).Decode(info)
	if err != nil {
		return nil, err
	}

	if retry := resp.Header.Get("Retry-After"); retry != "" {
		info.RetryAfter, err = ParseRetryAfter(retry)
		if err != nil {
			return nil, fmt.Errorf("failed to parse Retry-After header: %w", err)
		}
	}

	return info, nil
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
