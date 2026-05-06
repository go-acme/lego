package root

import (
	"context"
	"crypto/x509"
	"fmt"

	"github.com/go-acme/lego/v5/certificate"
	"github.com/go-acme/lego/v5/cmd/internal/configuration"
	"github.com/go-acme/lego/v5/cmd/internal/hook"
	"github.com/go-acme/lego/v5/cmd/internal/storage"
	"github.com/go-acme/lego/v5/lego"
)

func obtain(ctx context.Context, lazySetup lzSetUp, certID string, certConfig *configuration.Certificate, certsStorage *storage.CertificatesStorage, hookManager *hook.Manager) error {
	client, err := lazySetup()
	if err != nil {
		return fmt.Errorf("set up client: %w", err)
	}

	if certConfig.CSR != "" {
		return obtainForCSR(ctx, client, certID, certConfig, certsStorage, hookManager)
	}

	return obtainForDomains(ctx, client, certID, certConfig, certsStorage, hookManager)
}

func obtainForDomains(ctx context.Context, client *lego.Client, certID string, certConfig *configuration.Certificate, certsStorage *storage.CertificatesStorage, hookManager *hook.Manager) error {
	request := newObtainRequest(certConfig, certConfig.Domains)

	// NOTE(ldez): I didn't add an option to set a private key as the file.
	// I didn't find a use case for it when using the file configuration.
	// Maybe this can be added in the future.

	err := hookManager.PreForDomains(ctx, certID, request)
	if err != nil {
		return err
	}

	defer func() { _ = hookManager.Post(ctx) }()

	certRes, err := client.Certificate.Obtain(ctx, request)
	if err != nil {
		return err
	}

	if certID != "" {
		certRes.ID = certID
	}

	options := newSaveOptions(certConfig)

	err = certsStorage.Save(
		&storage.Certificate{
			Resource: certRes,
			Origin:   storage.OriginConfiguration,
		},
		options,
	)
	if err != nil {
		return fmt.Errorf("could not save the resource: %w", err)
	}

	return hookManager.Deploy(ctx, certRes, options)
}

func obtainForCSR(ctx context.Context, client *lego.Client, certID string, certConfig *configuration.Certificate, certsStorage *storage.CertificatesStorage, hookManager *hook.Manager) error {
	csr, err := storage.ReadCSRFile(certConfig.CSR)
	if err != nil {
		return err
	}

	// obtain a certificate for this CSR
	request := newObtainForCSRRequest(certConfig, csr)

	// NOTE(ldez): I didn't add an option to set a private key as the file.
	// I didn't find a use case for it when using the file configuration.
	// Maybe this can be added in the future.

	err = hookManager.PreForCSR(ctx, certID, request)
	if err != nil {
		return err
	}

	defer func() { _ = hookManager.Post(ctx) }()

	certRes, err := client.Certificate.ObtainForCSR(ctx, request)
	if err != nil {
		return err
	}

	if certID != "" {
		certRes.ID = certID
	}

	options := newSaveOptions(certConfig)

	err = certsStorage.Save(
		&storage.Certificate{
			Resource: certRes,
			Origin:   storage.OriginConfiguration,
		},
		options,
	)
	if err != nil {
		return fmt.Errorf("could not save the resource: %w", err)
	}

	return hookManager.Deploy(ctx, certRes, options)
}

func newObtainRequest(certConfig *configuration.Certificate, domains []string) certificate.ObtainRequest {
	return certificate.ObtainRequest{
		Domains:                        domains,
		KeyType:                        certConfig.KeyType,
		MustStaple:                     certConfig.MustStaple,
		NotBefore:                      certConfig.NotBefore,
		NotAfter:                       certConfig.NotAfter,
		Bundle:                         !certConfig.NoBundle,
		PreferredChain:                 certConfig.PreferredChain,
		EnableCommonName:               certConfig.EnableCommonName,
		Profile:                        certConfig.Profile,
		AlwaysDeactivateAuthorizations: certConfig.AlwaysDeactivateAuthorizations,
	}
}

func newObtainForCSRRequest(certConfig *configuration.Certificate, csr *x509.CertificateRequest) certificate.ObtainForCSRRequest {
	return certificate.ObtainForCSRRequest{
		CSR:                            csr,
		NotBefore:                      certConfig.NotBefore,
		NotAfter:                       certConfig.NotAfter,
		Bundle:                         !certConfig.NoBundle,
		PreferredChain:                 certConfig.PreferredChain,
		EnableCommonName:               certConfig.EnableCommonName,
		Profile:                        certConfig.Profile,
		AlwaysDeactivateAuthorizations: certConfig.AlwaysDeactivateAuthorizations,
	}
}

func newSaveOptions(certConfig *configuration.Certificate) *storage.SaveOptions {
	opt := &storage.SaveOptions{
		PEM: true,
	}

	if certConfig.PFX != nil {
		opt.PFX = true
		opt.PFXFormat = certConfig.PFX.Format
		opt.PFXPassword = certConfig.PFX.Password
	}

	return opt
}
