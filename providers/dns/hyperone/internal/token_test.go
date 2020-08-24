package internal

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const privateKey = `-----BEGIN RSA PRIVATE KEY-----
MIICWgIBAAKBgGFfgMY+DuO8l0RYrMLhcl6U/NigNIiOVhoo/xnYyoQALpWxBaBR
+iVJiBUYunQjKA33yAiY0AasCfSn1JB6asayQvGGn73xztLjkeCVLT+9e4nJ0A/o
dK8SOKBg9FFe70KJrWjJd626el0aVDJjtCE+QxJExA0UZbQp+XIyveQXAgMBAAEC
gYBHcL1XNWLRPaWx9GlUVfoGYMMd4HSKl/ueF+QKP59dt5B2LTnWhS7FOqzH5auu
17hkfx3ZCNzfeEuZn6T6F4bMtsQ6A5iT/DeRlG8tOPiCVZ/L0j6IFM78iIUT8XyA
miwnSy1xGSBA67yUmsLxFg2DtGCjamAkY0C5pccadaB7oQJBAKsIPpMXMni+Oo1I
kVxRyoIZgDxsMJiihG2YLVqo8rPtdErl+Lyg3ziVyg9KR6lFMaTBkYBTLoCPof3E
AB/jyucCQQCRv1cVnYNx+bfnXsBlcsCFDV2HkEuLTpxj7hauD4P3GcyLidSsUkn1
PiPunZqKpsQaIoxc/BzTOCcP19ifgqdRAkBJ8Cp9FE4xfKt7YJ/WtVVCoRubA3qO
wdNWPa99vgQOXN0lc/3wLevSXo8XxRjtyIgJndT1EQDNe0qglhcnsiaJAkBziAcR
/VAq0tZys2szf6kYTyXqxfj8Lo5NsHeN9oKXJ346xkEtb/VsT5vQFGJishsU1HoL
Y1W+IO7l4iW3G6xhAkACNwtqxSRRbVsNCUMENpKmYhsyN8QXJ8V+o2A9s+pl21Kz
HIIm179mUYCgO6iAHmkqxlFHFwprUBKdPrmP8qF9
-----END RSA PRIVATE KEY-----`

type Header struct {
	Algorithm string `json:"alg"`
	Type      string `json:"typ"`
	KeyID     string `json:"kid"`
}

func TestPayload_buildToken(t *testing.T) {
	signer, err := getRSASigner(privateKey, "sampleKeyId")
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
