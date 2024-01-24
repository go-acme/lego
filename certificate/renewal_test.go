package certificate

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/go-acme/lego/v4/acme"
	"github.com/go-acme/lego/v4/acme/api"
	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/go-jose/go-jose/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	ariLeafPEM = `-----BEGIN CERTIFICATE-----
MIIDMDCCAhigAwIBAgIIPqNFaGVEHxwwDQYJKoZIhvcNAQELBQAwIDEeMBwGA1UE
AxMVbWluaWNhIHJvb3QgY2EgM2ExMzU2MB4XDTIyMDMxNzE3NTEwOVoXDTI0MDQx
NjE3NTEwOVowFjEUMBIGA1UEAxMLZXhhbXBsZS5jb20wggEiMA0GCSqGSIb3DQEB
AQUAA4IBDwAwggEKAoIBAQCgm9K/c+il2Pf0f8qhgxn9SKqXq88cOm9ov9AVRbPA
OWAAewqX2yUAwI4LZBGEgzGzTATkiXfoJ3cN3k39cH6tBbb3iSPuEn7OZpIk9D+e
3Q9/hX+N/jlWkaTB/FNA+7aE5IVWhmdczYilXa10V9r+RcvACJt0gsipBZVJ4jfJ
HnWJJGRZzzxqG/xkQmpXxZO7nOPFc8SxYKWdfcgp+rjR2ogYhSz7BfKoVakGPbpX
vZOuT9z4kkHra/WjwlkQhtHoTXdAxH3qC2UjMzO57Tx+otj0CxAv9O7CTJXISywB
vEVcmTSZkHS3eZtvvIwPx7I30ITRkYk/tLl1MbyB3SiZAgMBAAGjeDB2MA4GA1Ud
DwEB/wQEAwIFoDAdBgNVHSUEFjAUBggrBgEFBQcDAQYIKwYBBQUHAwIwDAYDVR0T
AQH/BAIwADAfBgNVHSMEGDAWgBQ4zzDRUaXHVKqlSTWkULGU4zGZpTAWBgNVHREE
DzANggtleGFtcGxlLmNvbTANBgkqhkiG9w0BAQsFAAOCAQEAx0aYvmCk7JYGNEXe
+hrOfKawkHYzWvA92cI/Oi6h+oSdHZ2UKzwFNf37cVKZ37FCrrv5pFP/xhhHvrNV
EnOx4IaF7OrnaTu5miZiUWuvRQP7ZGmGNFYbLTEF6/dj+WqyYdVaWzxRqHFu1ptC
TXysJCeyiGnR+KOOjOOQ9ZlO5JUK3OE4hagPLfaIpDDy6RXQt3ss0iNLuB1+IOtp
1URpvffLZQ8xPsEgOZyPWOcabTwJrtqBwily+lwPFn2mChUx846LwQfxtsXU/lJg
HX2RteNJx7YYNeX3Uf960mgo5an6vE8QNAsIoNHYrGyEmXDhTRe9mCHyiW2S7fZq
o9q12g==
-----END CERTIFICATE-----`
	ariLeafCertID = "OM8w0VGlx1SqpUk1pFCxlOMxmaU.PqNFaGVEHxw"
)

func Test_makeCertID(t *testing.T) {
	leaf, err := certcrypto.ParsePEMCertificate([]byte(ariLeafPEM))
	require.NoError(t, err)

	actual, err := makeARICertID(leaf)
	require.NoError(t, err)
	assert.Equal(t, ariLeafCertID, actual)
}

func TestCertifier_GetRenewalInfo(t *testing.T) {
	leaf, err := certcrypto.ParsePEMCertificate([]byte(ariLeafPEM))
	require.NoError(t, err)

	// Test with a fake API.
	mux, apiURL := tester.SetupFakeAPI(t)
	mux.HandleFunc("/renewalInfo/"+ariLeafCertID, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, wErr := w.Write([]byte(`{
				"suggestedWindow": {
					"start": "2020-03-17T17:51:09Z",
					"end": "2020-03-17T18:21:09Z"
				},
				"explanationUrl": "https://aricapable.ca/docs/renewal-advice/"
			}
		}`))
		require.NoError(t, wErr)
	})

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
	assert.Equal(t, "https://aricapable.ca/docs/renewal-advice/", ri.ExplanationURL)
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
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			mux, apiURL := tester.SetupFakeAPI(t)
			mux.HandleFunc("/renewalInfo/"+ariLeafCertID, test.handler)

			core, err := api.New(test.httpClient, "lego-test", apiURL+"/dir", "", key)
			require.NoError(t, err)

			certifier := NewCertifier(core, &resolverMock{}, CertifierOptions{KeyType: certcrypto.RSA2048})

			response, err := certifier.GetRenewalInfo(test.request)
			require.Error(t, err)
			assert.Nil(t, response)
		})
	}
}

func TestCertifier_UpdateRenewalInfo(t *testing.T) {
	leaf, err := certcrypto.ParsePEMCertificate([]byte(ariLeafPEM))
	require.NoError(t, err)

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err, "Could not generate test key")

	// Test with a fake API.
	mux, apiURL := tester.SetupFakeAPI(t)
	mux.HandleFunc("/renewalInfo", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		body, rsbErr := readSignedBody(r, key)
		if rsbErr != nil {
			http.Error(w, rsbErr.Error(), http.StatusBadRequest)
			return
		}

		var req acme.RenewalInfoUpdateRequest
		err = json.Unmarshal(body, &req)
		assert.NoError(t, err)
		assert.True(t, req.Replaced)
		assert.Equal(t, ariLeafCertID, req.CertID)

		w.WriteHeader(http.StatusOK)
	})

	core, err := api.New(http.DefaultClient, "lego-test", apiURL+"/dir", "", key)
	require.NoError(t, err)

	certifier := NewCertifier(core, &resolverMock{}, CertifierOptions{KeyType: certcrypto.RSA2048})

	err = certifier.UpdateRenewalInfo(RenewalInfoRequest{leaf})
	require.NoError(t, err)
}

func TestCertifier_UpdateRenewalInfo_errors(t *testing.T) {
	leaf, err := certcrypto.ParsePEMCertificate([]byte(ariLeafPEM))
	require.NoError(t, err)

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err, "Could not generate test key")

	testCases := []struct {
		desc    string
		request RenewalInfoRequest
	}{
		{
			desc:    "API error",
			request: RenewalInfoRequest{leaf},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			mux, apiURL := tester.SetupFakeAPI(t)

			// Always returns an error.
			mux.HandleFunc("/renewalInfo", func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			})

			core, err := api.New(http.DefaultClient, "lego-test", apiURL+"/dir", "", key)
			require.NoError(t, err)

			certifier := NewCertifier(core, &resolverMock{}, CertifierOptions{KeyType: certcrypto.RSA2048})

			err = certifier.UpdateRenewalInfo(test.request)
			require.Error(t, err)
		})
	}
}

func TestRenewalInfoResponse_ShouldRenew(t *testing.T) {
	now := time.Now().UTC()

	t.Run("Window is in the past", func(t *testing.T) {
		ri := RenewalInfoResponse{
			acme.RenewalInfoResponse{
				SuggestedWindow: acme.Window{
					Start: now.Add(-2 * time.Hour),
					End:   now.Add(-1 * time.Hour),
				},
				ExplanationURL: "",
			},
		}

		rt := ri.ShouldRenewAt(now, 0)
		require.NotNil(t, rt)
		assert.Equal(t, now, *rt)
	})

	t.Run("Window is in the future", func(t *testing.T) {
		ri := RenewalInfoResponse{
			acme.RenewalInfoResponse{
				SuggestedWindow: acme.Window{
					Start: now.Add(1 * time.Hour),
					End:   now.Add(2 * time.Hour),
				},
				ExplanationURL: "",
			},
		}

		rt := ri.ShouldRenewAt(now, 0)
		assert.Nil(t, rt)
	})

	t.Run("Window is in the future, but caller is willing to sleep", func(t *testing.T) {
		ri := RenewalInfoResponse{
			acme.RenewalInfoResponse{
				SuggestedWindow: acme.Window{
					Start: now.Add(1 * time.Hour),
					End:   now.Add(2 * time.Hour),
				},
				ExplanationURL: "",
			},
		}

		rt := ri.ShouldRenewAt(now, 2*time.Hour)
		require.NotNil(t, rt)
		assert.True(t, rt.Before(now.Add(2*time.Hour)))
	})

	t.Run("Window is in the future, but caller isn't willing to sleep long enough", func(t *testing.T) {
		ri := RenewalInfoResponse{
			acme.RenewalInfoResponse{
				SuggestedWindow: acme.Window{
					Start: now.Add(1 * time.Hour),
					End:   now.Add(2 * time.Hour),
				},
				ExplanationURL: "",
			},
		}

		rt := ri.ShouldRenewAt(now, 59*time.Minute)
		assert.Nil(t, rt)
	})
}

func readSignedBody(r *http.Request, privateKey *rsa.PrivateKey) ([]byte, error) {
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	jws, err := jose.ParseSigned(string(reqBody))
	if err != nil {
		return nil, err
	}

	body, err := jws.Verify(&jose.JSONWebKey{
		Key:       privateKey.Public(),
		Algorithm: "RSA",
	})
	if err != nil {
		return nil, err
	}

	return body, nil
}
