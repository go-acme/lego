package internal

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

type token string

const tokenKey token = "token"

// login Logs in as API user.
// Authenticates and receives a token to be used in for subsequent requests.
// https://docs.bluecatnetworks.com/r/Address-Manager-Legacy-v1-API-Guide/GET/v1/login/9.5.0
func (c *Client) login(ctx context.Context) (string, error) {
	endpoint := c.createEndpoint("login")

	q := endpoint.Query()
	q.Set("username", c.username)
	q.Set("password", c.password)
	endpoint.RawQuery = q.Encode()

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", errutils.NewUnexpectedResponseStatusCodeError(req, resp)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	authResp := string(raw)
	if strings.Contains(authResp, "Authentication Error") {
		return "", fmt.Errorf("request failed: %s", strings.Trim(authResp, `"`))
	}

	// Upon success, API responds with "Session Token-> BAMAuthToken: dQfuRMTUxNjc3MjcyNDg1ODppcGFybXM= <- for User : username"
	tok := c.tokenExp.FindString(authResp)

	return tok, nil
}

// Logout Logs out of the current API session.
// https://docs.bluecatnetworks.com/r/Address-Manager-Legacy-v1-API-Guide/GET/v1/logout/9.5.0
func (c *Client) Logout(ctx context.Context) error {
	if getToken(ctx) == "" {
		// nothing to do
		return nil
	}

	endpoint := c.createEndpoint("logout")

	req, err := newJSONRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}

	resp, err := c.doAuthenticated(ctx, req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return errutils.NewUnexpectedResponseStatusCodeError(req, resp)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	authResp := string(raw)
	if !strings.Contains(authResp, "successfully") {
		return fmt.Errorf("request failed to delete session: %s", strings.Trim(authResp, `"`))
	}

	return nil
}

func (c *Client) CreateAuthenticatedContext(ctx context.Context) (context.Context, error) {
	tok, err := c.login(ctx)
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
