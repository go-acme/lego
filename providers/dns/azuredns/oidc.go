package azuredns

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func checkOIDCConfig(config *Config) error {
	if config.TenantID == "" {
		return fmt.Errorf("azuredns: TenantID is missing")
	}

	if config.ClientID == "" {
		return fmt.Errorf("azuredns: ClientID is missing")
	}

	if config.OIDCToken == "" && config.OIDCTokenFilePath == "" && (config.OIDCRequestURL == "" || config.OIDCRequestToken == "") {
		return fmt.Errorf("azuredns: OIDCToken, OIDCTokenFilePath or OIDCRequestURL and OIDCRequestToken must be set")
	}

	return nil
}

func getOIDCAssertion(config *Config) func(ctx context.Context) (string, error) {
	return func(ctx context.Context) (string, error) {
		var token string
		if config.OIDCToken != "" {
			token = strings.TrimSpace(config.OIDCToken)
		}

		if config.OIDCTokenFilePath != "" {
			fileTokenRaw, err := os.ReadFile(config.OIDCTokenFilePath)
			if err != nil {
				return "", fmt.Errorf("azuredns: error retrieving token file with path %s: %w", config.OIDCTokenFilePath, err)
			}

			fileToken := strings.TrimSpace(string(fileTokenRaw))
			if config.OIDCToken != fileToken {
				return "", fmt.Errorf("azuredns: token file with path %s does not match token from environment variable", config.OIDCTokenFilePath)
			}

			token = fileToken
		}

		if token == "" && config.OIDCRequestURL != "" && config.OIDCRequestToken != "" {
			return getOIDCToken(config)
		}

		return token, nil
	}
}

func getOIDCToken(config *Config) (string, error) {
	req, err := http.NewRequest(http.MethodGet, config.OIDCRequestURL, http.NoBody)
	if err != nil {
		return "", fmt.Errorf("azuredns: failed to build OIDC request: %w", err)
	}

	query, err := url.ParseQuery(req.URL.RawQuery)
	if err != nil {
		return "", fmt.Errorf("azuredns: cannot parse OIDC request URL query")
	}

	if query.Get("audience") == "" {
		query.Set("audience", "api://AzureADTokenExchange")
		req.URL.RawQuery = query.Encode()
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.OIDCRequestToken))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := config.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("azuredns: cannot request OIDC token: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", fmt.Errorf("azuredns: cannot parse OIDC token response: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusNoContent {
		return "", fmt.Errorf("azuredns: OIDC token request received HTTP status %d with response: %s", resp.StatusCode, body)
	}

	var returnedToken struct {
		Count int    `json:"count"`
		Value string `json:"value"`
	}
	if err := json.Unmarshal(body, &returnedToken); err != nil {
		return "", fmt.Errorf("azuredns: cannot unmarshal OIDC token response: %w", err)
	}

	return returnedToken.Value, nil
}
