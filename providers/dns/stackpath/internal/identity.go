package internal

import (
	"context"
	"net/http"

	"golang.org/x/oauth2/clientcredentials"
)

const defaultAuthURL = "https://gateway.stackpath.com/identity/v1/oauth2/token"

func createOAuthClient(ctx context.Context, clientID, clientSecret string) *http.Client {
	config := &clientcredentials.Config{
		TokenURL:     defaultAuthURL,
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}

	return config.Client(ctx)
}
