package internal

import (
	"bytes"
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// Key is resource managed by IAM Key Service.
// Can be issued for User or Service Account, but key authorization is supported only for Service Accounts.
// Issued key contains private part that is not saved on server side, and should be saved by client.
// https://github.com/yandex-cloud/go-sdk/blob/a66fac694b0c779debcae1c137b83d7974b420e6/iamkey/key.pb.go
type Key struct {
	// ID of the Key resource.
	Id string `json:"id,omitempty"`
	// ID of the user account that the Key resource belongs to.
	UserAccountId string `json:"user_account_id,omitempty"`
	// ID of the service account that the Key resource belongs to.
	ServiceAccountId string `json:"service_account_id,omitempty"`
	// Creation timestamp in [RFC3339](https://www.ietf.org/rfc/rfc3339.txt) text format.
	CreatedAt *time.Time `json:"created_at,omitempty"`
	// Description of the Key resource. 0-256 characters long.
	Description string `json:"description,omitempty"`
	// An algorithm used to generate a key pair of the Key resource.
	KeyAlgorithm string `json:"key_algorithm,omitempty"`
	// A public key of the Key resource.
	PublicKey string `json:"public_key,omitempty"`
	// A public key of the Key resource.
	PrivateKey string `json:"private_key,omitempty"`
}

// tokenRequest contains the token request structure
type tokenRequest struct {
	JWT string `json:"jwt"`
}

// tokenResponse contains the token response structure
type tokenResponse struct {
	IAMToken string `json:"iamToken"`
}

// JWT generation.
func signedToken(key Key) (string, error) {
	claims := jwt.RegisteredClaims{
		Issuer:    key.ServiceAccountId,
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(1 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		NotBefore: jwt.NewNumericDate(time.Now().UTC()),
		Audience:  []string{"https://iam.api.cloud.yandex.net/iam/v1/tokens"},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodPS256, claims)
	token.Header["kid"] = key.Id

	privateKey, err := loadPrivateKey(key)
	if err != nil {
		return "", fmt.Errorf("failed to load private key: %w", err)
	}

	signed, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signed, nil
}

func loadPrivateKey(keyData Key) (*rsa.PrivateKey, error) {
	rsaPrivateKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(keyData.PrivateKey))
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}
	return rsaPrivateKey, nil
}

// getIAMToken obtains an IAM token from Yandex Cloud.
func (c *Client) getIAMToken(ctx context.Context) (string, error) {
	jwtToken, err := signedToken(c.key)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	requestData := tokenRequest{
		JWT: jwtToken,
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal token request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, IAMAPIEndpoint+"/tokens", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	var token tokenResponse
	err = c.doJSONRequest(req, &token)
	if err != nil {
		return "", fmt.Errorf("failed to request IAM token: %w", err)
	}

	return token.IAMToken, nil
}
