package internal

import (
	"net/http"
	"net/url"
)

type Option func(c *Client) error

func WithAuthKey(authEmail, authKey string) Option {
	return func(c *Client) error {
		c.authEmail = authEmail
		c.authKey = authKey

		return nil
	}
}

func WithAuthToken(authToken string) Option {
	return func(c *Client) error {
		c.authToken = authToken

		return nil
	}
}

func WithBaseURL(baseURL string) Option {
	return func(c *Client) error {
		if baseURL == "" {
			return nil
		}

		bu, err := url.Parse(baseURL)
		if err != nil {
			return err
		}

		c.baseURL = bu

		return nil
	}
}

func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) error {
		if client != nil {
			c.HTTPClient = client
		}

		return nil
	}
}
