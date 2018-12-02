package api

import (
	"errors"

	"github.com/xenolf/lego/le"
)

type AuthorizationService service

// Get Gets an authorization.
func (c *AuthorizationService) Get(authzURL string) (le.Authorization, error) {
	if len(authzURL) == 0 {
		return le.Authorization{}, errors.New("authorization[get]: empty URL")
	}

	var authz le.Authorization
	_, err := c.core.postAsGet(authzURL, &authz)
	if err != nil {
		return le.Authorization{}, err
	}
	return authz, nil
}

// Deactivate Deactivates an authorization.
func (c *AuthorizationService) Deactivate(authzURL string) error {
	if len(authzURL) == 0 {
		return errors.New("authorization[deactivate]: empty URL")
	}

	var disabledAuth le.Authorization
	_, err := c.core.post(authzURL, le.Authorization{Status: le.StatusDeactivated}, &disabledAuth)
	return err
}
