package internal

import (
	"context"
	"encoding/json"
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

// GetToken gets valid token information.
// https://doc.conoha.jp/reference/api-vps2/api-identity-vps2/identity-post_tokens-v2/?btn_id=reference-paas-dns-delete-a-record-v2--sidebar_reference-identity-post_tokens-v2
func (c *Identifier) GetToken(ctx context.Context, auth Auth) (*IdentityResponse, error) {
	endpoint := c.baseURL.JoinPath("v2.0", "tokens")

	req, err := newJSONRequest(ctx, http.MethodPost, endpoint, &IdentityRequest{Auth: auth})
	if err != nil {
		return nil, err
	}

	identity := &IdentityResponse{}

	err = c.do(req, identity)
	if err != nil {
		return nil, err
	}

	return identity, nil
}

func (c *Identifier) do(req *http.Request, result any) error {
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return errutils.NewUnexpectedResponseStatusCodeError(req, resp)
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
