package auroradns

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var fakeAuroraDNSUserID = "asdf1234"
var fakeAuroraDNSKey = "key"

func TestAuroraDNSPresent(t *testing.T) {
	var requestReceived bool

	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/zones" {
			w.WriteHeader(http.StatusCreated)
			fmt.Fprintf(w, `[{
			        "id":   "c56a4180-65aa-42ec-a945-5fd21dec0538",
			        "name": "example.com"
			      }]`)
			return
		}

		requestReceived = true

		assert.Equal(t, http.MethodPost, r.Method, "method")
		assert.Equal(t, "/zones/c56a4180-65aa-42ec-a945-5fd21dec0538/records", r.URL.Path, "Path")
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"), "Content-Type")

		reqBody, err := ioutil.ReadAll(r.Body)
		require.NoError(t, err, "reading request body")
		assert.Equal(t, `{"type":"TXT","name":"_acme-challenge","content":"w6uP8Tcg6K2QR905Rms8iXTlksL6OD1KOWBxTK7wxPI","ttl":300}`, string(reqBody))

		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, `{
		      "id":   "c56a4180-65aa-42ec-a945-5fd21dec0538",
		      "type": "TXT",
		      "name": "_acme-challenge",
		      "ttl":  300
		    }`)
	}))

	defer mock.Close()

	config := NewDefaultConfig()
	config.UserID = fakeAuroraDNSUserID
	config.Key = fakeAuroraDNSKey
	config.BaseURL = mock.URL

	provider, err := NewDNSProviderConfig(config)
	require.NoError(t, err)
	require.NotNil(t, provider)

	err = provider.Present("example.com", "", "foobar")
	require.NoError(t, err, "fail to create TXT record")

	assert.True(t, requestReceived, "Expected request to be received by mock backend, but it wasn't")
}

func TestAuroraDNSCleanUp(t *testing.T) {
	var requestReceived bool

	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/zones" {
			w.WriteHeader(http.StatusCreated)
			fmt.Fprintf(w, `[{
			        "id":   "c56a4180-65aa-42ec-a945-5fd21dec0538",
			        "name": "example.com"
			      }]`)
			return
		}

		if r.Method == http.MethodPost && r.URL.Path == "/zones/c56a4180-65aa-42ec-a945-5fd21dec0538/records" {
			w.WriteHeader(http.StatusCreated)
			fmt.Fprintf(w, `{
			        "id":   "ec56a4180-65aa-42ec-a945-5fd21dec0538",
			        "type": "TXT",
			        "name": "_acme-challenge",
			        "ttl":  300
			      }`)
			return
		}

		requestReceived = true

		assert.Equal(t, http.MethodDelete, r.Method, "method")
		assert.Equal(t, "/zones/c56a4180-65aa-42ec-a945-5fd21dec0538/records/ec56a4180-65aa-42ec-a945-5fd21dec0538", r.URL.Path, "Path")
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"), "Content-Type")

		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, `{}`)
	}))
	defer mock.Close()

	config := NewDefaultConfig()
	config.UserID = fakeAuroraDNSUserID
	config.Key = fakeAuroraDNSKey
	config.BaseURL = mock.URL

	provider, err := NewDNSProviderConfig(config)
	require.NoError(t, err)
	require.NotNil(t, provider)

	err = provider.Present("example.com", "", "foobar")
	require.NoError(t, err, "fail to create TXT record")

	err = provider.CleanUp("example.com", "", "foobar")
	require.NoError(t, err, "fail to remove TXT record")

	assert.True(t, requestReceived, "Expected request to be received by mock backend, but it wasn't")
}
