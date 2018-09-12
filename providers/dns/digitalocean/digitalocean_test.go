package digitalocean

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var fakeDigitalOceanAuth = "asdf1234"

func TestDigitalOceanPresent(t *testing.T) {
	var requestReceived bool

	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestReceived = true

		assert.Equal(t, http.MethodPost, r.Method, "method")
		assert.Equal(t, "/v2/domains/example.com/records", r.URL.Path, "Path")
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"), "Content-Type")
		assert.Equal(t, "Bearer asdf1234", r.Header.Get("Authorization"), "Authorization")

		reqBody, err := ioutil.ReadAll(r.Body)
		require.NoError(t, err, "reading request body")
		assert.Equal(t, `{"type":"TXT","name":"_acme-challenge.example.com.","data":"w6uP8Tcg6K2QR905Rms8iXTlksL6OD1KOWBxTK7wxPI","ttl":30}`, string(reqBody))

		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, `{
			"domain_record": {
				"id": 1234567,
				"type": "TXT",
				"name": "_acme-challenge",
				"data": "w6uP8Tcg6K2QR905Rms8iXTlksL6OD1KOWBxTK7wxPI",
				"priority": null,
				"port": null,
				"weight": null
			}
		}`)
	}))
	defer mock.Close()

	config := NewDefaultConfig()
	config.AuthToken = fakeDigitalOceanAuth
	config.BaseURL = mock.URL

	provider, err := NewDNSProviderConfig(config)
	require.NoError(t, err)
	require.NotNil(t, provider)

	err = provider.Present("example.com", "", "foobar")
	require.NoError(t, err, "fail to create TXT record")

	assert.True(t, requestReceived, "Expected request to be received by mock backend, but it wasn't")
}

func TestDigitalOceanCleanUp(t *testing.T) {
	var requestReceived bool

	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestReceived = true

		assert.Equal(t, http.MethodDelete, r.Method, "method")
		assert.Equal(t, "/v2/domains/example.com/records/1234567", r.URL.Path, "Path")
		// NOTE: Even though the body is empty, DigitalOcean API docs still show setting this Content-Type...
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"), "Content-Type")
		assert.Equal(t, "Bearer asdf1234", r.Header.Get("Authorization"), "Authorization")

		w.WriteHeader(http.StatusNoContent)
	}))
	defer mock.Close()

	config := NewDefaultConfig()
	config.AuthToken = fakeDigitalOceanAuth
	config.BaseURL = mock.URL

	provider, err := NewDNSProviderConfig(config)
	require.NoError(t, err)
	require.NotNil(t, provider)

	provider.recordIDsMu.Lock()
	provider.recordIDs["_acme-challenge.example.com."] = 1234567
	provider.recordIDsMu.Unlock()

	err = provider.CleanUp("example.com", "", "")
	require.NoError(t, err, "fail to remove TXT record")

	assert.True(t, requestReceived, "Expected request to be received by mock backend, but it wasn't")
}
