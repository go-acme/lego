package acme

import (
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"log"
)

type tlsSNI02Challenge struct {
	jws      *jws
	validate validateFunc
	provider ChallengeProvider
}

func (t *tlsSNI02Challenge) Solve(chlng challenge, domain string) error {
	logf("[INFO][%s] acme: Trying to solve TLS-SNI-02", domain)

	// Generate the Key Authorization for the challenge
	keyAuth, err := getKeyAuthorization(chlng.Token, t.jws.privKey)
	if err != nil {
		return err
	}

	err = t.provider.Present(domain, chlng.Token, keyAuth)
	if err != nil {
		return fmt.Errorf("[%s] error presenting token: %v", domain, err)
	}
	defer func() {
		err := t.provider.CleanUp(domain, chlng.Token, keyAuth)
		if err != nil {
			log.Printf("[%s] error cleaning up: %v", domain, err)
		}
	}()
	return t.validate(t.jws, domain, chlng.URI, challenge{Resource: "challenge", Type: chlng.Type, Token: chlng.Token, KeyAuthorization: keyAuth})
}

// TLSSNI02ChallengeCert returns a certificate for the `tls-sni-02` challenge
func TLSSNI02ChallengeCert(token, keyAuth string) (tls.Certificate, error) {

	// Construct SanA value from token sha256
	tokenShaBytes := sha256.Sum256([]byte(token))
	tokenSha := hex.EncodeToString(tokenShaBytes[:sha256.Size])
	sanA := fmt.Sprintf("%s.%s.token.acme.invalid", tokenSha[:32], tokenSha[32:])

	// Construct SanB value from keyAuth sha256
	keyAuthShaBytes := sha256.Sum256([]byte(keyAuth))
	keyAuthSha := hex.EncodeToString(keyAuthShaBytes[:sha256.Size])
	sanB := fmt.Sprintf("%s.%s.ka.acme.invalid", keyAuthSha[:32], keyAuthSha[32:])

	// generate a new RSA key for the certificate
	tempPrivKey, err := generatePrivateKey(RSA2048)
	if err != nil {
		return tls.Certificate{}, err
	}
	rsaPrivKey := tempPrivKey.(*rsa.PrivateKey)
	rsaPrivPEM := pemEncode(rsaPrivKey)

	tempCertPEM, err := generatePemCert(rsaPrivKey, sanA, sanB)
	if err != nil {
		return tls.Certificate{}, err
	}

	certificate, err := tls.X509KeyPair(tempCertPEM, rsaPrivPEM)
	if err != nil {
		return tls.Certificate{}, err
	}

	return certificate, nil
}
