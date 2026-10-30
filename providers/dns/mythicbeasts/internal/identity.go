package internal

import (
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

const DefaultOAuthURL = "https://auth.mythic-beasts.com/login"

// CreateOAuthConfig create oauth2 configuration.
// https://www.mythic-beasts.com/support/api/auth#sec-obtaining-a-token
func CreateOAuthConfig(tokenURL, clientID, clientSecret string) *clientcredentials.Config {
	if tokenURL == "" {
		tokenURL = DefaultOAuthURL
	}

	return &clientcredentials.Config{
		TokenURL:     tokenURL,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		AuthStyle:    oauth2.AuthStyleInHeader,
	}
}
