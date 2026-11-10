package internal

import (
	"context"

	"golang.org/x/oauth2"
)

const authBaseURL = "https://api.nic.ru/oauth/token"

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

//nolint:containedctx // Because of the interface oauth2.TokenSource.
type TokenSource struct {
	ctx context.Context

	cfg *oauth2.Config

	username string
	password string
}

func NewTokenSource(ctx context.Context, cfg *oauth2.Config, username, password string) *TokenSource {
	return &TokenSource{ctx: ctx, cfg: cfg, username: username, password: password}
}

func (s *TokenSource) Token() (*oauth2.Token, error) {
	return s.cfg.PasswordCredentialsToken(s.ctx, s.username, s.password)
}
