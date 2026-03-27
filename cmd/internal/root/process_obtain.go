package root

import (
	"context"
	"crypto/x509"
	"fmt"

	"github.com/go-acme/lego/v5/certificate"
	"github.com/go-acme/lego/v5/cmd/internal/configuration"
	"github.com/go-acme/lego/v5/cmd/internal/storage"
	"github.com/go-acme/lego/v5/lego"
)

func obtain(ctx context.Context, lazySetup lzSetUp, certID string, certConfig *configuration.Certificate, certsStorage *storage.CertificatesStorage) error {
	client, err := lazySetup()
	if err != nil {
		return fmt.Errorf("set up client: %w", err)
	}

	if certConfig.CSR != "" {
		return obtainForCSR(ctx, client, certID, certConfig, certsStorage)
	}

	return obtainForDomains(ctx, client, certID, certConfig, certsStorage)
}

func obtainForDomains(ctx context.Context, client *lego.Client, certID string, certConfig *configuration.Certificate, certsStorage *storage.CertificatesStorage) error {
	request := newObtainRequest(certConfig, certConfig.Domains)

	// NOTE(ldez): I didn't add an option to set a private key as the file.
	// I didn't find a use case for it when using the file configuration.
	// Maybe this can be added in the future.

	certRes, err := client.Certificate.Obtain(ctx, request)
	if err != nil {
		return err
	}

	if certID != "" {
		certRes.ID = certID
	}

	err = certsStorage.Save(
		&storage.Certificate{
			Resource: certRes,
			Origin:   storage.OriginConfiguration,
		},
		&storage.SaveOptions{PEM: true},
	)
	if err != nil {
		return fmt.Errorf("could not save the resource: %w", err)
	}

	return nil
}

func obtainForCSR(ctx context.Context, client *lego.Client, certID string, certConfig *configuration.Certificate, certsStorage *storage.CertificatesStorage) error {
	csr, err := storage.ReadCSRFile(certConfig.CSR)
	if err != nil {
		return err
	}

	// obtain a certificate for this CSR
	request := newObtainForCSRRequest(certConfig, csr)

	// NOTE(ldez): I didn't add an option to set a private key as the file.
	// I didn't find a use case for it when using the file configuration.
	// Maybe this can be added in the future.

	certRes, err := client.Certificate.ObtainForCSR(ctx, request)
	if err != nil {
		return err
	}

	if certID != "" {
		certRes.ID = certID
	}

	err = certsStorage.Save(
		&storage.Certificate{
			Resource: certRes,
			Origin:   storage.OriginConfiguration,
		},
		&storage.SaveOptions{PEM: true},
	)
	if err != nil {
		return fmt.Errorf("could not save the resource: %w", err)
	}

	return nil
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
