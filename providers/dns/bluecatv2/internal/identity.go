package internal

import (
	"context"
	"fmt"
	"net/http"
)

type token string

const tokenKey token = "token"

const authorizationHeader = "Authorization"

// CreateSession creates a new session.
func (c *Client) CreateSession(ctx context.Context, info LoginInfo) (*Session, error) {
	endpoint := c.baseURL.JoinPath("api", "v2", "sessions")

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, info)
	if err != nil {
		return nil, err
	}

	result := new(Session)

	err = c.do(req, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// CreateAuthenticatedContext creates a new authenticated context.
func (c *Client) CreateAuthenticatedContext(ctx context.Context) (context.Context, error) {
	tok, err := c.CreateSession(ctx, LoginInfo{Username: c.username, Password: c.password})
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	return context.WithValue(ctx, tokenKey, tok.BasicAuthenticationCredentials), nil
}

func (c *Client) doAuthenticated(ctx context.Context, req *http.Request, result any) error {
	tok := getToken(ctx)
	if tok != "" {
		req.Header.Set(authorizationHeader, "Basic "+tok)
	}

	return c.do(req, result)
}

func getToken(ctx context.Context) string {
	tok, ok := ctx.Value(tokenKey).(string)
	if !ok {
		return ""
	}

	return tok
}
