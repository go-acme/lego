package api

import (
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/xenolf/lego/le"
)

type AccountService service

func (a *AccountService) New(req le.Account) (le.ExtendedAccount, error) {
	var account le.Account
	resp, err := a.core.post(a.core.GetDirectory().NewAccountURL, req, &account)
	location := getLocation(resp)

	if len(location) > 0 {
		a.core.jws.SetKid(location)
	}

	if err != nil {
		return le.ExtendedAccount{Location: location}, err
	}

	return le.ExtendedAccount{Account: account, Location: location}, nil
}

func (a *AccountService) NewEAB(accMsg le.Account, kid string, hmacEncoded string) (le.ExtendedAccount, error) {
	hmac, err := base64.RawURLEncoding.DecodeString(hmacEncoded)
	if err != nil {
		return le.ExtendedAccount{}, fmt.Errorf("acme: could not decode hmac key: %v", err)
	}

	eabJWS, err := a.core.signEABContent(a.core.GetDirectory().NewAccountURL, kid, hmac)
	if err != nil {
		return le.ExtendedAccount{}, fmt.Errorf("acme: error signing eab content: %v", err)
	}
	accMsg.ExternalAccountBinding = eabJWS

	return a.New(accMsg)
}

func (a *AccountService) Get(accountURL string) (le.Account, error) {
	if len(accountURL) == 0 {
		return le.Account{}, errors.New("account[get]: empty URL")
	}

	var account le.Account
	_, err := a.core.post(accountURL, le.Account{}, &account)
	if err != nil {
		return le.Account{}, err
	}
	return account, nil
}

func (a *AccountService) Deactivate(accountURL string) error {
	if len(accountURL) == 0 {
		return errors.New("account[deactivate]: empty URL")
	}

	req := le.Account{Status: le.StatusDeactivated}
	_, err := a.core.post(accountURL, req, nil)
	return err
}
