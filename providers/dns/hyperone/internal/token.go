package internal

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/go-jose/go-jose/v3"
	"github.com/go-jose/go-jose/v3/jwt"
)

type TokenSigner struct {
	PrivateKey string
	KeyID      string
	Audience   string
	Issuer     string
	Subject    string
}

func (input *TokenSigner) GetJWT() (string, error) {
	signer, err := getRSASigner(input.PrivateKey, input.KeyID)
	if err != nil {
		return "", err
	}

	issuedAt := time.Now()
	expiresAt := issuedAt.Add(5 * time.Minute)

	payload := Payload{IssuedAt: issuedAt.Unix(), Expiry: expiresAt.Unix(), Audience: input.Audience, Issuer: input.Issuer, Subject: input.Subject}
	token, err := payload.buildToken(&signer)

	return token, err
}

func getRSASigner(privateKey, keyID string) (jose.Signer, error) {
	parsedKey, err := parseRSAKey(privateKey)
	if err != nil {
		return nil, err
	}

	key := jose.SigningKey{Algorithm: jose.RS256, Key: parsedKey}

	signerOpts := jose.SignerOptions{}
	signerOpts.WithType("JWT")
	signerOpts.WithHeader("kid", keyID)

	rsaSigner, err := jose.NewSigner(key, &signerOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to create JWS RSA256 signer: %w", err)
	}

	return rsaSigner, nil
}

type Payload struct {
	IssuedAt int64  `json:"iat"`
	Expiry   int64  `json:"exp"`
	Audience string `json:"aud"`
	Issuer   string `json:"iss"`
	Subject  string `json:"sub"`
}

func (payload *Payload) buildToken(signer *jose.Signer) (string, error) {
	builder := jwt.Signed(*signer).Claims(payload)

	token, err := builder.CompactSerialize()
	if err != nil {
		return "", fmt.Errorf("failed to build JWT: %w", err)
	}

	return token, nil
}

func parseRSAKey(pemString string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemString))

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return key, nil
}
