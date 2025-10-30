package internal

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

type token string

const tokenKey token = "token"

const authTokenHeader = "Auth-Token"

// login Starts a new Dyn API Session. Authenticates using customerName, username, password
// and receives a token to be used in for subsequent requests.
// https://help.dyn.com/session-log-in/
func (c *Client) login(ctx context.Context) (session, error) {
	endpoint := c.baseURL.JoinPath("Session")

	payload := &credentials{Customer: c.customerName, User: c.username, Pass: c.password}

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, payload)
	if err != nil {
		return session{}, err
	}

	dynRes, err := c.do(req)
	if err != nil {
		return session{}, err
	}

	var s session

	err = json.Unmarshal(dynRes.Data, &s)
	if err != nil {
		return session{}, errutils.NewUnmarshalError(req, http.StatusOK, dynRes.Data, err)
	}

	return s, nil
}

// Logout Destroys Dyn Session.
// https://help.dyn.com/session-log-out/
func (c *Client) Logout(ctx context.Context) error {
	endpoint := c.baseURL.JoinPath("Session")

	req, err := newJSONRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	tok := getToken(ctx)
	if tok != "" {
		req.Header.Set(authTokenHeader, tok)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return errutils.NewUnexpectedResponseStatusCodeError(req, resp)
	}

	return nil
}

func (c *Client) CreateAuthenticatedContext(ctx context.Context) (context.Context, error) {
	tok, err := c.login(ctx)
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
