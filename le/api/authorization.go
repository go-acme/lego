package api

import (
	"github.com/xenolf/lego/le"
)

type AuthorizationService service

func (c *AuthorizationService) Get(authzURL string) (le.Authorization, error) {
	var authz le.Authorization
	_, err := c.core.postAsGet(authzURL, &authz)
	if err != nil {
		return le.Authorization{}, err
	}
	return authz, nil
}

func (c *AuthorizationService) Disable(authzURL string) error {
	var disabledAuth le.Authorization
	_, err := c.core.post(authzURL, le.Authorization{Status: le.StatusDeactivated}, &disabledAuth)
	return err
}
