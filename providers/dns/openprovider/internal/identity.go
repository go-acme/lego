package internal

import (
	"context"
	"net/http"
)

const authorizationHeader = "Authorization"

type token string

const tokenKey token = "token"

// Login gets an authentication token.
// https://docs.openprovider.com/doc/all#operation/Login
func (c *Client) Login(ctx context.Context) (string, error) {
	endpoint := c.BaseURL.JoinPath("auth", "login")

	payload := Login{
		Username: c.username,
		Password: c.password,
		IP:       "0.0.0.0",
	}

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, payload)
	if err != nil {
		return "", err
	}

	results := &APIResponse[Token]{}

	err = c.do(req, results)
	if err != nil {
		return "", err
	}

	return results.Data.Token, nil
}

func (c *Client) CreateAuthenticatedContext(ctx context.Context) (context.Context, error) {
	tok, err := c.Login(ctx)
	if err != nil {
		return nil, err
	}

	return context.WithValue(ctx, tokenKey, tok), nil
}

func getToken(ctx context.Context) string {
	tok, ok := ctx.Value(tokenKey).(string)
	if !ok {
		return ""
	}

	return tok
}
