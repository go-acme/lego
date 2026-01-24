package cmd

import (
	"github.com/go-acme/lego/v5/certificate"
	"github.com/go-acme/lego/v5/cmd/internal/hook"
	"github.com/go-acme/lego/v5/cmd/internal/storage"
)

func addPathToMetadata(meta map[string]string, domain string, certRes *certificate.Resource, certsStorage *storage.CertificatesStorage) {
	meta[hook.EnvCertDomain] = domain
	meta[hook.EnvCertPath] = certsStorage.GetFileName(domain, storage.ExtCert)
	meta[hook.EnvCertKeyPath] = certsStorage.GetFileName(domain, storage.ExtKey)

	if certRes.IssuerCertificate != nil {
		meta[hook.EnvIssuerCertKeyPath] = certsStorage.GetFileName(domain, storage.ExtIssuer)
	}

	if certsStorage.IsPEM() {
		meta[hook.EnvCertPEMPath] = certsStorage.GetFileName(domain, storage.ExtPEM)
	}

	if certsStorage.IsPFX() {
		meta[hook.EnvCertPFXPath] = certsStorage.GetFileName(domain, storage.ExtPFX)
	}
}
