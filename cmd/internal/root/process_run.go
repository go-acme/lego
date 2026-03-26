package root

import (
	"context"
	"fmt"

	"github.com/go-acme/lego/v5/certcrypto"
	"github.com/go-acme/lego/v5/certificate"
	"github.com/go-acme/lego/v5/cmd/internal/configuration"
	"github.com/go-acme/lego/v5/cmd/internal/storage"
	"github.com/go-acme/lego/v5/lego"
)

func runCertificate(ctx context.Context, lazySetup lzSetUp, certConfig *configuration.Certificate, certsStorage *storage.CertificatesStorage) error {
	client, err := lazySetup()
	if err != nil {
		return err
	}

	certRes, err := obtainCertificate(ctx, client, certConfig)
	if err != nil {
		return err
	}

	err = certsStorage.Save(certRes, &storage.SaveOptions{PEM: true})
	if err != nil {
		return fmt.Errorf("could not save the resource: %w", err)
	}

	return nil
}

func obtainCertificate(ctx context.Context, client *lego.Client, certConfig *configuration.Certificate) (*storage.Certificate, error) {
	if certConfig.CSR != "" {
		csr, err := storage.ReadCSRFile(certConfig.CSR)
		if err != nil {
			return nil, err
		}

		// obtain a certificate for this CSR
		request := certificate.ObtainForCSRRequest{
			CSR:                            csr,
			NotBefore:                      certConfig.NotBefore,
			NotAfter:                       certConfig.NotAfter,
			Bundle:                         !certConfig.NoBundle,
			PreferredChain:                 certConfig.PreferredChain,
			EnableCommonName:               certConfig.EnableCommonName,
			Profile:                        certConfig.Profile,
			AlwaysDeactivateAuthorizations: certConfig.AlwaysDeactivateAuthorizations,
		}

		// NOTE(ldez): I didn't add an option to set a private key as the file.
		// I didn't find a use case for it when using the file configuration.
		// Maybe this can be added in the future.

		certRes, err := client.Certificate.ObtainForCSR(ctx, request)
		if err != nil {
			return nil, err
		}

		return &storage.Certificate{
			Resource: certRes,
			Origin:   storage.OriginConfiguration,
		}, nil
	}

	keyType, err := certcrypto.ToKeyType(certConfig.KeyType)
	if err != nil {
		return nil, fmt.Errorf("get the key type: %w", err)
	}

	request := certificate.ObtainRequest{
		Domains:                        certConfig.Domains,
		KeyType:                        keyType,
		MustStaple:                     certConfig.MustStaple,
		NotBefore:                      certConfig.NotBefore,
		NotAfter:                       certConfig.NotAfter,
		Bundle:                         !certConfig.NoBundle,
		PreferredChain:                 certConfig.PreferredChain,
		EnableCommonName:               certConfig.EnableCommonName,
		Profile:                        certConfig.Profile,
		AlwaysDeactivateAuthorizations: certConfig.AlwaysDeactivateAuthorizations,
	}

	// NOTE(ldez): I didn't add an option to set a private key as the file.
	// I didn't find a use case for it when using the file configuration.
	// Maybe this can be added in the future.

	certRes, err := client.Certificate.Obtain(ctx, request)
	if err != nil {
		return nil, err
	}

	return &storage.Certificate{
		Resource: certRes,
		Origin:   storage.OriginConfiguration,
	}, nil
}
