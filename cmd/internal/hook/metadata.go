package hook

import (
	"strings"

	"github.com/go-acme/lego/v5/certificate"
	"github.com/go-acme/lego/v5/cmd/internal/storage"
)

const envPrefix = "LEGO_HOOK_"

// Metadata related to account.
const (
	EnvAccountID    = envPrefix + "ACCOUNT_ID"
	EnvAccountEmail = envPrefix + "ACCOUNT_EMAIL"
)

// Metadata related to certificate.
const (
	EnvCertName          = envPrefix + "CERT_NAME"
	EnvCertNameSanitized = envPrefix + "CERT_NAME_SANITIZED"
	EnvCertDomains       = envPrefix + "CERT_DOMAINS"
	EnvCertPath          = envPrefix + "CERT_PATH"
	EnvCertKeyPath       = envPrefix + "CERT_KEY_PATH"
	EnvIssuerCertKeyPath = envPrefix + "ISSUER_CERT_PATH"
	EnvCertPEMPath       = envPrefix + "CERT_PEM_PATH"
	EnvCertPFXPath       = envPrefix + "CERT_PFX_PATH"
)

func addAccountMetadata(meta map[string]string, account *storage.Account) {
	meta[EnvAccountID] = account.GetID()
	meta[EnvAccountEmail] = account.GetEmail()
}

func addCertificatePathsMetadata(meta map[string]string, certRes *certificate.Resource, certsStorage *storage.CertificatesStorage, options *storage.SaveOptions) {
	meta[EnvCertPath] = certsStorage.GetFileName(certRes.ID, storage.ExtCert)
	meta[EnvCertKeyPath] = certsStorage.GetFileName(certRes.ID, storage.ExtKey)

	if certRes.IssuerCertificate != nil {
		meta[EnvIssuerCertKeyPath] = certsStorage.GetFileName(certRes.ID, storage.ExtIssuer)
	}

	if options.PEM {
		meta[EnvCertPEMPath] = certsStorage.GetFileName(certRes.ID, storage.ExtPEM)
	}

	if options.PFX {
		meta[EnvCertPFXPath] = certsStorage.GetFileName(certRes.ID, storage.ExtPFX)
	}
}

func addCertificateMetadata(meta map[string]string, certID string, domains []string) {
	if certID == "" {
		meta[EnvCertName] = certID
		meta[EnvCertNameSanitized] = storage.SanitizedName(certID)
	}

	meta[EnvCertDomains] = strings.Join(domains, ",")
}

func metaToEnv(meta map[string]string) []string {
	var envs []string

	for k, v := range meta {
		envs = append(envs, k+"="+v)
	}

	return envs
}
