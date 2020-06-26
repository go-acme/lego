package internal

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"

	jose "gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
)

type TokenSigner struct {
	PrivateKey string
	KeyID      string
	Audience   string
	Issuer     string
	Subject    string
}

type Payload struct {
	IssuedAt int64  `json:"iat"`
	Expiry   int64  `json:"exp"`
	Audience string `json:"aud"`
	Issuer   string `json:"iss"`
	Subject  string `json:"sub"`
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
		return nil, fmt.Errorf("Failed to create JWS RSA256 signer:%+v", err)
	}
	return rsaSigner, nil
}

func (payload *Payload) buildToken(signer *jose.Signer) (string, error) {
	builder := jwt.Signed(*signer)
	builder = builder.Claims(payload)
	token, err := builder.CompactSerialize()
	if err != nil {
		return "", fmt.Errorf("Failed to build JWT:%+v", err)
	}
	return token, nil
}

func (input *TokenSigner) GetJWT() (string, error) {
	signer, err := getRSASigner(input.PrivateKey, input.KeyID)
	if err != nil {
		return "", err
	}

	var expiryTime int64 = 60 * 5
	issuedAt, expiresAt := getTokenTimings(expiryTime)

	payload := Payload{IssuedAt: issuedAt, Expiry: expiresAt, Audience: input.Audience, Issuer: input.Issuer, Subject: input.Subject}
	token, err := payload.buildToken(&signer)

	return token, err
}

func getTokenTimings(expiryTime int64) (iat int64, exp int64) {
	now := time.Now()
	issuedAt := now.Unix()
	expiresAt := issuedAt + expiryTime
	return issuedAt, expiresAt
}

func parseRSAKey(pemString string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemString))

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("An error occurred when parsing private key:%+v", err)
	}

	return key, nil
}
