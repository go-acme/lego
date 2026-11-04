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
	"golang.org/x/oauth2"
)

type Config struct {
	BaseURL    *url.URL
	HTTPClient *http.Client

	username string
	password string
	apiKey   string
}

func NewAuthConfig(username, password, apiKey string) (*Config, error) {
	if username == "" || password == "" || apiKey == "" {
		return nil, errors.New("credentials missing")
	}

	baseURL, _ := url.Parse(defaultBaseURL)

	return &Config{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
		username:   username,
		password:   password,
		apiKey:     apiKey,
	}, nil
}

func (c *Config) TokenSource(ctx context.Context) *TokenSource {
	return &TokenSource{
		ctx:  ctx,
		conf: c,
	}
}

//nolint:containedctx // Because of the interface oauth2.TokenSource.
type TokenSource struct {
	ctx  context.Context
	conf *Config
}

func (s *TokenSource) Token() (*oauth2.Token, error) {
	data, err := s.login(s.ctx)
	if err != nil {
		return nil, err
	}

	tok := &oauth2.Token{
		AccessToken:  data.AccessToken,
		TokenType:    data.TokenType,
		RefreshToken: data.RefreshToken,
		ExpiresIn:    data.ExpiresIn,
	}

	if data.ExpiresIn > 0 {
		tok.Expiry = time.Now().Add(time.Duration(data.ExpiresIn) * time.Second)
	}

	return tok, nil
}

// login gets a TokenInfo.
// https://api.comlaude.com/docs#/Auth/POST.api_login
func (s *TokenSource) login(ctx context.Context) (*TokenInfo, error) {
	endpoint := s.conf.BaseURL.JoinPath("api_login")

	alr := APILoginRequest{
		Username: s.conf.username,
		Password: s.conf.password,
		APIKey:   s.conf.apiKey,
	}

	req, err := newRequest(ctx, http.MethodPost, endpoint, alr)
	if err != nil {
		return nil, err
	}

	var info LoginResponse

	err = s.do(req, &info)
	if err != nil {
		return nil, err
	}

	return info.Data, nil
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
