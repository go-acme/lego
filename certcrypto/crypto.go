package certcrypto

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"errors"
	"math/big"
	"net"
	"slices"
	"time"

	"golang.org/x/crypto/ocsp"
)

const (
	// OCSPGood means that the certificate is valid.
	OCSPGood = ocsp.Good
	// OCSPRevoked means that the certificate has been deliberately revoked.
	OCSPRevoked = ocsp.Revoked
	// OCSPUnknown means that the OCSP responder doesn't know about the certificate.
	OCSPUnknown = ocsp.Unknown
	// OCSPServerFailed means that the OCSP responder failed to process the request.
	OCSPServerFailed = ocsp.ServerFailed
)

// Constants for OCSP must staple.
var (
	tlsFeatureExtensionOID = asn1.ObjectIdentifier{1, 3, 6, 1, 5, 5, 7, 1, 24}
	ocspMustStapleFeature  = []byte{0x30, 0x03, 0x02, 0x01, 0x05}
)

type DERCertificateBytes []byte

// ParsePEMBundle parses a certificate bundle from top to bottom and returns
// a slice of x509 certificates. This function will error if no certificates are found.
func ParsePEMBundle(bundle []byte) ([]*x509.Certificate, error) {
	var (
		certificates []*x509.Certificate
		certDERBlock *pem.Block
	)

	for {
		certDERBlock, bundle = pem.Decode(bundle)
		if certDERBlock == nil {
			break
		}

		if certDERBlock.Type == "CERTIFICATE" {
			cert, err := x509.ParseCertificate(certDERBlock.Bytes)
			if err != nil {
				return nil, err
			}

			certificates = append(certificates, cert)
		}
	}

	if len(certificates) == 0 {
		return nil, errors.New("no certificates were found while parsing the bundle")
	}

	return certificates, nil
}

type CSROptions struct {
	Domain         string
	SAN            []string
	MustStaple     bool
	EmailAddresses []string
}

func CreateCSR(privateKey crypto.Signer, opts CSROptions) ([]byte, error) {
	var (
		dnsNames    []string
		ipAddresses []net.IP
	)

	for _, altname := range opts.SAN {
		if ip := net.ParseIP(altname); ip != nil {
			ipAddresses = append(ipAddresses, ip)
		} else {
			dnsNames = append(dnsNames, altname)
		}
	}

	template := x509.CertificateRequest{
		Subject:        pkix.Name{CommonName: opts.Domain},
		DNSNames:       dnsNames,
		EmailAddresses: opts.EmailAddresses,
		IPAddresses:    ipAddresses,
	}

	if opts.MustStaple {
		template.ExtraExtensions = append(template.ExtraExtensions, pkix.Extension{
			Id:    tlsFeatureExtensionOID,
			Value: ocspMustStapleFeature,
		})
	}

	return x509.CreateCertificateRequest(rand.Reader, &template, privateKey)
}

func PEMEncode(data any) []byte {
	return pem.EncodeToMemory(PEMBlock(data))
}

func pemDecode(data []byte) (*pem.Block, error) {
	pemBlock, _ := pem.Decode(data)
	if pemBlock == nil {
		return nil, errors.New("PEM decode did not yield a valid block. Is the certificate in the right format?")
	}

	return pemBlock, nil
}

func PemDecodeTox509CSR(data []byte) (*x509.CertificateRequest, error) {
	pemBlock, err := pemDecode(data)
	if pemBlock == nil {
		return nil, err
	}

	if pemBlock.Type != "CERTIFICATE REQUEST" && pemBlock.Type != "NEW CERTIFICATE REQUEST" {
		return nil, errors.New("PEM block is not a certificate request")
	}

	return x509.ParseCertificateRequest(pemBlock.Bytes)
}

// ParsePEMCertificate returns Certificate from a PEM encoded certificate.
// The certificate has to be PEM encoded. Any other encodings like DER will fail.
func ParsePEMCertificate(cert []byte) (*x509.Certificate, error) {
	pemBlock, err := pemDecode(cert)
	if pemBlock == nil {
		return nil, err
	}

	// from a DER encoded certificate
	return x509.ParseCertificate(pemBlock.Bytes)
}

func GetCertificateMainDomain(cert *x509.Certificate) (string, error) {
	return getMainDomain(cert.Subject, cert.DNSNames, cert.IPAddresses)
}

func GetCSRMainDomain(cert *x509.CertificateRequest) (string, error) {
	return getMainDomain(cert.Subject, cert.DNSNames, cert.IPAddresses)
}

func getMainDomain(subject pkix.Name, dnsNames []string, ips []net.IP) (string, error) {
	if subject.CommonName == "" && len(dnsNames) == 0 && len(ips) == 0 {
		return "", errors.New("missing domain")
	}

	if subject.CommonName != "" {
		return subject.CommonName, nil
	}

	if len(dnsNames) > 0 {
		return dnsNames[0], nil
	}

	return ips[0].String(), nil
}

func ExtractDomains(cert *x509.Certificate) []string {
	var domains []string
	if cert.Subject.CommonName != "" {
		domains = append(domains, cert.Subject.CommonName)
	}

	// Check for SAN certificate
	for _, sanDomain := range cert.DNSNames {
		if sanDomain == cert.Subject.CommonName {
			continue
		}

		domains = append(domains, sanDomain)
	}

	commonNameIP := net.ParseIP(cert.Subject.CommonName)
	for _, sanIP := range cert.IPAddresses {
		if !commonNameIP.Equal(sanIP) {
			domains = append(domains, sanIP.String())
		}
	}

	return domains
}

func ExtractDomainsCSR(csr *x509.CertificateRequest) []string {
	var domains []string
	if csr.Subject.CommonName != "" {
		domains = append(domains, csr.Subject.CommonName)
	}

	// loop over the SubjectAltName DNS names
	for _, sanName := range csr.DNSNames {
		if slices.Contains(domains, sanName) {
			// Duplicate; skip this name
			continue
		}

		// Name is unique
		domains = append(domains, sanName)
	}

	cnip := net.ParseIP(csr.Subject.CommonName)
	for _, sanIP := range csr.IPAddresses {
		if !cnip.Equal(sanIP) {
			domains = append(domains, sanIP.String())
		}
	}

	return domains
}

func GeneratePemCert(privateKey *rsa.PrivateKey, domain string, extensions []pkix.Extension) ([]byte, error) {
	derBytes, err := generateDerCert(privateKey, time.Time{}, domain, extensions)
	if err != nil {
		return nil, err
	}

	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes}), nil
}

func generateDerCert(privateKey *rsa.PrivateKey, expiration time.Time, domain string, extensions []pkix.Extension) ([]byte, error) {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)

	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, err
	}

	if expiration.IsZero() {
		expiration = time.Now().AddDate(1, 0, 0)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: "ACME Challenge TEMP",
		},
		NotBefore: time.Now(),
		NotAfter:  expiration,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		ExtraExtensions:       extensions,
	}

	// handling SAN filling as type suspected
	if ip := net.ParseIP(domain); ip != nil {
		template.IPAddresses = []net.IP{ip}
	} else {
		template.DNSNames = []string{domain}
	}

	return x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
}
