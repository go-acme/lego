package certificate

import (
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"testing"
	"time"

	"github.com/go-acme/lego/v4/acme"
	"github.com/go-acme/lego/v4/acme/api"
	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/go-acme/lego/v4/platform/tester/servermock"
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

func Test_makeCertID(t *testing.T) {
	leaf, err := certcrypto.ParsePEMCertificate([]byte(ariLeafPEM))
	require.NoError(t, err)

	actual, err := MakeARICertID(leaf)
	require.NoError(t, err)
	assert.Equal(t, ariLeafCertID, actual)
}

func TestCertifier_GetRenewalInfo(t *testing.T) {
	leaf, err := certcrypto.ParsePEMCertificate([]byte(ariLeafPEM))
	require.NoError(t, err)

	// Test with a fake API.
	apiURL := tester.MockACMEServer().
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
		Build(t)

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err, "Could not generate test key")

	core, err := api.New(http.DefaultClient, "lego-test", apiURL+"/dir", "", key)
	require.NoError(t, err)

	certifier := NewCertifier(core, &resolverMock{}, CertifierOptions{KeyType: certcrypto.RSA2048})

	ri, err := certifier.GetRenewalInfo(RenewalInfoRequest{leaf})
	require.NoError(t, err)
	require.NotNil(t, ri)
	assert.Equal(t, "2020-03-17T17:51:09Z", ri.SuggestedWindow.Start.Format(time.RFC3339))
	assert.Equal(t, "2020-03-17T18:21:09Z", ri.SuggestedWindow.End.Format(time.RFC3339))
	assert.Equal(t, "https://aricapable.ca.example/docs/renewal-advice/", ri.ExplanationURL)
	assert.Equal(t, time.Duration(21600000000000), ri.RetryAfter)
}

func TestCertifier_GetRenewalInfo_errors(t *testing.T) {
	leaf, err := certcrypto.ParsePEMCertificate([]byte(ariLeafPEM))
	require.NoError(t, err)

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err, "Could not generate test key")

	testCases := []struct {
		desc       string
		httpClient *http.Client
		request    RenewalInfoRequest
		handler    http.HandlerFunc
	}{
		{
			desc:       "API timeout",
			httpClient: &http.Client{Timeout: 500 * time.Millisecond}, // HTTP client that times out after 500ms.
			request:    RenewalInfoRequest{leaf},
			handler: func(w http.ResponseWriter, r *http.Request) {
				// API that takes 2ms to respond.
				time.Sleep(2 * time.Millisecond)
			},
		},
		{
			desc:       "API error",
			httpClient: http.DefaultClient,
			request:    RenewalInfoRequest{leaf},
			handler: func(w http.ResponseWriter, r *http.Request) {
				// API that responds with error instead of renewal info.
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			apiURL := tester.MockACMEServer().
				Route("GET /renewalInfo/"+ariLeafCertID, test.handler).
				Build(t)

			core, err := api.New(test.httpClient, "lego-test", apiURL+"/dir", "", key)
			require.NoError(t, err)

			certifier := NewCertifier(core, &resolverMock{}, CertifierOptions{KeyType: certcrypto.RSA2048})

			response, err := certifier.GetRenewalInfo(test.request)
			require.Error(t, err)
			assert.Nil(t, response)
		})
	}
}

func TestRenewalInfoResponse_ShouldRenew(t *testing.T) {
	now := time.Now().UTC()

	t.Run("Window is in the past", func(t *testing.T) {
		ri := RenewalInfoResponse{
			RenewalInfoResponse: acme.RenewalInfoResponse{
				SuggestedWindow: acme.Window{
					Start: now.Add(-2 * time.Hour),
					End:   now.Add(-1 * time.Hour),
				},
				ExplanationURL: "",
			},
			RetryAfter: 0,
		}

		rt := ri.ShouldRenewAt(now, 0)
		require.NotNil(t, rt)
		assert.Equal(t, now, *rt)
	})

	t.Run("Window is in the future", func(t *testing.T) {
		ri := RenewalInfoResponse{
			RenewalInfoResponse: acme.RenewalInfoResponse{
				SuggestedWindow: acme.Window{
					Start: now.Add(1 * time.Hour),
					End:   now.Add(2 * time.Hour),
				},
				ExplanationURL: "",
			},
			RetryAfter: 0,
		}

		rt := ri.ShouldRenewAt(now, 0)
		assert.Nil(t, rt)
	})

	t.Run("Window is in the future, but caller is willing to sleep", func(t *testing.T) {
		ri := RenewalInfoResponse{
			RenewalInfoResponse: acme.RenewalInfoResponse{
				SuggestedWindow: acme.Window{
					Start: now.Add(1 * time.Hour),
					End:   now.Add(2 * time.Hour),
				},
				ExplanationURL: "",
			},
			RetryAfter: 0,
		}

		rt := ri.ShouldRenewAt(now, 2*time.Hour)
		require.NotNil(t, rt)
		assert.True(t, rt.Before(now.Add(2*time.Hour)))
	})

	t.Run("Window is in the future, but caller isn't willing to sleep long enough", func(t *testing.T) {
		ri := RenewalInfoResponse{
			RenewalInfoResponse: acme.RenewalInfoResponse{
				SuggestedWindow: acme.Window{
					Start: now.Add(1 * time.Hour),
					End:   now.Add(2 * time.Hour),
				},
				ExplanationURL: "",
			},
			RetryAfter: 0,
		}

		rt := ri.ShouldRenewAt(now, 59*time.Minute)
		assert.Nil(t, rt)
	})
}
