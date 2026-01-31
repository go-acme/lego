package certcrypto

import (
	"fmt"
	"strings"
)

// Constants for all key types we support.
const (
	EC256   = KeyType("EC256")
	EC384   = KeyType("EC384")
	RSA2048 = KeyType("RSA2048")
	RSA3072 = KeyType("RSA3072")
	RSA4096 = KeyType("RSA4096")
	RSA8192 = KeyType("RSA8192")
)

// KeyType represents the key algo as well as the key size or curve to use.
type KeyType string

// GetKeyType gets key type from string.
func GetKeyType(keyType string) (KeyType, error) {
	switch strings.ToUpper(keyType) {
	case string(RSA2048):
		return RSA2048, nil
	case string(RSA3072):
		return RSA3072, nil
	case string(RSA4096):
		return RSA4096, nil
	case string(RSA8192):
		return RSA8192, nil
	case string(EC256):
		return EC256, nil
	case string(EC384):
		return EC384, nil
	}

	return "", fmt.Errorf("unsupported key type: %s", keyType)
}
