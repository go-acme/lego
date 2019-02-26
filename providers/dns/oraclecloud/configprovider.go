package oraclecloud

import (
	"crypto/rsa"
	"encoding/base64"
	"fmt"

	"github.com/oracle/oci-go-sdk/common"
	"github.com/xenolf/lego/platform/config/env"
)

const (
	ociPrivkeyBase64     = "OCI_PRIVKEY_BASE64"
	ociPrivkeyPass       = "OCI_PRIVKEY_PASS"
	ociTenancyOCID       = "OCI_TENANCY_OCID"
	ociUserOCID          = "OCI_USER_OCID"
	ociPubkeyFingerprint = "OCI_PUBKEY_FINGERPRINT"
	ociRegion            = "OCI_REGION"
)

type configProvider struct {
	values               map[string]string
	privateKeyPassphrase string
}

func newConfigProvider(values map[string]string) *configProvider {
	return &configProvider{
		values:               values,
		privateKeyPassphrase: env.GetOrFile(ociPrivkeyPass),
	}
}

func (p *configProvider) PrivateRSAKey() (*rsa.PrivateKey, error) {
	privateKeyDecoded, err := base64.StdEncoding.DecodeString(p.values[ociPrivkeyBase64])
	if err != nil {
		return nil, err
	}

	return common.PrivateKeyFromBytes(privateKeyDecoded, common.String(p.privateKeyPassphrase))
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
	return p.values[ociTenancyOCID], nil
}

func (p *configProvider) UserOCID() (string, error) {
	return p.values[ociUserOCID], nil
}

func (p *configProvider) KeyFingerprint() (string, error) {
	return p.values[ociPubkeyFingerprint], nil
}

func (p *configProvider) Region() (string, error) {
	return p.values[ociRegion], nil
}
