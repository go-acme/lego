package internal

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
)

// OauthConfiguration credentials.
type OauthConfiguration struct {
	OAuth2ClientID string
	OAuth2SecretID string
	Username       string
	Password       string
}

func (config *OauthConfiguration) Validate() error {
	msg := " is missing in credentials information"

	if config.Username == "" {
		return errors.New("username" + msg)
	}

	if config.Password == "" {
		return errors.New("password" + msg)
	}

	if config.OAuth2ClientID == "" {
		return errors.New("serviceID" + msg)
	}

	if config.OAuth2SecretID == "" {
		return errors.New("secret" + msg)
	}

	return nil
}

func NewOauthClient(ctx context.Context, config *OauthConfiguration) (*http.Client, error) {
	err := config.Validate()
	if err != nil {
		return nil, err
	}

	oauth2Config := oauth2.Config{
		ClientID:     config.OAuth2ClientID,
		ClientSecret: config.OAuth2SecretID,
		Endpoint: oauth2.Endpoint{
			TokenURL:  tokenURL,
			AuthStyle: oauth2.AuthStyleInParams,
		},
		Scopes: []string{".+:/dns-master/.+"},
	}

	oauth2Token, err := oauth2Config.PasswordCredentialsToken(ctx, config.Username, config.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to create oauth2 token: %w", err)
	}

	return oauth2Config.Client(ctx, oauth2Token), nil
}
