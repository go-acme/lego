package api

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io"
	"net/http"

	"github.com/go-acme/lego/v4/acme"
	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/log"
)

// maxBodySize is the maximum size of body that we will read.
const maxBodySize = 1024 * 1024

type CertificateService service

// Get Returns the certificate and the issuer certificate.
// 'bundle' is only applied if the issuer is provided by the 'up' link.
func (c *CertificateService) Get(certURL string, bundle bool) ([]byte, []byte, error) {
	cert, _, err := c.get(certURL, bundle)
	if err != nil {
		return nil, nil, err
	}

	return cert.Cert, cert.Issuer, nil
}

// GetAll the certificates and the alternate certificates.
// bundle' is only applied if the issuer is provided by the 'up' link.
func (c *CertificateService) GetAll(certURL string, bundle bool) (map[string]*acme.RawCertificate, error) {
	cert, headers, err := c.get(certURL, bundle)
	if err != nil {
		return nil, err
	}

	certs := map[string]*acme.RawCertificate{certURL: cert}

	// URLs of "alternate" link relation
	// - https://www.rfc-editor.org/rfc/rfc8555.html#section-7.4.2
	alts := getLinks(headers, "alternate")

	for _, alt := range alts {
		altCert, _, err := c.get(alt, bundle)
		if err != nil {
			return nil, err
		}

		certs[alt] = altCert
	}

	return certs, nil
}

// Revoke Revokes a certificate.
func (c *CertificateService) Revoke(req acme.RevokeCertMessage) error {
	_, err := c.core.post(c.core.GetDirectory().RevokeCertURL, req, nil)
	return err
}

// ErrNoARI is returned when the server does not advertise a renewal info
// endpoint.
var ErrNoARI = errors.New("renewalInfo[get/post]: server does not advertise a renewal info endpoint")

// GetRenewalInfo GETs renewal information for a certificate from the
// renewalInfo endpoint. This is used to determine if a certificate needs to be
// renewed.
//
// Note: this endpoint is part of a draft specification, not all ACME servers
// will implement it. This method will return api.ErrNoARI if the server does
// not advertise a renewal info endpoint.
//
// https://datatracker.ietf.org/doc/draft-ietf-acme-ari
func (c *CertificateService) GetRenewalInfo(certID string) (*http.Response, error) {
	if c.core.GetDirectory().RenewalInfo == "" {
		return nil, ErrNoARI
	}
	if certID == "" {
		return nil, errors.New("renewalInfo[get]: 'certID' cannot be empty")
	}
	return c.core.HTTPClient.Get(c.core.GetDirectory().RenewalInfo + "/" + certID)
}

// PostRenewalInfo POSTs updated renewal information for a certificate to the
// renewalInfo endpoint. This is used to indicate that a certificate has been
// replaced.
//
// Note: this endpoint is part of a draft specification, not all ACME servers
// will implement it. This method will return api.ErrNoARI if the server does
// not advertise a renewal info endpoint.
//
// https://datatracker.ietf.org/doc/draft-ietf-acme-ari
func (c *CertificateService) UpdateRenewalInfo(req acme.RenewalInfoUpdateRequest) (*http.Response, error) {
	if c.core.GetDirectory().RenewalInfo == "" {
		return nil, ErrNoARI
	}
	if req.CertID == "" {
		return nil, errors.New("renewalInfo[post]: 'certID' cannot be empty")
	}

	if !req.Replaced {
		return nil, errors.New("renewalInfo[post]: 'replaced' cannot be false")
	}
	return c.core.post(c.core.GetDirectory().RenewalInfo, req, nil)
}

// get Returns the certificate and the "up" link.
func (c *CertificateService) get(certURL string, bundle bool) (*acme.RawCertificate, http.Header, error) {
	if certURL == "" {
		return nil, nil, errors.New("certificate[get]: empty URL")
	}

	resp, err := c.core.postAsGet(certURL, nil)
	if err != nil {
		return nil, nil, err
	}

	data, err := io.ReadAll(http.MaxBytesReader(nil, resp.Body, maxBodySize))
	if err != nil {
		return nil, resp.Header, err
	}

	cert := c.getCertificateChain(data, resp.Header, bundle, certURL)

	return cert, resp.Header, err
}

// getCertificateChain Returns the certificate and the issuer certificate.
func (c *CertificateService) getCertificateChain(cert []byte, headers http.Header, bundle bool, certURL string) *acme.RawCertificate {
	// Get issuerCert from bundled response from Let's Encrypt
	// See https://community.letsencrypt.org/t/acme-v2-no-up-link-in-response/64962
	_, issuer := pem.Decode(cert)
	if issuer != nil {
		// If bundle is false, we want to return a single certificate.
		// To do this, we remove the issuer cert(s) from the issued cert.
		if !bundle {
			cert = bytes.TrimSuffix(cert, issuer)
		}
		return &acme.RawCertificate{Cert: cert, Issuer: issuer}
	}

	// The issuer certificate link may be supplied via an "up" link
	// in the response headers of a new certificate.
	// See https://www.rfc-editor.org/rfc/rfc8555.html#section-7.4.2
	up := getLink(headers, "up")

	issuer, err := c.getIssuerFromLink(up)
	if err != nil {
		// If we fail to acquire the issuer cert, return the issued certificate - do not fail.
		log.Warnf("acme: Could not bundle issuer certificate [%s]: %v", certURL, err)
	} else if len(issuer) > 0 {
		// If bundle is true, we want to return a certificate bundle.
		// To do this, we append the issuer cert to the issued cert.
		if bundle {
			cert = append(cert, issuer...)
		}
	}

	return &acme.RawCertificate{Cert: cert, Issuer: issuer}
}

// getIssuerFromLink requests the issuer certificate.
func (c *CertificateService) getIssuerFromLink(up string) ([]byte, error) {
	if up == "" {
		return nil, nil
	}

	log.Infof("acme: Requesting issuer cert from %s", up)

	cert, _, err := c.get(up, false)
	if err != nil {
		return nil, err
	}

	_, err = x509.ParseCertificate(cert.Cert)
	if err != nil {
		return nil, err
	}

	return certcrypto.PEMEncode(certcrypto.DERCertificateBytes(cert.Cert)), nil
}
