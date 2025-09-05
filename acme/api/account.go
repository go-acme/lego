package api

import (
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/go-acme/lego/v4/acme"
)

type AccountService service

// New Creates a new account.
func (a *AccountService) New(req acme.Account) (acme.ExtendedAccount, error) {
	var account acme.Account
	resp, err := a.core.post(a.core.GetDirectory().NewAccountURL, req, &account)
	location := getLocation(resp)

	if location != "" {
		a.core.jws.SetKid(location)
	}

	if err != nil {
		return acme.ExtendedAccount{Location: location}, err
	}

	return acme.ExtendedAccount{Account: account, Location: location}, nil
}

// NewEAB Creates a new account with an External Account Binding.
func (a *AccountService) NewEAB(accMsg acme.Account, kid, hmacEncoded string) (acme.ExtendedAccount, error) {
	hmac, err := decodeEABHmac(hmacEncoded)
	if err != nil {
		return acme.ExtendedAccount{}, err
	}

	eabJWS, err := a.core.signEABContent(a.core.GetDirectory().NewAccountURL, kid, hmac)
	if err != nil {
		return acme.ExtendedAccount{}, fmt.Errorf("acme: error signing eab content: %w", err)
	}

	accMsg.ExternalAccountBinding = eabJWS

	return a.New(accMsg)
}

// Get Retrieves an account.
func (a *AccountService) Get(accountURL string) (acme.Account, error) {
	if accountURL == "" {
		return acme.Account{}, errors.New("account[get]: empty URL")
	}

	var account acme.Account
	_, err := a.core.postAsGet(accountURL, &account)
	if err != nil {
		return acme.Account{}, err
	}
	return account, nil
}

// Update Updates an account.
func (a *AccountService) Update(accountURL string, req acme.Account) (acme.Account, error) {
	if accountURL == "" {
		return acme.Account{}, errors.New("account[update]: empty URL")
	}

	var account acme.Account
	_, err := a.core.post(accountURL, req, &account)
	if err != nil {
		return acme.Account{}, err
	}

	return account, nil
}

// Deactivate Deactivates an account.
func (a *AccountService) Deactivate(accountURL string) error {
	if accountURL == "" {
		return errors.New("account[deactivate]: empty URL")
	}

	req := acme.Account{Status: acme.StatusDeactivated}
	_, err := a.core.post(accountURL, req, nil)
	return err
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
