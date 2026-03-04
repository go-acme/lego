package secure

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"encoding/base64"
	"fmt"

	"github.com/go-jose/go-jose/v4"
)

type nonceSourceCreator interface {
	NewNonceSource(ctx context.Context) jose.NonceSource
}

// JWS Represents a JWS.
type JWS struct {
	privKey crypto.PrivateKey
	kid     string // Key identifier
	nonces  nonceSourceCreator
}

// NewJWS Create a new JWS.
func NewJWS(privateKey crypto.PrivateKey, kid string, nonceManager nonceSourceCreator) *JWS {
	return &JWS{
		privKey: privateKey,
		nonces:  nonceManager,
		kid:     kid,
	}
}

// SignContent Signs a content with the JWS.
func (j *JWS) SignContent(ctx context.Context, url string, content []byte) (*jose.JSONWebSignature, error) {
	signKey := jose.SigningKey{
		Algorithm: signatureAlgorithm(j.privKey),
		Key:       jose.JSONWebKey{Key: j.privKey, KeyID: j.kid},
	}

	options := &jose.SignerOptions{
		NonceSource: j.nonces.NewNonceSource(ctx),
		ExtraHeaders: map[jose.HeaderKey]any{
			"url": url,
		},
		EmbedJWK: j.kid == "",
	}

	return sign(content, signKey, options)
}

// SignEAB Signs an external account binding with the JWS.
func (j *JWS) SignEAB(url, kid string, hmac []byte) (*jose.JSONWebSignature, error) {
	jwk := jose.JSONWebKey{Key: j.privKey}

	jwkJSON, err := jwk.Public().MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("acme: error encoding eab jwk key: %w", err)
	}

	signKey := jose.SigningKey{Algorithm: jose.HS256, Key: hmac}

	options := &jose.SignerOptions{
		EmbedJWK: false,
		ExtraHeaders: map[jose.HeaderKey]any{
			"kid": kid,
			"url": url,
		},
	}

	signed, err := sign(jwkJSON, signKey, options)
	if err != nil {
		return nil, fmt.Errorf("EAB: %w", err)
	}

	return signed, nil
}

// GetKeyAuthorization Gets the key authorization for a token.
func (j *JWS) GetKeyAuthorization(token string) (string, error) {
	var publicKey crypto.PublicKey

	signer, ok := j.privKey.(crypto.Signer)
	if ok {
		publicKey = signer.Public()
	}

	// Generate the Key Authorization for the challenge
	jwk := &jose.JSONWebKey{Key: publicKey}

	thumbBytes, err := jwk.Thumbprint(crypto.SHA256)
	if err != nil {
		return "", err
	}

	// unpad the base64URL
	keyThumb := base64.RawURLEncoding.EncodeToString(thumbBytes)

	return token + "." + keyThumb, nil
}

func sign(content []byte, signKey jose.SigningKey, options *jose.SignerOptions) (*jose.JSONWebSignature, error) {
	signer, err := jose.NewSigner(signKey, options)
	if err != nil {
		return nil, fmt.Errorf("new jose signer: %w", err)
	}

	signed, err := signer.Sign(content)
	if err != nil {
		return nil, fmt.Errorf("sign content: %w", err)
	}

	return signed, nil
}

func signatureAlgorithm(privKey crypto.PrivateKey) jose.SignatureAlgorithm {
	var alg jose.SignatureAlgorithm

	switch k := privKey.(type) {
	case *rsa.PrivateKey:
		alg = jose.RS256
	case *ecdsa.PrivateKey:
		if k.Curve == elliptic.P256() {
			alg = jose.ES256
		} else if k.Curve == elliptic.P384() {
			alg = jose.ES384
		}
	}

	return alg
}
