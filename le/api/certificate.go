package api

import (
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/xenolf/lego/le"
)

// maxBodySize is the maximum size of body that we will read.
const maxBodySize = 1024 * 1024

type CertificateService service

// Get Returns the certificate and the "up" link.
func (c *CertificateService) Get(certURL string) ([]byte, string, error) {
	if len(certURL) == 0 {
		return nil, "", errors.New("certificate[get]: empty URL")
	}

	resp, err := c.core.postAsGet(certURL, nil)
	if err != nil {
		return nil, "", err
	}

	cert, err := ioutil.ReadAll(http.MaxBytesReader(nil, resp.Body, maxBodySize))
	if err != nil {
		return nil, "", err
	}

	// The issuer certificate link may be supplied via an "up" link
	// in the response headers of a new certificate.
	// See https://tools.ietf.org/html/draft-ietf-acme-acme-12#section-7.4.2
	up := getLink(resp.Header, "up")

	return cert, up, err
}

// Revoke Revokes a certificate.
func (c *CertificateService) Revoke(req le.RevokeCertMessage) error {
	_, err := c.core.post(c.core.GetDirectory().RevokeCertURL, req, nil)
	return err
}
