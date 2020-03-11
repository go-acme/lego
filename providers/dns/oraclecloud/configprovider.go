package oraclecloud

import (
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/go-acme/lego/v3/platform/config/env"
	"github.com/oracle/oci-go-sdk/common"
)

type configProvider struct {
	values               map[string]string
	privateKeyPassphrase string
}

func newConfigProvider(values map[string]string) *configProvider {
	return &configProvider{
		values:               values,
		privateKeyPassphrase: env.GetOrFile(EnvPrivKeyPass),
	}
}

func (p *configProvider) PrivateRSAKey() (*rsa.PrivateKey, error) {
	privateKey, err := getPrivateKey(envPrivKey)
	if err != nil {
		return nil, err
	}

	return common.PrivateKeyFromBytes(privateKey, common.String(p.privateKeyPassphrase))
}

func (p *configProvider) KeyID() (string, error) {
	tenancy, err := p.TenancyOCID()
	if err != nil {
		return "", err
	}

	user, err := p.UserOCID()
	if err != nil {
		return "", err
	}

	fingerprint, err := p.KeyFingerprint()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s/%s", tenancy, user, fingerprint), nil
}

func (p *configProvider) TenancyOCID() (value string, err error) {
	return p.values[EnvTenancyOCID], nil
}

func (p *configProvider) UserOCID() (string, error) {
	return p.values[EnvUserOCID], nil
}

func (p *configProvider) KeyFingerprint() (string, error) {
	return p.values[EnvPubKeyFingerprint], nil
}

func (p *configProvider) Region() (string, error) {
	return p.values[EnvRegion], nil
}

func getPrivateKey(envVar string) ([]byte, error) {
	envVarValue := os.Getenv(envVar)
	if envVarValue != "" {
		bytes, err := base64.StdEncoding.DecodeString(envVarValue)
		if err != nil {
			return nil, fmt.Errorf("failed to read base64 value %s (defined by env var %s): %w", envVarValue, envVar, err)
		}
		return bytes, nil
	}

	fileVar := envVar + "_FILE"
	fileVarValue := os.Getenv(fileVar)
	if fileVarValue == "" {
		return nil, fmt.Errorf("no value provided for: %s or %s", envVar, fileVar)
	}

	fileContents, err := ioutil.ReadFile(fileVarValue)
	if err != nil {
		return nil, fmt.Errorf("failed to read the file %s (defined by env var %s): %w", fileVarValue, fileVar, err)
	}

	return fileContents, nil
}
