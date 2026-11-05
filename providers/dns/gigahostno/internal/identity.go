package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/go-acme/lego/v5/internal/errutils"
	"github.com/go-acme/lego/v5/internal/useragent"
	"github.com/pquerna/otp/totp"
	"golang.org/x/oauth2"
)

type Config struct {
	BaseURL    *url.URL
	HTTPClient *http.Client

	apiKey   string
	username string
	password string
	Secret   string
}

func NewAuthConfig(apiKey, username, password, secret string) (*Config, error) {
	if apiKey != "" {
		return &Config{apiKey: apiKey}, nil
	}

	if username == "" || password == "" {
		return nil, errors.New("credentials missing")
	}

	baseURL, _ := url.Parse(defaultBaseURL)

	return &Config{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
		username:   username,
		password:   password,
		Secret:     secret,
	}, nil
}

func (c *Config) TokenSource(ctx context.Context) oauth2.TokenSource {
	if c.apiKey != "" {
		return oauth2.StaticTokenSource(&oauth2.Token{AccessToken: c.apiKey})
	}

	return &TokenSource{ctx: ctx, conf: c}
}

//nolint:containedctx // Because of the interface oauth2.TokenSource.
type TokenSource struct {
	ctx  context.Context
	conf *Config
}

func (s *TokenSource) Token() (*oauth2.Token, error) {
	data, err := s.authenticate(s.ctx)
	if err != nil {
		return nil, err
	}

	tok := &oauth2.Token{
		AccessToken: data.Token,
		TokenType:   "Bearer",
		Expiry:      time.Unix(data.TokenExpire, 0),
	}

	tok.ExpiresIn = data.TokenExpire - time.Now().Unix()

	return tok, nil
}

func (s *TokenSource) authenticate(ctx context.Context) (*Token, error) {
	endpoint := s.conf.BaseURL.JoinPath("authenticate")

	auth := Auth{Username: s.conf.username, Password: s.conf.password}

	if s.conf.Secret != "" {
		tan, err := totp.GenerateCode(s.conf.Secret, time.Now())
		if err != nil {
			return nil, fmt.Errorf("generate TOTP: %w", err)
		}

		auth.Code, err = strconv.Atoi(tan)
		if err != nil {
			return nil, fmt.Errorf("parse TOTP: %w", err)
		}
	}

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, auth)
	if err != nil {
		return nil, err
	}

	var result APIResponse[*Token]

	err = s.do(req, &result)
	if err != nil {
		return nil, err
	}

	return result.Data, nil
}

func (s *TokenSource) do(req *http.Request, result any) error {
	useragent.SetHeader(req.Header)

	resp, err := s.conf.HTTPClient.Do(req)
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
