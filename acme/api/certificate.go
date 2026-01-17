package api

import (
	"bytes"
	"context"
	"encoding/pem"
	"errors"
	"io"
	"net/http"

	"github.com/go-acme/lego/v5/acme"
)

// maxBodySize is the maximum size of body that we will read.
const maxBodySize = 1024 * 1024

type CertificateService service

// Get Returns the certificate and the issuer certificate.
// 'bundle' is only applied if the issuer is provided by the 'up' link.
func (c *CertificateService) Get(ctx context.Context, certURL string, bundle bool) ([]byte, []byte, error) {
	cert, _, err := c.get(ctx, certURL, bundle)
	if err != nil {
		return nil, nil, err
	}

	return cert.Cert, cert.Issuer, nil
}

// GetAll the certificates and the alternate certificates.
// bundle' is only applied if the issuer is provided by the 'up' link.
func (c *CertificateService) GetAll(ctx context.Context, certURL string, bundle bool) (map[string]*acme.RawCertificate, error) {
	cert, headers, err := c.get(ctx, certURL, bundle)
	if err != nil {
		return nil, err
	}

	certs := map[string]*acme.RawCertificate{certURL: cert}

	// URLs of "alternate" link relation
	// - https://www.rfc-editor.org/rfc/rfc8555.html#section-7.4.2
	alts := getLinks(headers, "alternate")

	for _, alt := range alts {
		altCert, _, err := c.get(ctx, alt, bundle)
		if err != nil {
			return nil, err
		}

		certs[alt] = altCert
	}

	return certs, nil
}

// Revoke Revokes a certificate.
func (c *CertificateService) Revoke(ctx context.Context, req acme.RevokeCertMessage) error {
	_, err := c.core.post(ctx, c.core.GetDirectory().RevokeCertURL, req, nil)
	return err
}

// get Returns the certificate and the "up" link.
func (c *CertificateService) get(ctx context.Context, certURL string, bundle bool) (*acme.RawCertificate, http.Header, error) {
	if certURL == "" {
		return nil, nil, errors.New("certificate[get]: empty URL")
	}

	resp, err := c.core.postAsGet(ctx, certURL, nil)
	if err != nil {
		return nil, nil, err
	}

	data, err := io.ReadAll(http.MaxBytesReader(nil, resp.Body, maxBodySize))
	if err != nil {
		return nil, resp.Header, err
	}

	cert := c.getCertificateChain(data, bundle)

	return cert, resp.Header, err
}

// getCertificateChain Returns the certificate and the issuer certificate.
func (c *CertificateService) getCertificateChain(cert []byte, bundle bool) *acme.RawCertificate {
	// Get issuerCert from bundled response from Let's Encrypt
	// See https://community.letsencrypt.org/t/acme-v2-no-up-link-in-response/64962
	_, issuer := pem.Decode(cert)

	// If bundle is false, we want to return a single certificate.
	// To do this, we remove the issuer cert(s) from the issued cert.
	if !bundle {
		cert = bytes.TrimSuffix(cert, issuer)
	}

	return &acme.RawCertificate{Cert: cert, Issuer: issuer}
}
