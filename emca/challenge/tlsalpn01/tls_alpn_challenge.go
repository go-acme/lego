package tlsalpn01

import (
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509/pkix"
	"encoding/asn1"
	"fmt"

	"github.com/xenolf/lego/emca/certificate/certcrypto"
	"github.com/xenolf/lego/emca/challenge"
	"github.com/xenolf/lego/emca/internal/secure"
	"github.com/xenolf/lego/emca/le"
	"github.com/xenolf/lego/log"
)

// idPeAcmeIdentifierV1 is the SMI Security for PKIX Certification Extension OID referencing the ACME extension.
// Reference: https://tools.ietf.org/html/draft-ietf-acme-tls-alpn-05#section-5.1
var idPeAcmeIdentifierV1 = asn1.ObjectIdentifier{1, 3, 6, 1, 5, 5, 7, 1, 31}

// FIXME refactor
type validateFunc func(j *secure.JWS, domain, uri string, chlng le.Challenge) error

type Challenge struct {
	jws      *secure.JWS
	validate validateFunc
	provider challenge.Provider
}

func NewChallenge(jws *secure.JWS, validate validateFunc, provider challenge.Provider) *Challenge {
	return &Challenge{
		jws:      jws,
		validate: validate,
		provider: provider,
	}
}

func (c *Challenge) SetProvider(provider challenge.Provider) {
	c.provider = provider
}

// Solve manages the provider to validate and solve the challenge.
func (c *Challenge) Solve(chlng le.Challenge, domain string) error {
	log.Infof("[%s] acme: Trying to solve TLS-ALPN-01", domain)

	// Generate the Key Authorization for the challenge
	keyAuth, err := c.jws.GetKeyAuthorization(chlng.Token)
	if err != nil {
		return err
	}

	err = c.provider.Present(domain, chlng.Token, keyAuth)
	if err != nil {
		return fmt.Errorf("[%s] error presenting token: %v", domain, err)
	}
	defer func() {
		err := c.provider.CleanUp(domain, chlng.Token, keyAuth)
		if err != nil {
			log.Warnf("[%s] error cleaning up: %v", domain, err)
		}
	}()

	return c.validate(c.jws, domain, chlng.URL, le.Challenge{Type: chlng.Type, Token: chlng.Token, KeyAuthorization: keyAuth})
}

// ChallengeBlocks returns PEM blocks (certPEMBlock, keyPEMBlock) with the acmeValidation-v1 extension
// and domain name for the `tls-alpn-01` challenge.
func ChallengeBlocks(domain, keyAuth string) ([]byte, []byte, error) {
	// Compute the SHA-256 digest of the key authorization.
	zBytes := sha256.Sum256([]byte(keyAuth))

	value, err := asn1.Marshal(zBytes[:sha256.Size])
	if err != nil {
		return nil, nil, err
	}

	// Add the keyAuth digest as the acmeValidation-v1 extension
	// (marked as critical such that it won't be used by non-ACME software).
	// Reference: https://tools.ietf.org/html/draft-ietf-acme-tls-alpn-05#section-3
	extensions := []pkix.Extension{
		{
			Id:       idPeAcmeIdentifierV1,
			Critical: true,
			Value:    value,
		},
	}

	// Generate a new RSA key for the certificates.
	tempPrivKey, err := certcrypto.GeneratePrivateKey(certcrypto.RSA2048)
	if err != nil {
		return nil, nil, err
	}

	rsaPrivKey := tempPrivKey.(*rsa.PrivateKey)

	// Generate the PEM certificate using the provided private key, domain, and extra extensions.
	tempCertPEM, err := certcrypto.GeneratePemCert(rsaPrivKey, domain, extensions)
	if err != nil {
		return nil, nil, err
	}

	// Encode the private key into a PEM format. We'll need to use it to generate the x509 keypair.
	rsaPrivPEM := certcrypto.PEMEncode(rsaPrivKey)

	return tempCertPEM, rsaPrivPEM, nil
}

// ChallengeCert returns a certificate with the acmeValidation-v1 extension
// and domain name for the `tls-alpn-01` challenge.
func ChallengeCert(domain, keyAuth string) (*tls.Certificate, error) {
	tempCertPEM, rsaPrivPEM, err := ChallengeBlocks(domain, keyAuth)
	if err != nil {
		return nil, err
	}

	cert, err := tls.X509KeyPair(tempCertPEM, rsaPrivPEM)
	if err != nil {
		return nil, err
	}

	return &cert, nil
}
