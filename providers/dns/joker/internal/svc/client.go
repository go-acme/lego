// Package svc Client for the SVC API.
// https://joker.com/faq/content/6/496/en/let_s-encrypt-support.html
package svc

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/providers/dns/internal/errutils"
	querystring "github.com/google/go-querystring/query"
)

const defaultBaseURL = "https://svc.joker.com/nic/replace"

type request struct {
	Username string `url:"username"`
	Password string `url:"password"`
	Zone     string `url:"zone"`
	Label    string `url:"label"`
	Type     string `url:"type"`
	Value    string `url:"value"`
}

type Client struct {
	username string
	password string

	BaseURL    string
	HTTPClient *http.Client
}

func NewClient(username, password string) *Client {
	return &Client{
		username:   username,
		password:   password,
		BaseURL:    defaultBaseURL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *Client) SendRequest(ctx context.Context, zone, label, value string) error {
	payload := request{
		Username: c.username,
		Password: c.password,
		Zone:     zone,
		Label:    label,
		Type:     "TXT",
		Value:    value,
	}

	v, err := querystring.Values(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL, strings.NewReader(v.Encode()))
	if err != nil {
		return fmt.Errorf("unable to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return errutils.NewReadResponseError(req, resp.StatusCode, err)
	}

	if resp.StatusCode == http.StatusOK && strings.HasPrefix(string(raw), "OK") {
		return nil
	}

	return fmt.Errorf("error: %d: %s", resp.StatusCode, string(raw))
}
