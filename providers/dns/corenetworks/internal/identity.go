package internal

import (
	"context"
	"net/http"
)

const authorizationHeader = "Authorization"

type token string

const tokenKey token = "token"

// CreateAuthenticationToken gets an authentication token.
// https://beta.api.core-networks.de/doc/#functon_auth_token
func (c Client) CreateAuthenticationToken(ctx context.Context) (*Token, error) {
	endpoint := c.baseURL.JoinPath("auth", "token")

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, Auth{Login: c.login, Password: c.password})
	if err != nil {
		return nil, err
	}

	var token Token
	err = c.do(req, &token)
	if err != nil {
		return nil, err
	}

	return &token, nil
}

func (c Client) CreateAuthenticatedContext(ctx context.Context) (context.Context, error) {
	tok, err := c.CreateAuthenticationToken(ctx)
	if err != nil {
		return nil, err
	}

	return context.WithValue(ctx, tokenKey, tok.Token), nil
}

func getToken(ctx context.Context) string {
	tok, ok := ctx.Value(tokenKey).(string)
	if !ok {
		return ""
	}

	return tok
}
