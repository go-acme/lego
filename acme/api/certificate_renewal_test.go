package api

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"net/http"
	"testing"
	"time"

	"github.com/go-acme/lego/v5/certcrypto"
	"github.com/go-acme/lego/v5/internal/tester"
	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	ariLeafPEM = `-----BEGIN CERTIFICATE-----
MIIBQzCB66ADAgECAgUAh2VDITAKBggqhkjOPQQDAjAVMRMwEQYDVQQDEwpFeGFt
cGxlIENBMCIYDzAwMDEwMTAxMDAwMDAwWhgPMDAwMTAxMDEwMDAwMDBaMBYxFDAS
BgNVBAMTC2V4YW1wbGUuY29tMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEeBZu
7cbpAYNXZLbbh8rNIzuOoqOOtmxA1v7cRm//AwyMwWxyHz4zfwmBhcSrf47NUAFf
qzLQ2PPQxdTXREYEnKMjMCEwHwYDVR0jBBgwFoAUaYhba4dGQEHhs3uEe6CuLN4B
yNQwCgYIKoZIzj0EAwIDRwAwRAIge09+S5TZAlw5tgtiVvuERV6cT4mfutXIlwTb
+FYN/8oCIClDsqBklhB9KAelFiYt9+6FDj3z4KGVelYM5MdsO3pK
-----END CERTIFICATE-----`
	ariLeafCertID = "aYhba4dGQEHhs3uEe6CuLN4ByNQ.AIdlQyE"
)

func TestMakeCertID(t *testing.T) {
	leaf, err := certcrypto.ParsePEMCertificate([]byte(ariLeafPEM))
	require.NoError(t, err)

	actual, err := MakeARICertID(leaf)
	require.NoError(t, err)
	assert.Equal(t, ariLeafCertID, actual)
}

func TestCertificateService_GetRenewalInfo(t *testing.T) {
	// small value keeps test fast
	privateKey, errK := rsa.GenerateKey(rand.Reader, 1024)
	require.NoError(t, errK, "Could not generate test key")

	server := tester.MockACMEServer().
		Route("GET /renewalInfo/"+ariLeafCertID,
			servermock.RawStringResponse(`{
					"suggestedWindow": {
						"start": "2020-03-17T17:51:09Z",
						"end": "2020-03-17T18:21:09Z"
					},
					"explanationUrl": "https://aricapable.ca.example/docs/renewal-advice/"
				}
			}`).
				WithHeader("Content-Type", "application/json").
				WithHeader("Retry-After", "21600")).
		BuildHTTPS(t)

	core, err := New(server.Client(), "lego-test", server.URL+"/dir", "", privateKey)
	require.NoError(t, err)

	leaf, err := certcrypto.ParsePEMCertificate([]byte(ariLeafPEM))
	require.NoError(t, err)

	ri, err := core.Certificates.GetRenewalInfo(t.Context(), mustMakeARICertID(t, leaf))
	require.NoError(t, err)
	require.NotNil(t, ri)
	assert.Equal(t, "2020-03-17T17:51:09Z", ri.SuggestedWindow.Start.Format(time.RFC3339))
	assert.Equal(t, "2020-03-17T18:21:09Z", ri.SuggestedWindow.End.Format(time.RFC3339))
	assert.Equal(t, "https://aricapable.ca.example/docs/renewal-advice/", ri.ExplanationURL)
	assert.Equal(t, time.Duration(21600000000000), ri.RetryAfter)
}

func TestCertificateService_GetRenewalInfo_retryAfter(t *testing.T) {
	leaf, err := certcrypto.ParsePEMCertificate([]byte(ariLeafPEM))
	require.NoError(t, err)

	server := tester.MockACMEServer().
		Route("GET /renewalInfo/"+ariLeafCertID,
			servermock.RawStringResponse(`{
				"suggestedWindow": {
					"start": "2020-03-17T17:51:09Z",
					"end": "2020-03-17T18:21:09Z"
				},
				"explanationUrl": "https://aricapable.ca.example/docs/renewal-advice/"
			}
		}`).
				WithHeader("Content-Type", "application/json").
				WithHeader("Retry-After", time.Now().UTC().Add(6*time.Hour).Format(time.RFC1123))).
		BuildHTTPS(t)

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err, "Could not generate test key")

	core, err := New(server.Client(), "lego-test", server.URL+"/dir", "", key)
	require.NoError(t, err)

	ri, err := core.Certificates.GetRenewalInfo(t.Context(), mustMakeARICertID(t, leaf))
	require.NoError(t, err)
	require.NotNil(t, ri)
	assert.Equal(t, "2020-03-17T17:51:09Z", ri.SuggestedWindow.Start.Format(time.RFC3339))
	assert.Equal(t, "2020-03-17T18:21:09Z", ri.SuggestedWindow.End.Format(time.RFC3339))
	assert.Equal(t, "https://aricapable.ca.example/docs/renewal-advice/", ri.ExplanationURL)

	assert.InDelta(t, 6, ri.RetryAfter.Hours(), 0.001)
}

func TestCertificateService_GetRenewalInfo_errors(t *testing.T) {
	leaf, err := certcrypto.ParsePEMCertificate([]byte(ariLeafPEM))
	require.NoError(t, err)

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err, "Could not generate test key")

	testCases := []struct {
		desc    string
		timeout time.Duration
		request string
		handler http.HandlerFunc
	}{
		{
			desc:    "API timeout",
			timeout: 500 * time.Millisecond, // HTTP client that times out after 500ms.
			request: mustMakeARICertID(t, leaf),
			handler: func(w http.ResponseWriter, r *http.Request) {
				// API that takes 2ms to respond.
				time.Sleep(2 * time.Millisecond)
			},
		},
		{
			desc:    "API error",
			request: mustMakeARICertID(t, leaf),
			handler: func(w http.ResponseWriter, r *http.Request) {
				// API that responds with error instead of renewal info.
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			server := tester.MockACMEServer().
				Route("GET /renewalInfo/"+ariLeafCertID, test.handler).
				BuildHTTPS(t)

			client := server.Client()

			if test.timeout != 0 {
				client.Timeout = test.timeout
			}

			core, err := New(client, "lego-test", server.URL+"/dir", "", key)
			require.NoError(t, err)

			response, err := core.Certificates.GetRenewalInfo(t.Context(), test.request)
			require.Error(t, err)
			assert.Nil(t, response)
		})
	}
}

func mustMakeARICertID(t *testing.T, leaf *x509.Certificate) string {
	t.Helper()

	certID, err := MakeARICertID(leaf)
	require.NoError(t, err)

	return certID
}
