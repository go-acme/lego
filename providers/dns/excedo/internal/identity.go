package internal

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type ExpirableToken struct {
	Token   string
	Expires time.Time
}

func (t *ExpirableToken) IsExpired() bool {
	return time.Now().After(t.Expires)
}

func (c *Client) Login(ctx context.Context) (string, error) {
	endpoint := c.baseURL.JoinPath("/authenticate/login/")

	req, err := newFormRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	result := new(LoginResponse)

	err = c.do(req, result)
	if err != nil {
		return "", err
	}

	if result.Code != 1000 && result.Code != 1300 {
		return "", fmt.Errorf("%d: %s", result.Code, result.Description)
	}

	return result.Parameters.Token, nil
}

func (c *Client) authenticate(ctx context.Context) (string, error) {
	c.muToken.Lock()
	defer c.muToken.Unlock()

	if c.token == nil || c.token.IsExpired() {
		token, err := c.Login(ctx)
		if err != nil {
			return "", err
		}

		c.token = &ExpirableToken{
			Token:   token,
			Expires: time.Now().Add(2*time.Hour - time.Minute),
		}

		return token, nil
	}

	return c.token.Token, nil
}

func (c *Client) doAuthenticated(ctx context.Context, req *http.Request, result responseChecker) error {
	token, err := c.authenticate(ctx)
	if err != nil {
		return err
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	return c.do(req, result)
}
