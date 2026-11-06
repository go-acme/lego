package internal

import (
	"net/url"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

const AuthBaseURL = "https://auth.iam.cloud.ru/auth/system/openid/token"

// CreateOAuthConfig creates oauth2 client configuration.
// https://cloud.ru/ru/docs/clouddns/ug/topics/api-ref_authentication.html
func CreateOAuthConfig(tokenURL, clientID, clientSecret string) *clientcredentials.Config {
	if tokenURL == "" {
		tokenURL = AuthBaseURL
	}

	return &clientcredentials.Config{
		TokenURL:     tokenURL,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		AuthStyle:    oauth2.AuthStyleInParams,
		EndpointParams: url.Values{
			// custom implementation
			"grant_type": {"access_key"},
		},
	}
}
