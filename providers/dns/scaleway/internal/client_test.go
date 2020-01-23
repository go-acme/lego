package internal

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const fakeToken = "test"

func setupTest() (*Client, *http.ServeMux, func()) {
	mux := http.NewServeMux()
	svr := httptest.NewServer(mux)

	opts := ClientOpts{
		BaseURL: svr.URL,
		Token:   fakeToken,
	}
	client := NewClient(opts, nil)

	return client, mux, func() {
		svr.Close()
	}
}

func TestClient_AddRecord(t *testing.T) {
	client, mux, tearDown := setupTest()
	defer tearDown()

	mux.HandleFunc("/dns-zones/zone/records", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPatch {
			http.Error(rw, fmt.Sprintf("unsupported method: %s", req.Method), http.StatusBadRequest)
			return
		}

		auth := req.Header.Get("X-Auth-Token")
		if auth != fakeToken {
			http.Error(rw, fmt.Sprintf("invalid token: %s", auth), http.StatusUnauthorized)
			return
		}

		raw, err := ioutil.ReadAll(req.Body)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		expected := `{"dns_zone":"zone","changes":[{"add":{"records":[{"data":"\"value\"","name":"fqdn","ttl":30,"type":"TXT"}]}}]}`
		assert.Equal(t, expected+"\n", string(raw))
	})

	record := Record{
		Type: "TXT",
		TTL:  30,
		Name: "fqdn",
		Data: fmt.Sprintf(`"%s"`, "value"),
	}

	err := client.AddRecord("zone", record)
	require.NoError(t, err)
}

func TestClient_AddRecord_error(t *testing.T) {
	client, mux, tearDown := setupTest()
	defer tearDown()

	mux.HandleFunc("/dns-zones/zone/records", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPatch {
			http.Error(rw, fmt.Sprintf("unsupported method: %s", req.Method), http.StatusBadRequest)
			return
		}

		auth := req.Header.Get("X-Auth-Token")
		if auth != fakeToken {
			http.Error(rw, fmt.Sprintf("invalid token: %s", auth), http.StatusUnauthorized)
			return
		}

		rw.WriteHeader(http.StatusNotFound)
		err := json.NewEncoder(rw).Encode(APIError{"oops"})
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	record := Record{
		Type: "TXT",
		TTL:  30,
		Name: "fqdn",
		Data: fmt.Sprintf(`"%s"`, "value"),
	}

	err := client.AddRecord("zone", record)
	require.EqualError(t, err, "request failed with status code 404: oops")
}

func TestClient_SetRecord(t *testing.T) {
	client, mux, tearDown := setupTest()
	defer tearDown()

	mux.HandleFunc("/dns-zones/zone/records", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPatch {
			http.Error(rw, fmt.Sprintf("unsupported method: %s", req.Method), http.StatusBadRequest)
			return
		}

		auth := req.Header.Get("X-Auth-Token")
		if auth != fakeToken {
			http.Error(rw, fmt.Sprintf("invalid token: %s", auth), http.StatusUnauthorized)
			return
		}

		raw, err := ioutil.ReadAll(req.Body)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		expected := `{"dns_zone":"zone","changes":[{"set":{"name":"fqdn","type":"TXT","records":[{"data":"\"value\"","name":"fqdn","ttl":30,"type":"TXT"}]}}]}`
		assert.Equal(t, expected+"\n", string(raw))
	})

	record := Record{
		Type: "TXT",
		TTL:  30,
		Name: "fqdn",
		Data: fmt.Sprintf(`"%s"`, "value"),
	}

	err := client.SetRecord("zone", record)
	require.NoError(t, err)
}

func TestClient_DeleteRecord(t *testing.T) {
	client, mux, tearDown := setupTest()
	defer tearDown()

	mux.HandleFunc("/dns-zones/zone/records", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPatch {
			http.Error(rw, fmt.Sprintf("unsupported method: %s", req.Method), http.StatusBadRequest)
			return
		}

		auth := req.Header.Get("X-Auth-Token")
		if auth != fakeToken {
			http.Error(rw, fmt.Sprintf("invalid token: %s", auth), http.StatusUnauthorized)
			return
		}

		raw, err := ioutil.ReadAll(req.Body)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		expected := `{"dns_zone":"zone","changes":[{"delete":{"data":"\"value\"","name":"fqdn","type":"TXT"}}]}`
		assert.Equal(t, expected+"\n", string(raw))
	})

	record := Record{
		Type: "TXT",
		TTL:  30,
		Name: "fqdn",
		Data: fmt.Sprintf(`"%s"`, "value"),
	}

	err := client.DeleteRecord("zone", record)
	require.NoError(t, err)
}
