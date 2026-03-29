package certcrypto

import (
	"fmt"
	"slices"

	"software.sslmate.com/src/go-pkcs12"
)

// Constants for all PKCS#12 encryption format we support.
const (
	PKCS12LegacyDES  = "DES"
	PKCS12LegacyRC2  = "RC2"
	PKCS12Modern2023 = "SHA256"
	PKCS12Modern2026 = "PBMAC1"
)

func AllPKCS12Formats() []string {
	return []string{
		PKCS12LegacyDES,
		PKCS12LegacyRC2,
		PKCS12Modern2023,
		PKCS12Modern2026,
	}
}

// IsPKCS12Supported checks if the provided format is a supported PKCS#12 encryption format.
func IsPKCS12Supported(format string) bool {
	return slices.Contains(AllPKCS12Formats(), format)
}

// GetPKCS12Encoder returns a PKCS12 encoder based on the provided format.
func GetPKCS12Encoder(format string) (*pkcs12.Encoder, error) {
	var encoder *pkcs12.Encoder

	switch format {
	case PKCS12Modern2026:
		encoder = pkcs12.Modern2026
	case PKCS12Modern2023:
		encoder = pkcs12.Modern2023
	case PKCS12LegacyDES:
		encoder = pkcs12.LegacyDES
	case PKCS12LegacyRC2:
		encoder = pkcs12.LegacyRC2
	default:
		return nil, fmt.Errorf("invalid PKCS12 format: %s", format)
	}

	return encoder, nil
}
