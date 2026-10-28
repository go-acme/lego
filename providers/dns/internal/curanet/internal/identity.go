package internal

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

const defaultAuthURL = "https://apiauth.dk.team.blue/realms/%s/protocol/openid-connect/token"

func CreateOAuthClient(ctx context.Context, oauthSource, clientID, clientSecret string) *http.Client {
	config := &clientcredentials.Config{
		TokenURL:     fmt.Sprintf(defaultAuthURL, oauthSource),
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       []string{"dns"},
	}

	return config.Client(ctx)
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
