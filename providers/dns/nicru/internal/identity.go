package internal

import (
	"context"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
)

const authBaseURL = "https://api.nic.ru/oauth/token"

func NewOAuthClient(ctx context.Context, cfg *oauth2.Config, username, password string) (*http.Client, error) {
	// Note: under the hood, this function is doing a request to the token endpoint.
	token, err := cfg.PasswordCredentialsToken(ctx, username, password)
	if err != nil {
		return nil, fmt.Errorf("failed to create oauth2 token: %w", err)
	}

	return cfg.Client(ctx, token), nil
}

func CreateOAuthConfig(tokenURL, clientID, clientSecret string) *oauth2.Config {
	if tokenURL == "" {
		tokenURL = authBaseURL
	}

	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint: oauth2.Endpoint{
			TokenURL:  tokenURL,
			AuthStyle: oauth2.AuthStyleInParams,
		},
		Scopes: []string{".+:/dns-master/.+"},
	}
}
