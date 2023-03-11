package internal

import (
	"context"
	"net/http"
)

const loginURL = "https://admin.vshosting.cloud/api/public/auth/login"

type token string

const accessTokenKey token = "accessToken"

func (c *Client) login(ctx context.Context) (*AuthResponse, error) {
	authorization := Authorization{Email: c.email, Password: c.password}

	req, err := newJSONRequest(ctx, http.MethodPost, c.loginURL, authorization)
	if err != nil {
		return nil, err
	}

	var result AuthResponse
	err = c.do(req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) CreateAuthenticatedContext(ctx context.Context) (context.Context, error) {
	tok, err := c.login(ctx)
	if err != nil {
		return nil, err
	}

	return context.WithValue(ctx, accessTokenKey, tok.Auth.AccessToken), nil
}

func getAccessToken(ctx context.Context) string {
	tok, ok := ctx.Value(accessTokenKey).(string)
	if !ok {
		return ""
	}

	return tok
}
