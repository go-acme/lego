package certificate

import (
	"crypto"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/xenolf/lego/emca/certificate/certcrypto"
	"github.com/xenolf/lego/emca/internal/secure"
	"github.com/xenolf/lego/emca/le"
	"github.com/xenolf/lego/log"
)

const (
	// maxBodySize is the maximum size of body that we will read.
	maxBodySize = 1024 * 1024
)

// orderResource representing an account's requests to issue certificates.
type orderResource struct {
	le.OrderMessage `json:"body,omitempty"`
	URL             string   `json:"url,omitempty"`
	Domains         []string `json:"domains,omitempty"`
}

// Resource represents a CA issued certificate.
// PrivateKey, Certificate and IssuerCertificate are all
// already PEM encoded and can be directly written to disk.
// Certificate may be a certificate bundle,
// depending on the options supplied to create it.
type Resource struct {
	Domain            string `json:"domain"`
	CertURL           string `json:"certUrl"`
	CertStableURL     string `json:"certStableUrl"`
	AccountRef        string `json:"accountRef,omitempty"`
	PrivateKey        []byte `json:"-"`
	Certificate       []byte `json:"-"`
	IssuerCertificate []byte `json:"-"`
	CSR               []byte `json:"-"`
}

type resolver interface {
	Solve(authorizations []le.Authorization) error
}

type Certifier struct {
	jws       *secure.JWS
	keyType   certcrypto.KeyType
	directory le.Directory
	resolver  resolver
}

func NewCertifier(jws *secure.JWS, keyType certcrypto.KeyType, directory le.Directory, resolver resolver) *Certifier {
	return &Certifier{
		jws:       jws,
		keyType:   keyType,
		directory: directory,
		resolver:  resolver,
	}
}

// Obtain tries to obtain a single certificate using all domains passed into it.
// The first domain in domains is used for the CommonName field of the certificate,
// all other domains are added using the Subject Alternate Names extension.
// A new private key is generated for every invocation of this function.
// If you do not want that you can supply your own private key in the privKey parameter.
// If this parameter is non-nil it will be used instead of generating a new one.
// If bundle is true, the []byte contains both the issuer certificate and your issued certificate as a bundle.
// This function will never return a partial certificate.
// If one domain in the list fails, the whole certificate will fail.
func (c *Certifier) Obtain(domains []string, bundle bool, privKey crypto.PrivateKey, mustStaple bool) (*Resource, error) {
	if len(domains) == 0 {
		return nil, errors.New("no domains to obtain a certificate for")
	}

	if bundle {
		log.Infof("[%s] acme: Obtaining bundled SAN certificate", strings.Join(domains, ", "))
	} else {
		log.Infof("[%s] acme: Obtaining SAN certificate", strings.Join(domains, ", "))
	}

	order, err := c.createOrderForIdentifiers(domains)
	if err != nil {
		return nil, err
	}

	authz, err := c.getAuthzForOrder(order)
	if err != nil {
		// If any challenge fails, return. Do not generate partial SAN certificates.
		for _, auth := range order.Authorizations {
			errD := c.disableAuthz(auth)
			if errD != nil {
				log.Infof("unable to deactivated authorizations: %s", auth)
			}
		}
		return nil, err
	}

	err = c.resolver.Solve(authz)
	if err != nil {
		// If any challenge fails, return. Do not generate partial SAN certificates.
		return nil, err
	}

	log.Infof("[%s] acme: Validations succeeded; requesting certificates", strings.Join(domains, ", "))

	failures := make(obtainError)
	cert, err := c.createForOrder(order, bundle, privKey, mustStaple)
	if err != nil {
		for _, auth := range authz {
			failures[auth.Identifier.Value] = err
		}
	}

	// Do not return an empty failures map, because
	// it would still be a non-nil error value
	if len(failures) > 0 {
		return cert, failures
	}
	return cert, nil
}

// ObtainForCSR tries to obtain a certificate matching the CSR passed into it.
// The domains are inferred from the CommonName and SubjectAltNames, if any.
// The private key for this CSR is not required.
// If bundle is true, the []byte contains both the issuer certificate and your issued certificate as a bundle.
// This function will never return a partial certificate. If one domain in the list fails,
// the whole certificate will fail.
func (c *Certifier) ObtainForCSR(csr x509.CertificateRequest, bundle bool) (*Resource, error) {
	// figure out what domains it concerns
	// start with the common name
	domains := []string{csr.Subject.CommonName}

	// loop over the SubjectAltName DNS names
	for _, sanName := range csr.DNSNames {
		if containsSAN(domains, sanName) {
			// Duplicate; skip this name
			continue
		}

		// Name is unique
		domains = append(domains, sanName)
	}

	if bundle {
		log.Infof("[%s] acme: Obtaining bundled SAN certificate given a CSR", strings.Join(domains, ", "))
	} else {
		log.Infof("[%s] acme: Obtaining SAN certificate given a CSR", strings.Join(domains, ", "))
	}

	order, err := c.createOrderForIdentifiers(domains)
	if err != nil {
		return nil, err
	}
	authz, err := c.getAuthzForOrder(order)
	if err != nil {
		// If any challenge fails, return. Do not generate partial SAN certificates.
		for _, auth := range order.Authorizations {
			errD := c.disableAuthz(auth)
			if errD != nil {
				log.Infof("unable to deactivated authorizations: %s", auth)
			}
		}
		return nil, err
	}

	err = c.resolver.Solve(authz)
	if err != nil {
		// If any challenge fails, return. Do not generate partial SAN certificates.
		return nil, err
	}

	log.Infof("[%s] acme: Validations succeeded; requesting certificates", strings.Join(domains, ", "))

	failures := make(obtainError)
	cert, err := c.createForCSR(order, bundle, csr.Raw, nil)
	if err != nil {
		for _, chln := range authz {
			failures[chln.Identifier.Value] = err
		}
	}

	if cert != nil {
		// Add the CSR to the certificate so that it can be used for renewals.
		cert.CSR = certcrypto.PEMEncode(&csr)
	}

	// Do not return an empty failures map,
	// because it would still be a non-nil error value
	if len(failures) > 0 {
		return cert, failures
	}
	return cert, nil
}

func (c *Certifier) createForOrder(order orderResource, bundle bool, privKey crypto.PrivateKey, mustStaple bool) (*Resource, error) {
	if privKey == nil {
		var err error
		privKey, err = certcrypto.GeneratePrivateKey(c.keyType)
		if err != nil {
			return nil, err
		}
	}

	// Determine certificate name(s) based on the authorization resources
	commonName := order.Domains[0]

	// ACME draft Section 7.4 "Applying for Certificate Issuance"
	// https://tools.ietf.org/html/draft-ietf-acme-acme-12#section-7.4
	// says:
	//   Clients SHOULD NOT make any assumptions about the sort order of
	//   "identifiers" or "authorizations" elements in the returned order
	//   object.
	san := []string{commonName}
	for _, auth := range order.Identifiers {
		if auth.Value != commonName {
			san = append(san, auth.Value)
		}
	}

	// TODO: should the CSR be customizable?
	csr, err := certcrypto.GenerateCsr(privKey, commonName, san, mustStaple)
	if err != nil {
		return nil, err
	}

	return c.createForCSR(order, bundle, csr, certcrypto.PEMEncode(privKey))
}

func (c *Certifier) createForCSR(order orderResource, bundle bool, csr []byte, privateKeyPem []byte) (*Resource, error) {
	csrString := base64.RawURLEncoding.EncodeToString(csr)

	var retOrder le.OrderMessage
	_, err := c.jws.PostJSON(order.Finalize, le.CSRMessage{Csr: csrString}, &retOrder)
	if err != nil {
		return nil, err
	}

	if retOrder.Status == le.StatusInvalid {
		return nil, err
	}

	commonName := order.Domains[0]
	certRes := Resource{
		Domain:     commonName,
		CertURL:    retOrder.Certificate,
		PrivateKey: privateKeyPem,
	}

	if retOrder.Status == le.StatusValid {
		// if the certificate is available right away, short cut!
		ok, err := c.checkResponse(retOrder, &certRes, bundle)
		if err != nil {
			return nil, err
		}

		if ok {
			return &certRes, nil
		}
	}

	stopTimer := time.NewTimer(30 * time.Second)
	defer stopTimer.Stop()
	retryTick := time.NewTicker(500 * time.Millisecond)
	defer retryTick.Stop()

	for {
		select {
		case <-stopTimer.C:
			return nil, errors.New("certificate polling timed out")
		case <-retryTick.C:
			_, err := c.jws.PostAsGet(order.URL, &retOrder)
			if err != nil {
				return nil, err
			}

			done, err := c.checkResponse(retOrder, &certRes, bundle)
			if err != nil {
				return nil, err
			}
			if done {
				return &certRes, nil
			}
		}
	}
}

// Revoke takes a PEM encoded certificate or bundle and tries to revoke it at the CA.
func (c *Certifier) Revoke(cert []byte) error {
	certificates, err := certcrypto.ParsePEMBundle(cert)
	if err != nil {
		return err
	}

	x509Cert := certificates[0]
	if x509Cert.IsCA {
		return fmt.Errorf("certificate bundle starts with a CA certificate")
	}

	encodedCert := base64.URLEncoding.EncodeToString(x509Cert.Raw)

	_, err = c.jws.PostJSON(c.directory.RevokeCertURL, le.RevokeCertMessage{Certificate: encodedCert}, nil)
	return err
}

// Renew takes a Resource and tries to renew the certificate.
// If the renewal process succeeds, the new certificate will ge returned in a new CertResource.
// Please be aware that this function will return a new certificate in ANY case that is not an error.
// If the server does not provide us with a new cert on a GET request to the CertURL
// this function will start a new-cert flow where a new certificate gets generated.
// If bundle is true, the []byte contains both the issuer certificate and your issued certificate as a bundle.
// For private key reuse the PrivateKey property of the passed in Resource should be non-nil.
func (c *Certifier) Renew(cert Resource, bundle, mustStaple bool) (*Resource, error) {
	// Input certificate is PEM encoded. Decode it here as we may need the decoded
	// cert later on in the renewal process. The input may be a bundle or a single certificate.
	certificates, err := certcrypto.ParsePEMBundle(cert.Certificate)
	if err != nil {
		return nil, err
	}

	x509Cert := certificates[0]
	if x509Cert.IsCA {
		return nil, fmt.Errorf("[%s] Certificate bundle starts with a CA certificate", cert.Domain)
	}

	// This is just meant to be informal for the user.
	timeLeft := x509Cert.NotAfter.Sub(time.Now().UTC())
	log.Infof("[%s] acme: Trying renewal with %d hours remaining", cert.Domain, int(timeLeft.Hours()))

	// We always need to request a new certificate to renew.
	// Start by checking to see if the certificate was based off a CSR,
	// and use that if it's defined.
	if len(cert.CSR) > 0 {
		csr, errP := certcrypto.PemDecodeTox509CSR(cert.CSR)
		if errP != nil {
			return nil, errP
		}

		newCert, failures := c.ObtainForCSR(*csr, bundle)
		return newCert, failures
	}

	var privKey crypto.PrivateKey
	if cert.PrivateKey != nil {
		privKey, err = certcrypto.ParsePEMPrivateKey(cert.PrivateKey)
		if err != nil {
			return nil, err
		}
	}

	var domains []string
	// Check for SAN certificate
	if len(x509Cert.DNSNames) > 1 {
		domains = append(domains, x509Cert.Subject.CommonName)
		for _, sanDomain := range x509Cert.DNSNames {
			if sanDomain == x509Cert.Subject.CommonName {
				continue
			}
			domains = append(domains, sanDomain)
		}
	} else {
		domains = append(domains, x509Cert.Subject.CommonName)
	}

	return c.Obtain(domains, bundle, privKey, mustStaple)
}

func (c *Certifier) createOrderForIdentifiers(domains []string) (orderResource, error) {
	var identifiers []le.Identifier
	for _, domain := range domains {
		identifiers = append(identifiers, le.Identifier{Type: "dns", Value: domain})
	}

	order := le.OrderMessage{Identifiers: identifiers}

	var response le.OrderMessage
	hdr, err := c.jws.PostJSON(c.directory.NewOrderURL, order, &response)
	if err != nil {
		return orderResource{}, err
	}

	return orderResource{
		URL:          hdr.Get("Location"),
		Domains:      domains,
		OrderMessage: response,
	}, nil
}

// checkResponse checks to see if the certificate is ready and a link is contained in the response.
// If so, loads it into certRes and returns true.
// If the cert is not yet ready, it returns false.
// The certRes input should already have the Domain (common name) field populated.
// If bundle is true, the certificate will be bundled with the issuer's cert.
func (c *Certifier) checkResponse(order le.OrderMessage, certRes *Resource, bundle bool) (bool, error) {
	switch order.Status {
	// TODO extract function?
	case le.StatusValid:
		resp, err := c.jws.PostAsGet(order.Certificate, nil)
		if err != nil {
			return false, err
		}

		cert, err := ioutil.ReadAll(http.MaxBytesReader(nil, resp.Body, maxBodySize))
		if err != nil {
			return false, err
		}

		// The issuer certificate link may be supplied via an "up" link
		// in the response headers of a new certificate.
		// See https://tools.ietf.org/html/draft-ietf-acme-acme-12#section-7.4.2
		links := parseLinks(resp.Header["Link"])
		if link, ok := links["up"]; ok {
			issuerCert, err := c.getIssuerCertificateFromLink(link)

			if err != nil {
				// If we fail to acquire the issuer cert, return the issued certificate - do not fail.
				log.Warnf("[%s] acme: Could not bundle issuer certificate: %v", certRes.Domain, err)
			} else {
				issuerCert = certcrypto.PEMEncode(certcrypto.DERCertificateBytes(issuerCert))

				// If bundle is true, we want to return a certificate bundle.
				// To do this, we append the issuer cert to the issued cert.
				if bundle {
					cert = append(cert, issuerCert...)
				}

				certRes.IssuerCertificate = issuerCert
			}
		} else {
			// Get issuerCert from bundled response from Let's Encrypt
			// See https://community.letsencrypt.org/t/acme-v2-no-up-link-in-response/64962
			_, rest := pem.Decode(cert)
			if rest != nil {
				certRes.IssuerCertificate = rest
			}
		}

		certRes.Certificate = cert
		certRes.CertURL = order.Certificate
		certRes.CertStableURL = order.Certificate

		log.Infof("[%s] Server responded with a certificate.", certRes.Domain)
		return true, nil
	case le.StatusInvalid:
		return false, errors.New("order has invalid state: invalid")
	default:
		return false, nil
	}
}

// getIssuerCertificateFromLink requests the issuer certificate
func (c *Certifier) getIssuerCertificateFromLink(url string) ([]byte, error) {
	log.Infof("acme: Requesting issuer cert from %s", url)

	resp, err := c.jws.PostAsGet(url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	issuerRaw, err := ioutil.ReadAll(http.MaxBytesReader(nil, resp.Body, maxBodySize))
	if err != nil {
		return nil, err
	}

	_, err = x509.ParseCertificate(issuerRaw)
	if err != nil {
		return nil, err
	}

	return issuerRaw, err
}

func parseLinks(links []string) map[string]string {
	aBrkt := regexp.MustCompile("[<>]")
	slver := regexp.MustCompile("(.+) *= *\"(.+)\"")
	linkMap := make(map[string]string)

	for _, link := range links {

		link = aBrkt.ReplaceAllString(link, "")
		parts := strings.Split(link, ";")

		matches := slver.FindStringSubmatch(parts[1])
		if len(matches) > 0 {
			linkMap[matches[2]] = parts[0]
		}
	}

	return linkMap
}

func containsSAN(domains []string, sanName string) bool {
	for _, existingName := range domains {
		if existingName == sanName {
			return true
		}
	}
	return false
}
