package tlsalpn01

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/subtle"
	"crypto/tls"
	"encoding/asn1"
	"net/http"
	"testing"

	"github.com/go-acme/lego/v4/acme"
	"github.com/go-acme/lego/v4/acme/api"
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChallenge(t *testing.T) {
	_, apiURL := tester.SetupFakeAPI(t)

	domain := "localhost:23457"

	mockValidate := func(_ *api.Core, _ string, chlng acme.Challenge) error {
		conn, err := tls.Dial("tcp", domain, &tls.Config{
			InsecureSkipVerify: true,
		})
		require.NoError(t, err, "Expected to connect to challenge server without an error")

		// Expect the server to only return one certificate
		connState := conn.ConnectionState()
		assert.Len(t, connState.PeerCertificates, 1, "Expected the challenge server to return exactly one certificate")

		remoteCert := connState.PeerCertificates[0]
		assert.Len(t, remoteCert.DNSNames, 1, "Expected the challenge certificate to have exactly one DNSNames entry")
		assert.Equal(t, domain, remoteCert.DNSNames[0], "challenge certificate DNSName ")
		assert.NotEmpty(t, remoteCert.Extensions, "Expected the challenge certificate to contain extensions")

		idx := -1
		for i, ext := range remoteCert.Extensions {
			if idPeAcmeIdentifierV1.Equal(ext.Id) {
				idx = i
				break
			}
		}

		require.NotEqual(t, -1, idx, "Expected the challenge certificate to contain an extension with the id-pe-acmeIdentifier id,")

		ext := remoteCert.Extensions[idx]
		assert.True(t, ext.Critical, "Expected the challenge certificate id-pe-acmeIdentifier extension to be marked as critical")

		zBytes := sha256.Sum256([]byte(chlng.KeyAuthorization))
		value, err := asn1.Marshal(zBytes[:sha256.Size])
		require.NoError(t, err, "Expected marshaling of the keyAuth to return no error")

		if subtle.ConstantTimeCompare(value, ext.Value) != 1 {
			t.Errorf("Expected the challenge certificate id-pe-acmeIdentifier extension to contain the SHA-256 digest of the keyAuth, %v, but was %v", zBytes[:], ext.Value)
		}

		return nil
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 512)
	require.NoError(t, err, "Could not generate test key")

	core, err := api.New(http.DefaultClient, "lego-test", apiURL+"/dir", "", privateKey)
	require.NoError(t, err)

	solver := NewChallenge(
		core,
		mockValidate,
		&ProviderServer{port: "23457"},
	)

	authz := acme.Authorization{
		Identifier: acme.Identifier{
			Value: domain,
		},
		Challenges: []acme.Challenge{
			{Type: challenge.TLSALPN01.String(), Token: "tlsalpn1"},
		},
	}

	err = solver.Solve(authz)
	require.NoError(t, err)
}

func TestChallengeInvalidPort(t *testing.T) {
	_, apiURL := tester.SetupFakeAPI(t)

	privateKey, err := rsa.GenerateKey(rand.Reader, 128)
	require.NoError(t, err, "Could not generate test key")

	core, err := api.New(http.DefaultClient, "lego-test", apiURL+"/dir", "", privateKey)
	require.NoError(t, err)

	solver := NewChallenge(
		core,
		func(_ *api.Core, _ string, _ acme.Challenge) error { return nil },
		&ProviderServer{port: "123456"},
	)

	authz := acme.Authorization{
		Identifier: acme.Identifier{
			Value: "localhost:123456",
		},
		Challenges: []acme.Challenge{
			{Type: challenge.TLSALPN01.String(), Token: "tlsalpn1"},
		},
	}

	err = solver.Solve(authz)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid port")
	assert.Contains(t, err.Error(), "123456")
}
