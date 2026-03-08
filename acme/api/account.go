package api

import (
	"context"
	"crypto"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/go-acme/lego/v5/acme"
	"github.com/go-acme/lego/v5/acme/api/internal/sender"
)

type AccountService service

// New Creates a new account.
func (a *AccountService) New(ctx context.Context, req acme.Account) (acme.ExtendedAccount, error) {
	var account acme.Account

	resp, err := a.core.post(ctx, a.core.GetDirectory().NewAccountURL, req, &account)

	location := sender.GetLocation(resp)

	a.core.setKid(location)

	if err != nil {
		return acme.ExtendedAccount{Location: location}, err
	}

	return acme.ExtendedAccount{Account: account, Location: location}, nil
}

// NewEAB Creates a new account with an External Account Binding.
func (a *AccountService) NewEAB(ctx context.Context, req acme.Account, kid, hmacEncoded string) (acme.ExtendedAccount, error) {
	hmac, err := decodeEABHmac(hmacEncoded)
	if err != nil {
		return acme.ExtendedAccount{}, err
	}

	eabJWS, err := a.core.signer().SignEAB(a.core.GetDirectory().NewAccountURL, kid, hmac)
	if err != nil {
		return acme.ExtendedAccount{}, fmt.Errorf("acme: error signing eab content: %w", err)
	}

	req.ExternalAccountBinding = []byte(eabJWS.FullSerialize())

	return a.New(ctx, req)
}

// Get Retrieves an account.
func (a *AccountService) Get(ctx context.Context, accountURL string) (acme.Account, error) {
	if accountURL == "" {
		return acme.Account{}, errors.New("account[get]: empty URL")
	}

	var account acme.Account

	_, err := a.core.postAsGet(ctx, accountURL, &account)
	if err != nil {
		return acme.Account{}, err
	}

	return account, nil
}

// Update Updates an account.
func (a *AccountService) Update(ctx context.Context, accountURL string, req acme.Account) (acme.Account, error) {
	if accountURL == "" {
		return acme.Account{}, errors.New("account[update]: empty URL")
	}

	var account acme.Account

	_, err := a.core.post(ctx, accountURL, req, &account)
	if err != nil {
		return acme.Account{}, err
	}

	return account, nil
}

// Deactivate Deactivates an account.
func (a *AccountService) Deactivate(ctx context.Context, accountURL string) error {
	if accountURL == "" {
		return errors.New("account[deactivate]: empty URL")
	}

	req := acme.Account{Status: acme.StatusDeactivated}
	_, err := a.core.post(ctx, accountURL, req, nil)

	return err
}

// KeyChange Changes the account key.
func (a *AccountService) KeyChange(ctx context.Context, newKey crypto.Signer) error {
	uri := a.core.GetDirectory().KeyChangeURL

	eabJWS, err := a.core.signer().SignKeyChange(uri, newKey)
	if err != nil {
		return err
	}

	_, err = a.core.retrievablePost(ctx, uri, []byte(eabJWS.FullSerialize()), nil)
	if err != nil {
		return err
	}

	a.core.setPrivateKey(newKey)

	return nil
}

func decodeEABHmac(hmacEncoded string) ([]byte, error) {
	hmac, errRaw := base64.RawURLEncoding.DecodeString(hmacEncoded)
	if errRaw == nil {
		return hmac, nil
	}

	hmac, err := base64.URLEncoding.DecodeString(hmacEncoded)
	if err == nil {
		return hmac, nil
	}

	return nil, fmt.Errorf("acme: could not decode hmac key: %w", errors.Join(errRaw, err))
}
