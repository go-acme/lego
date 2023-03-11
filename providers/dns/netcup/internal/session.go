package internal

import (
	"context"
	"fmt"
)

type sessionKey string

const sessionIDKey sessionKey = "sessionID"

// login performs the login as specified by the netcup WSDL
// returns sessionID needed to perform remaining actions.
// https://ccp.netcup.net/run/webservice/servers/endpoint.php
func (c *Client) login(ctx context.Context) (string, error) {
	payload := &Request{
		Action: "login",
		Param: &LoginRequest{
			CustomerNumber:  c.customerNumber,
			APIKey:          c.apiKey,
			APIPassword:     c.apiPassword,
			ClientRequestID: "",
		},
	}

	var responseData LoginResponse
	err := c.doRequest(ctx, payload, &responseData)
	if err != nil {
		return "", fmt.Errorf("loging error: %w", err)
	}

	return responseData.APISessionID, nil
}

// Logout performs the logout with the supplied sessionID as specified by the netcup WSDL.
// https://ccp.netcup.net/run/webservice/servers/endpoint.php
func (c *Client) Logout(ctx context.Context) error {
	payload := &Request{
		Action: "logout",
		Param: &LogoutRequest{
			CustomerNumber:  c.customerNumber,
			APIKey:          c.apiKey,
			APISessionID:    getSessionID(ctx),
			ClientRequestID: "",
		},
	}

	err := c.doRequest(ctx, payload, nil)
	if err != nil {
		return fmt.Errorf("logout error: %w", err)
	}

	return nil
}

func (c *Client) CreateSessionContext(ctx context.Context) (context.Context, error) {
	sessID, err := c.login(ctx)
	if err != nil {
		return nil, err
	}

	return context.WithValue(ctx, sessionIDKey, sessID), nil
}

func getSessionID(ctx context.Context) string {
	sessID, ok := ctx.Value(sessionIDKey).(string)
	if !ok {
		return ""
	}

	return sessID
}
