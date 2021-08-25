// Package svc Client for the SVC API.
// https://joker.com/faq/content/6/496/en/let_s-encrypt-support.html
package svc

import (
	"fmt"
	"io"
	"net/http"
	"strings"

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
	HTTPClient *http.Client
	BaseURL    string

	username string
	password string
}

func NewClient(username, password string) *Client {
	return &Client{
		HTTPClient: http.DefaultClient,
		BaseURL:    defaultBaseURL,
		username:   username,
		password:   password,
	}
}

func (c *Client) Send(zone, label, value string) error {
	req := request{
		Username: c.username,
		Password: c.password,
		Zone:     zone,
		Label:    label,
		Type:     "TXT",
		Value:    value,
	}

	v, err := querystring.Values(req)
	if err != nil {
		return err
	}

	resp, err := c.HTTPClient.PostForm(c.BaseURL, v)
	if err != nil {
		return err
	}

	all, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusOK && strings.HasPrefix(string(all), "OK") {
		return nil
	}

	return fmt.Errorf("error: %d: %s", resp.StatusCode, string(all))
}
