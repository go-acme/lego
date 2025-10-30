// internal/identity.go

package internal

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

const identityBaseURL = "https://identity.%s.conoha.io"

type Identifier struct {
	baseURL    *url.URL
	HTTPClient *http.Client
}

// NewIdentifier creates a new Identifier.
func NewIdentifier(region string) (*Identifier, error) {
	baseURL, err := url.Parse(fmt.Sprintf(identityBaseURL, region))
	if err != nil {
		return nil, err
	}

	return &Identifier{
		baseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}, nil
}

// GetToken returns the x-subject-token from Identity API.
// https://doc.conoha.jp/reference/api-vps3/api-identity-vps3/identity-post_tokens-v3/?btn_id=reference-api-guideline-v3--sidebar_reference-identity-post_tokens-v3
func (c *Identifier) GetToken(ctx context.Context, auth Auth) (string, error) {
	endpoint := c.baseURL.JoinPath("v3", "auth", "tokens")

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, &IdentityRequest{Auth: auth})
	if err != nil {
		return "", err
	}

	return c.do(req)
}

// do sends the request and returns the token from x-subject-token header.
func (c *Identifier) do(req *http.Request) (string, error) {
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated {
		return "", errutils.NewUnexpectedResponseStatusCodeError(req, resp)
	}

	token := resp.Header.Get("x-subject-token")
	if token == "" {
		return "", errors.New("x-subject-token header is missing in response")
	}

	_, _ = io.Copy(io.Discard, resp.Body)

	return token, nil
}
