package internal

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
)

// DefaultIdentityURL represents the Identity API endpoint to call.
const DefaultIdentityURL = "https://identity.api.rackspacecloud.com/v2.0/tokens"

type Identifier struct {
	baseURL    string
	httpClient *http.Client
}

// NewIdentifier creates a new Identifier.
func NewIdentifier(httpClient *http.Client, baseURL string) *Identifier {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 5 * time.Second}
	}

	if baseURL == "" {
		baseURL = DefaultIdentityURL
	}

	return &Identifier{baseURL: baseURL, httpClient: httpClient}
}

// Login sends an authentication request.
// https://docs.rackspace.com/docs/cloud-dns/v1/getting-started/authenticate
func (a *Identifier) Login(ctx context.Context, apiUser, apiKey string) (*Identity, error) {
	authData := AuthData{
		Auth: Auth{
			APIKeyCredentials: APIKeyCredentials{
				Username: apiUser,
				APIKey:   apiKey,
			},
		},
	}

	req, err := newJSONRequest(ctx, http.MethodPost, a.baseURL, authData)
	if err != nil {
		return nil, err
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, errutils.NewUnexpectedResponseStatusCodeError(req, resp)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	var identity Identity
	err = json.Unmarshal(raw, &identity)
	if err != nil {
		return nil, errutils.NewUnmarshalError(req, resp.StatusCode, raw, err)
	}

	return &identity, nil
}
