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

type token string

const tokenKey token = "token"

type Identifier struct {
	BaseURL    *url.URL
	HTTPClient *http.Client
}

func NewIdentifier() *Identifier {
	baseURL, _ := url.Parse(defaultBaseURL)

	return &Identifier{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Identifier) APILogin(ctx context.Context, username, password, apiKey string) (*TokenInfo, error) {
	if username == "" || password == "" || apiKey == "" {
		return nil, errors.New("credentials missing")
	}

	endpoint := c.BaseURL.JoinPath("api_login")

	alr := APILoginRequest{
		Username: username,
		Password: password,
		APIKey:   apiKey,
	}

	req, err := newRequest(ctx, http.MethodPost, endpoint, alr)
	if err != nil {
		return nil, err
	}

	info := new(LoginResponse)

	err = c.do(req, info)
	if err != nil {
		return nil, err
	}

	return &info.Data[0], nil
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
