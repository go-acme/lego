package internal

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/go-acme/lego/v5/internal/errutils"
	"github.com/go-acme/lego/v5/internal/useragent"
)

type token string

const tokenKey token = "token"

type Identifier struct {
	username string
	token    string

	BaseURL    *url.URL
	HTTPClient *http.Client
}

func NewIdentifier(username, token string) (*Identifier, error) {
	if username == "" || token == "" {
		return nil, errors.New("credentials missing")
	}

	baseURL, _ := url.Parse(defaultBaseURL)

	return &Identifier{
		username:   username,
		token:      token,
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

func (c *Identifier) Login(ctx context.Context) (*Token, error) {
	endpoint := c.BaseURL.JoinPath("auth", "login")

	auth := Auth{Login: c.username, Token: c.token}

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, auth)
	if err != nil {
		return nil, err
	}

	var result APIResponse[*Token]

	err = c.do(req, &result)
	if err != nil {
		return nil, err
	}

	return result.Data, nil
}

func (c *Identifier) do(req *http.Request, result any) error {
	useragent.SetHeader(req.Header)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode/100 != 2 {
		return parseError(req, resp)
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

func WithContext(ctx context.Context, credential string) context.Context {
	return context.WithValue(ctx, tokenKey, credential)
}

func getToken(ctx context.Context) string {
	credential, ok := ctx.Value(tokenKey).(string)
	if !ok {
		return ""
	}

	return credential
}
