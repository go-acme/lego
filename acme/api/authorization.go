package api

import (
	"context"
	"errors"

	"github.com/go-acme/lego/v5/acme"
)

type AuthorizationService service

// Get Gets an authorization.
func (c *AuthorizationService) Get(ctx context.Context, authzURL string) (acme.Authorization, error) {
	if authzURL == "" {
		return acme.Authorization{}, errors.New("authorization[get]: empty URL")
	}

	var authz acme.Authorization

	_, err := c.core.postAsGet(ctx, authzURL, &authz)
	if err != nil {
		return acme.Authorization{}, err
	}

	return authz, nil
}

// Deactivate Deactivates an authorization.
func (c *AuthorizationService) Deactivate(ctx context.Context, authzURL string) error {
	if authzURL == "" {
		return errors.New("authorization[deactivate]: empty URL")
	}

	var disabledAuth acme.Authorization

	_, err := c.core.post(ctx, authzURL, acme.Authorization{Status: acme.StatusDeactivated}, &disabledAuth)

	return err
}
