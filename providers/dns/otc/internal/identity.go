package internal

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

// DefaultIdentityEndpoint the default API identity endpoint.
const DefaultIdentityEndpoint = "https://iam.eu-de.otc.t-systems.com:443/v3/auth/tokens"

// Login Starts a new OTC API Session. Authenticates using userName, password
// and receives a token to be used in for subsequent requests.
func (c *Client) Login(ctx context.Context) error {
	payload := LoginRequest{
		Auth: Auth{
			Identity: Identity{
				Methods: []string{"password"},
				Password: Password{
					User: User{
						Name:     c.username,
						Password: c.password,
						Domain: Domain{
							Name: c.domainName,
						},
					},
				},
			},
			Scope: Scope{
				Project: Project{
					Name: c.projectName,
				},
			},
		},
	}

	tokenResp, token, err := c.obtainUserToken(ctx, payload)
	if err != nil {
		return err
	}

	c.muToken.Lock()
	defer c.muToken.Unlock()

	c.token = token

	if c.token == "" {
		return errors.New("unable to get auth token")
	}

	baseURL, err := getBaseURL(tokenResp)
	if err != nil {
		return err
	}

	c.muBaseURL.Lock()
	c.baseURL = baseURL
	c.muBaseURL.Unlock()

	return nil
}

// https://docs.otc.t-systems.com/identity-access-management/api-ref/apis/token_management/obtaining_a_user_token.html
func (c *Client) obtainUserToken(ctx context.Context, payload LoginRequest) (*TokenResponse, string, error) {
	req, err := newJSONRequest(ctx, http.MethodPost, c.IdentityEndpoint, payload)
	if err != nil {
		return nil, "", err
	}

	client := &http.Client{Timeout: c.HTTPClient.Timeout}

	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode/100 != 2 {
		return nil, "", errutils.NewUnexpectedResponseStatusCodeError(req, resp)
	}

	token := resp.Header.Get("X-Subject-Token")

	if token == "" {
		return nil, "", errors.New("unable to get auth token")
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	var newToken TokenResponse

	err = json.Unmarshal(raw, &newToken)
	if err != nil {
		return nil, "", errutils.NewUnmarshalError(req, resp.StatusCode, raw, err)
	}

	return &newToken, token, nil
}

func getBaseURL(tokenResp *TokenResponse) (*url.URL, error) {
	var endpoints []Endpoint

	for _, v := range tokenResp.Token.Catalog {
		if v.Type == "dns" {
			endpoints = append(endpoints, v.Endpoints...)
		}
	}

	if len(endpoints) == 0 {
		return nil, errors.New("unable to get dns endpoint")
	}

	baseURL, err := url.JoinPath(endpoints[0].URL, "v2")
	if err != nil {
		return nil, err
	}

	return url.Parse(baseURL)
}
