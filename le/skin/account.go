package skin

import (
	"encoding/base64"
	"fmt"

	"github.com/xenolf/lego/le"
)

type AccountService service

func (a *AccountService) New(req le.AccountMessage) (le.AccountExtend, error) {
	var account le.AccountMessage
	resp, err := a.core.post(a.core.GetDirectory().NewAccountURL, req, &account)
	location := getLocation(resp)

	// TODO good or not?
	if len(location) > 0 {
		a.core.jws.SetKid(location)
	}

	if err != nil {
		return le.AccountExtend{Location: location}, err
	}

	return le.AccountExtend{AccountMessage: account, Location: location}, nil
}

func (a *AccountService) NewEAB(accMsg le.AccountMessage, kid string, hmacEncoded string) (le.AccountExtend, error) {
	hmac, err := base64.RawURLEncoding.DecodeString(hmacEncoded)
	if err != nil {
		return le.AccountExtend{}, fmt.Errorf("acme: could not decode hmac key: %v", err)
	}

	eabJWS, err := a.core.signEABContent(a.core.GetDirectory().NewAccountURL, kid, hmac)
	if err != nil {
		return le.AccountExtend{}, fmt.Errorf("acme: error signing eab content: %v", err)
	}
	accMsg.ExternalAccountBinding = eabJWS

	return a.New(accMsg)
}

func (a *AccountService) Get(accountURL string) (le.AccountMessage, error) {
	var account le.AccountMessage
	_, err := a.core.post(accountURL, le.AccountMessage{}, &account)
	if err != nil {
		return le.AccountMessage{}, err
	}
	return account, nil
}

func (a *AccountService) Deactivate(accountURL string) error {
	req := le.AccountMessage{Status: le.StatusDeactivated}
	_, err := a.core.post(accountURL, req, nil)
	return err
}
