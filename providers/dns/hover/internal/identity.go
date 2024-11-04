package internal

import (
	"context"
	"errors"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
)

const cookieName = "hoverauth"

// Login returns the authentication key for the username and password,
// performing a login if the key is not already known from a previous login.
func (c *Client) Login(ctx context.Context) error {
	values := url.Values{
		"username": {c.username},
		"password": {c.password},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL.JoinPath("login").String(), strings.NewReader(values.Encode()))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	if c.getCookie(req.URL) != "" {
		return nil
	}

	err = c.do(req, nil)
	if err != nil {
		return err
	}

	if c.getCookie(req.URL) != "" {
		return nil
	}

	return errors.New("login failed")
}

func (c *Client) Logout() {
	c.resetCookieJar()
}

func (c *Client) resetCookieJar() {
	c.HTTPClient.Jar, _ = cookiejar.New(nil)
}

func (c *Client) getCookie(endpoint *url.URL) string {
	if c.HTTPClient.Jar == nil {
		c.resetCookieJar()
	}

	if len(c.HTTPClient.Jar.Cookies(endpoint)) < 1 {
		return ""
	}

	for _, v := range c.HTTPClient.Jar.Cookies(endpoint) {
		if v.Name == cookieName {
			return v.Value
		}
	}

	return ""
}
