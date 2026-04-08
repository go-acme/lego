package internal

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
	"github.com/go-acme/lego/v4/providers/dns/internal/useragent"
)

const defaultAPIVersion = "2.14.2-0"

type sessionIDKeyType string

const sessionIDKey sessionIDKeyType = "sessionID"

type Identifier struct {
	email    string
	password string
	orderID  string

	BaseURL    string
	HTTPClient *http.Client
}

func NewIdentifier(email, password, orderID string) (*Identifier, error) {
	if email == "" || password == "" || orderID == "" {
		return nil, errors.New("credentials missing")
	}

	return &Identifier{
		email:      email,
		password:   password,
		orderID:    orderID,
		BaseURL:    defaultBaseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// Login creates a new session and performs a customer login.
func (c *Identifier) Login(ctx context.Context) (string, error) {
	sessionID, err := c.getSessionID(ctx)
	if err != nil {
		return "", err
	}

	return c.login(WithContext(ctx, sessionID), LoginRequest{
		Email:      c.email,
		Password:   c.password,
		OrderID:    c.orderID,
		APIVersion: defaultAPIVersion,
	})
}

// login performs a customer login.
// https://support.euserv.com/api-doc/#api-Customer-login
func (c *Identifier) login(ctx context.Context, request LoginRequest) (string, error) {
	endpoint, err := url.Parse(c.BaseURL)
	if err != nil {
		return "", err
	}

	query := endpoint.Query()
	query.Set("subaction", "login")
	endpoint.RawQuery = query.Encode()

	req, err := newHTTPRequest(ctx, endpoint, request)
	if err != nil {
		return "", err
	}

	var response APIResponse

	err = c.do(req, &response)
	if err != nil {
		return "", err
	}

	data, err := extractResponse[Session](response)
	if err != nil {
		return "", err
	}

	return data.ID.Value, nil
}

// Logout performs a customer logout and end the given session.
// https://support.euserv.com/api-doc/#api-Customer-logout
func (c *Identifier) Logout(ctx context.Context) error {
	endpoint, err := url.Parse(c.BaseURL)
	if err != nil {
		return err
	}

	query := endpoint.Query()
	query.Set("action", "logout")
	endpoint.RawQuery = query.Encode()

	req, err := newHTTPRequest(ctx, endpoint, nil)
	if err != nil {
		return err
	}

	var response APIResponse

	err = c.do(req, &response)
	if err != nil {
		return err
	}

	_, err = extractResponse[any](response)
	if err != nil {
		return err
	}

	return nil
}

// getSessionID gets a new session id.
// https://support.euserv.com/api-doc/#api-Session-Get_a_new_session_id
func (c *Identifier) getSessionID(ctx context.Context) (string, error) {
	endpoint, err := url.Parse(c.BaseURL)
	if err != nil {
		return "", err
	}

	req, err := newHTTPRequest(ctx, endpoint, nil)
	if err != nil {
		return "", err
	}

	var response APIResponse

	err = c.do(req, &response)
	if err != nil {
		return "", err
	}

	data, err := extractResponse[Session](response)
	if err != nil {
		return "", err
	}

	return data.ID.Value, nil
}

func (c *Identifier) do(req *http.Request, result any) error {
	useragent.SetHeader(req.Header)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode/100 != 2 {
		raw, _ := io.ReadAll(resp.Body)

		return errutils.NewUnexpectedStatusCodeError(req, resp.StatusCode, raw)
	}

	if result == nil {
		return nil
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	err = json.Unmarshal(raw, result)
	if err != nil {
		return errutils.NewUnmarshalError(req, resp.StatusCode, raw, err)
	}

	return nil
}

func WithContext(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, sessionIDKey, id)
}

func getSessionID(ctx context.Context) string {
	id, ok := ctx.Value(sessionIDKey).(string)
	if !ok {
		return ""
	}

	return id
}
