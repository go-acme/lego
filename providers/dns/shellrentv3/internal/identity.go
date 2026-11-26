package internal

import (
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

const defaultOAuthURL = "https://api.shellrent.com/oauth/token"

// CreateOAuthConfig create oauth2 configuration.
// https://guide.shellrent.com/api/oauth-client
func CreateOAuthConfig(tokenURL, clientID, clientSecret string) *clientcredentials.Config {
	if tokenURL == "" {
		tokenURL = defaultOAuthURL
	}

	return &clientcredentials.Config{
		TokenURL:     tokenURL,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		AuthStyle:    oauth2.AuthStyleInHeader,
	}
}
