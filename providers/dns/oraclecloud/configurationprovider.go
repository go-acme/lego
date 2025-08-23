package oraclecloud

import (
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/go-acme/lego/v4/log"
	"github.com/go-acme/lego/v4/platform/config/env"
	"github.com/nrdcg/oci-go-sdk/common/v1065"
)

type environmentConfigurationProvider struct {
	values map[string]string
}

func newEnvironmentConfigurationProvider() (*environmentConfigurationProvider, error) {
	values, err := env.GetWithFallback(
		[]string{EnvRegion, altEnvTFVarRegion},
		[]string{EnvUserOCID, altEnvTFVarUserOCID},
		[]string{EnvTenancyOCID, altEnvTFVarTenancyOCID},
		[]string{EnvPubKeyFingerprint, altEnvFingerprint, altEnvTFVarFingerprint},
	)
	if err != nil {
		return nil, err
	}

	return &environmentConfigurationProvider{
		values: values,
	}, nil
}

func (p *environmentConfigurationProvider) PrivateRSAKey() (*rsa.PrivateKey, error) {
	privateKey, err := getPrivateKey()
	if err != nil {
		return nil, err
	}

	return common.PrivateKeyFromBytesWithPassword(privateKey, []byte(p.privateKeyPassword()))
}

func (p *environmentConfigurationProvider) KeyID() (string, error) {
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

func (p *environmentConfigurationProvider) TenancyOCID() (string, error) {
	return p.values[EnvTenancyOCID], nil
}

func (p *environmentConfigurationProvider) UserOCID() (string, error) {
	return p.values[EnvUserOCID], nil
}

func (p *environmentConfigurationProvider) KeyFingerprint() (string, error) {
	return p.values[EnvPubKeyFingerprint], nil
}

func (p *environmentConfigurationProvider) Region() (string, error) {
	return p.values[EnvRegion], nil
}

func (p *environmentConfigurationProvider) AuthType() (common.AuthConfig, error) {
	// Inspired by https://github.com/oracle/oci-go-sdk/blob/e7635c292e60d0a9dcdd3a1e7de180d7c99b1eee/common/configuration.go#L231-L234
	return common.AuthConfig{AuthType: common.UnknownAuthenticationType}, errors.New("unsupported, keep the interface")
}

func (p *environmentConfigurationProvider) privateKeyPassword() string {
	return env.GetOneWithFallback(EnvPrivKeyPass, "", env.ParseString, altEnvPrivateKeyPassword, altEnvTFVarPrivateKeyPassword)
}

func getPrivateKey() ([]byte, error) {
	base64EnvKeys := []string{envPrivKey, altEnvPrivateKey}

	envVarValue := getEnvWithStrictFallback(base64EnvKeys...)
	if envVarValue != "" {
		bytes, err := base64.StdEncoding.DecodeString(envVarValue)
		if err != nil {
			return nil, fmt.Errorf("failed to read base64 value %s (defined by env vars %s): %w", envVarValue,
				strings.Join(base64EnvKeys, " or "), err)
		}

		return bytes, nil
	}

	fileEnvKeys := []string{EnvPrivKeyFile, altEnvPrivateKeyPath, altEnvTFVarPrivateKeyPath}

	fileVarValue := getEnvFileWithStrictFallback(fileEnvKeys...)
	if len(fileVarValue) == 0 {
		return nil, fmt.Errorf("no value provided for: %s",
			strings.Join(slices.Concat(base64EnvKeys, fileEnvKeys), " or "),
		)
	}

	return fileVarValue, nil
}

func getEnvWithStrictFallback(keys ...string) string {
	for _, key := range keys {
		envVarValue := os.Getenv(key)
		if envVarValue != "" {
			return envVarValue
		}
	}

	return ""
}

func getEnvFileWithStrictFallback(keys ...string) []byte {
	for _, key := range keys {
		fileVarValue := os.Getenv(key)
		if fileVarValue == "" {
			continue
		}

		fileContents, err := os.ReadFile(fileVarValue)
		if err != nil {
			log.Printf("Failed to read the file %s (defined by env var %s): %s", fileVarValue, key, err)
			return nil
		}

		return fileContents
	}

	return nil
}
