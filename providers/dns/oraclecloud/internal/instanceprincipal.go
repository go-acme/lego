package internal

import (
	"crypto/rsa"
	"errors"

	"github.com/nrdcg/oci-go-sdk/common/v1065"
)

// InstancePrincipalConfigurationProvider is used by [auth.GetGenericConfigurationProvider] to set up the real [auth.InstancePrincipalConfigurationProvider].
//
// When the `AuthType` is `instance_principal` and `IsFromConfigFile` is set to true,
// all the other options are ignored.
// https://github.com/oracle/oci-go-sdk/blob/413a2f277f95c5eb76e26a0e0833c396a518bf50/common/auth/utils.go#L91-L92
type InstancePrincipalConfigurationProvider struct{}

func (p *InstancePrincipalConfigurationProvider) PrivateRSAKey() (*rsa.PrivateKey, error) {
	return nil, errors.New("not implemented")
}

func (p *InstancePrincipalConfigurationProvider) KeyID() (string, error) {
	return "", errors.New("not implemented")
}

func (p *InstancePrincipalConfigurationProvider) TenancyOCID() (string, error) {
	return "", errors.New("not implemented")
}

func (p *InstancePrincipalConfigurationProvider) UserOCID() (string, error) {
	return "", errors.New("not implemented")
}

func (p *InstancePrincipalConfigurationProvider) KeyFingerprint() (string, error) {
	return "", errors.New("not implemented")
}

func (p *InstancePrincipalConfigurationProvider) Region() (string, error) {
	return "", errors.New("not implemented")
}

func (p *InstancePrincipalConfigurationProvider) AuthType() (common.AuthConfig, error) {
	return common.AuthConfig{AuthType: common.InstancePrincipal, IsFromConfigFile: true}, nil
}
