package secure

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/go-jose/go-jose/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockNonceSource struct{}

func (b *MockNonceSource) Nonce() (string, error) {
	return "xxxNoncexxx", nil
}

func TestSigner_SignContent(t *testing.T) {
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	signer := NewSigner(privKey, "https://example.com")

	content, err := signer.SignContent(&MockNonceSource{}, "https://foo.example", []byte("{}"))
	require.NoError(t, err)

	check(t, content)
}

func TestSigner_SignEAB(t *testing.T) {
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	signer := NewSigner(privKey, "https://example.com")

	content, err := signer.SignEAB("https://foo.example/a", "https://foo.example/b", x509.MarshalPKCS1PrivateKey(privKey))
	require.NoError(t, err)

	check(t, content)
}

func TestSigner_SignKeyChange(t *testing.T) {
	const (
		kid      = "https://example.com/acme/acct/evOfKhNU60wg"
		endpoint = "https://example.com/acme/key-change"
	)

	oldKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	signer := NewSigner(oldKey, kid)

	newKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	require.NoError(t, err)

	content, err := signer.SignKeyChange(endpoint, newKey)
	require.NoError(t, err)

	check(t, content)
}

func check(t *testing.T, content *jose.JSONWebSignature) {
	t.Helper()

	serialized := content.FullSerialize()

	var data map[string]any

	err := json.Unmarshal([]byte(serialized), &data)
	require.NoError(t, err)

	assert.Len(t, data, 3)

	assert.Contains(t, data, "protected")
	assert.Contains(t, data, "payload")
	assert.Contains(t, data, "signature")

	dec, err := base64.RawStdEncoding.DecodeString(data["protected"].(string))
	require.NoError(t, err)

	t.Log("protected:", string(dec))

	dec, err = base64.RawStdEncoding.DecodeString(data["payload"].(string))
	require.NoError(t, err)

	t.Log("payload:", string(dec))
}
