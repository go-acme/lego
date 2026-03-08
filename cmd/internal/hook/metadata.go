package hook

import (
	"strings"

	"github.com/go-acme/lego/v5/certificate"
	"github.com/go-acme/lego/v5/cmd/internal/storage"
)

// Metadata related to account.
const (
	EnvAccountID    = "LEGO_HOOK_ACCOUNT_ID"
	EnvAccountEmail = "LEGO_HOOK_ACCOUNT_EMAIL"
)

// Metadata related to certificate.
const (
	EnvCertName          = "LEGO_HOOK_CERT_NAME"
	EnvCertNameSanitized = "LEGO_HOOK_CERT_NAME_SANITIZED"
	EnvCertDomains       = "LEGO_HOOK_CERT_DOMAINS"
	EnvCertPath          = "LEGO_HOOK_CERT_PATH"
	EnvCertKeyPath       = "LEGO_HOOK_CERT_KEY_PATH"
	EnvIssuerCertKeyPath = "LEGO_HOOK_ISSUER_CERT_PATH"
	EnvCertPEMPath       = "LEGO_HOOK_CERT_PEM_PATH"
	EnvCertPFXPath       = "LEGO_HOOK_CERT_PFX_PATH"
)

func addAccountMetadata(meta map[string]string, account *storage.Account) {
	meta[EnvAccountID] = account.ID
	meta[EnvAccountEmail] = account.Email
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
