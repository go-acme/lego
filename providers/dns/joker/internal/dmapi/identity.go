package dmapi

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"
)

type token string

const sessionIDKey token = "session-id"

// Token session ID.
// > Every request (except "login") requires the presence of the Auth-Sid variable ("Session ID"),
// > which is returned by the "login" request (login). An active session will expire after some inactivity period (default: 1 hour).
// https://joker.com/faq/content/22/12/en/commonalities-for-all-requests.html
type Token struct {
	SessionID string
	ExpireAt  time.Time
}

// login performs a log in to Joker's DMAPI.
func (c *Client) login(ctx context.Context) (*Response, error) {
	var values url.Values

	switch {
	case c.username != "" && c.password != "":
		values = url.Values{
			"username": {c.username},
			"password": {c.password},
		}
	case c.apiKey != "":
		values = url.Values{"api-key": {c.apiKey}}
	default:
		return nil, errors.New("no username and password or api-key")
	}

	response, err := c.postRequest(ctx, "login", values)
	if err != nil {
		return response, err
	}

	if response == nil {
		return nil, errors.New("login returned nil response")
	}

	if response.AuthSid == "" {
		return response, errors.New("login did not return valid Auth-Sid")
	}

	return response, nil
}

// Logout closes authenticated session with Joker's DMAPI.
func (c *Client) Logout(ctx context.Context) (*Response, error) {
	if c.token == nil {
		return nil, errors.New("already logged out")
	}

	response, err := c.postRequest(ctx, "logout", url.Values{})

	c.muToken.Lock()
	c.token = nil
	c.muToken.Unlock()

	if err != nil {
		return response, err
	}

	return response, nil
}

func (c *Client) CreateAuthenticatedContext(ctx context.Context) (context.Context, error) {
	c.muToken.Lock()
	defer c.muToken.Unlock()

	if c.token != nil && time.Now().UTC().Before(c.token.ExpireAt) {
		return context.WithValue(ctx, sessionIDKey, c.token.SessionID), nil
	}

	response, err := c.login(ctx)
	if err != nil {
		return nil, formatResponseError(response, err)
	}

	c.token = &Token{
		SessionID: response.AuthSid,
		ExpireAt:  time.Now().UTC().Add(1 * time.Hour),
	}

	return context.WithValue(ctx, sessionIDKey, response.AuthSid), nil
}

func getSessionID(ctx context.Context) string {
	tok, ok := ctx.Value(sessionIDKey).(string)
	if !ok {
		return ""
	}

	return tok
}

// formatResponseError formats error with optional details from DMAPI response.
func formatResponseError(response *Response, err error) error {
	if response != nil {
		return fmt.Errorf("joker: DMAPI error: %w Response: %v", err, response.Headers)
	}

	return fmt.Errorf("joker: DMAPI error: %w", err)
}
