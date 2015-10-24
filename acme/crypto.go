package acme

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/binary"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"time"

	"golang.org/x/crypto/ocsp"
	"golang.org/x/crypto/sha3"
)

type keyType int
type derCertificateBytes []byte

const (
	eckey keyType = iota
	rsakey
)

// GetOCSPForCert takes a PEM encoded cert or cert bundle and returns a OCSP
// response from the OCSP endpoint in the certificate.
// This []byte can be passed directly  into the OCSPStaple property of a tls.Certificate.
// If the bundle only contains the issued certificate, this function will try
// to get the issuer certificate from the IssuingCertificateURL in the certificate.
func GetOCSPForCert(bundle []byte) ([]byte, error) {
	certificates, err := parsePEMBundle(bundle)
	if err != nil {
		return nil, err
	}

	// We only got one certificate, means we have no issuer certificate - get it.
	if len(certificates) == 1 {
		// TODO: build fallback. If this fails, check the remaining array entries.
		resp, err := http.Get(certificates[0].IssuingCertificateURL[0])
		if err != nil {
			return nil, err
		}

		issuerBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		issuerCert, err := x509.ParseCertificate(issuerBytes)
		if err != nil {
			return nil, err
		}

		// Insert it into the slice on position 0
		// We want it ordered right CA -> CRT
		certificates = append(certificates, nil)
		copy(certificates[1:], certificates[0:])
		certificates[0] = issuerCert
	}

	// We expect the certificate slice to be ordered downwards the chain.
	// CA -> CRT. We need to pull the cert and issuer cert out of it, which should
	// always be the last two certificates.
	issuedCert := certificates[len(certificates)-1]
	issuerCert := certificates[len(certificates)-2]

	// Finally kick off the OCSP request.
	ocspReq, err := ocsp.CreateRequest(issuedCert, issuerCert, nil)
	if err != nil {
		return nil, err
	}

	reader := bytes.NewReader(ocspReq)
	req, err := http.Post(issuedCert.OCSPServer[0], "application/ocsp-request", reader)
	if err != nil {
		return nil, err
	}

	ocspResBytes, err := ioutil.ReadAll(req.Body)
	_, err = ocsp.ParseResponse(ocspResBytes, nil)
	if err != nil {
		log.Printf("OCSPParse Error: %v", err)
		return nil, err
	}

	return ocspResBytes, nil
}

// Derive the shared secret according to acme spec 5.6
func performECDH(priv *ecdsa.PrivateKey, pub *ecdsa.PublicKey, outLen int, label string) []byte {
	// Derive Z from the private and public keys according to SEC 1 Ver. 2.0 - 3.3.1
	Z, _ := priv.PublicKey.ScalarMult(pub.X, pub.Y, priv.D.Bytes())

	if len(Z.Bytes())+len(label)+4 > 384 {
		return nil
	}

	if outLen < 384*(2^32-1) {
		return nil
	}

	// Derive the shared secret key using the ANS X9.63 KDF - SEC 1 Ver. 2.0 - 3.6.1
	hasher := sha3.New384()
	buffer := make([]byte, outLen)
	bufferLen := 0
	for i := 0; i < outLen/384; i++ {
		hasher.Reset()

		// Ki = Hash(Z || Counter || [SharedInfo])
		hasher.Write(Z.Bytes())
		binary.Write(hasher, binary.BigEndian, i)
		hasher.Write([]byte(label))

		hash := hasher.Sum(nil)
		copied := copy(buffer[bufferLen:], hash)
		bufferLen += copied
	}

	return buffer
}

// parsePEMBundle parses a certificate bundle from top to bottom and returns
// a slice of x509 certificates. This function will error if no certificates are found.
func parsePEMBundle(bundle []byte) ([]*x509.Certificate, error) {
	var certificates []*x509.Certificate

	remaining := bundle
	for len(remaining) != 0 {
		certBlock, rem := pem.Decode(remaining)
		// Thanks golang for having me do this :[
		remaining = rem
		if certBlock == nil {
			return nil, errors.New("Could not decode certificate.")
		}

		cert, err := x509.ParseCertificate(certBlock.Bytes)
		if err != nil {
			return nil, err
		}

		certificates = append(certificates, cert)
	}

	if len(certificates) == 0 {
		return nil, errors.New("No certificates were found while parsing the bundle.")
	}

	return certificates, nil
}

func generatePrivateKey(t keyType, keyLength int) (crypto.PrivateKey, error) {
	switch t {
	case eckey:
		return ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	case rsakey:
		return rsa.GenerateKey(rand.Reader, keyLength)
	}

	return nil, fmt.Errorf("Invalid keytype: %d", t)
}

func generateCsr(privateKey *rsa.PrivateKey, domain string) ([]byte, error) {
	template := x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName: domain,
		},
	}

	return x509.CreateCertificateRequest(rand.Reader, &template, privateKey)
}

func pemEncode(data interface{}) []byte {
	var pemBlock *pem.Block
	switch key := data.(type) {
	case *rsa.PrivateKey:
		pemBlock = &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}
		break
	case derCertificateBytes:
		pemBlock = &pem.Block{Type: "CERTIFICATE", Bytes: []byte(data.(derCertificateBytes))}
	}

	return pem.EncodeToMemory(pemBlock)
}

func pemDecode(data []byte) (*pem.Block, error) {
	pemBlock, _ := pem.Decode(data)
	if pemBlock == nil {
		return nil, fmt.Errorf("Pem decode did not yield a valid block. Is the certificate in the right format?")
	}

	return pemBlock, nil
}

func pemDecodeTox509(pem []byte) (*x509.Certificate, error) {
	pemBlock, err := pemDecode(pem)
	if pemBlock == nil {
		return nil, err
	}

	return x509.ParseCertificate(pemBlock.Bytes)
}

// GetPEMCertExpiration returns the "NotAfter" date of a PEM encoded certificate.
// The certificate has to be PEM encoded. Any other encodings like DER will fail.
func GetPEMCertExpiration(cert []byte) (time.Time, error) {
	pemBlock, err := pemDecode(cert)
	if pemBlock == nil {
		return time.Time{}, err
	}

	return getCertExpiration(pemBlock.Bytes)
}

// getCertExpiration returns the "NotAfter" date of a DER encoded certificate.
func getCertExpiration(cert []byte) (time.Time, error) {
	pCert, err := x509.ParseCertificate(cert)
	if err != nil {
		return time.Time{}, err
	}

	return pCert.NotAfter, nil
}

func generatePemCert(privKey *rsa.PrivateKey, domain string) ([]byte, error) {
	derBytes, err := generateDerCert(privKey, time.Time{}, domain)
	if err != nil {
		return nil, err
	}

	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes}), nil
}

func generateDerCert(privKey *rsa.PrivateKey, expiration time.Time, domain string) ([]byte, error) {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, err
	}

	if expiration.IsZero() {
		expiration = time.Now().Add(365)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: "ACME Challenge TEMP",
		},
		NotBefore: time.Now(),
		NotAfter:  expiration,

		KeyUsage:              x509.KeyUsageKeyEncipherment,
		BasicConstraintsValid: true,
		DNSNames:              []string{domain},
	}

	return x509.CreateCertificate(rand.Reader, &template, &template, &privKey.PublicKey, privKey)
}
