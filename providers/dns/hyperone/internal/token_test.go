package internal

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Header struct {
	Algorithm string `json:"alg"`
	Type      string `json:"typ"`
	KeyID     string `json:"kid"`
}

func TestPayload_buildToken(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	require.NoError(t, err)

	signer, err := getRSASigner(string(certcrypto.PEMEncode(key)), "sampleKeyId")
	require.NoError(t, err)

	payload := Payload{IssuedAt: 1234, Expiry: 4321, Audience: "api.url", Issuer: "issuer", Subject: "subject"}

	token, err := payload.buildToken(&signer)
	require.NoError(t, err)

	segments := strings.Split(token, ".")
	require.Len(t, segments, 3)

	headerString, err := base64.RawStdEncoding.DecodeString(segments[0])
	require.NoError(t, err)

	var headerStruct Header

	err = json.Unmarshal(headerString, &headerStruct)
	require.NoError(t, err)

	payloadString, err := base64.RawStdEncoding.DecodeString(segments[1])
	require.NoError(t, err)

	var payloadStruct Payload

	err = json.Unmarshal(payloadString, &payloadStruct)
	require.NoError(t, err)

	expectedHeader := Header{Algorithm: "RS256", Type: "JWT", KeyID: "sampleKeyId"}

	assert.Equal(t, expectedHeader, headerStruct)
	assert.Equal(t, payload, payloadStruct)
}
