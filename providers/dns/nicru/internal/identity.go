package internal

import (
	"context"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
)

const tokenURL = "https://api.nic.ru/oauth/token"

func NewOauthClient(ctx context.Context, clientID, clientSecret, username, password string) (*http.Client, error) {
	oauth2Config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint: oauth2.Endpoint{
			TokenURL:  tokenURL,
			AuthStyle: oauth2.AuthStyleInParams,
		},
		Scopes: []string{".+:/dns-master/.+"},
	}

	// Note: under the hood, this function is doing a request to the token endpoint.
	oauth2Token, err := oauth2Config.PasswordCredentialsToken(ctx, username, password)
	if err != nil {
		return nil, fmt.Errorf("failed to create oauth2 token: %w", err)
	}

	return oauth2Config.Client(ctx, oauth2Token), nil
}
