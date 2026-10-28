package internal

import (
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

const BaseOAuthURLTemplate = "https://apiauth.dk.team.blue/realms/%s/protocol/openid-connect/token"

func CreateOAuthConfig(tokenURL, clientID, clientSecret string) *clientcredentials.Config {
	return &clientcredentials.Config{
		TokenURL:     tokenURL,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       []string{"dns"},
		AuthStyle:    oauth2.AuthStyleInParams,
	}
}

func OAuthStaticAccessToken(client *http.Client, accessToken string) *http.Client {
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}

	client.Transport = &oauth2.Transport{
		Source: oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken}),
		Base:   client.Transport,
	}

	return client
}
