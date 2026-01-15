package internal

import (
	"errors"
	"net/url"
)

type Authentication struct {
	Username string `url:"user,omitempty"`
	Password string `url:"pwd,omitempty"`
	Key      string `url:"key,omitempty"`
}

func (a *Authentication) validate() error {
	if a.Username == "" && a.Password == "" && a.Key == "" {
		return errors.New("missing credentials")
	}

	if a.Username != "" && a.Password != "" && a.Key != "" {
		return errors.New("only one of username, password or key can be set")
	}

	if (a.Username != "" && a.Password == "") || a.Username == "" && a.Password != "" {
		return errors.New("username and password must be set together")
	}

	return nil
}

func (a *Authentication) set(query url.Values) {
	if a.Key != "" {
		query.Set("key", a.Key)

		return
	}

	query.Set("user", a.Username)
	query.Set("pwd", a.Password)
}
