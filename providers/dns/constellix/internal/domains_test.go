package internal

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupAPIMock() (*Client, *http.ServeMux, func()) {
	handler := http.NewServeMux()
	svr := httptest.NewServer(handler)

	client := NewClient(nil)
	client.BaseURL = svr.URL

	return client, handler, svr.Close
}

func TestDomainService_GetAll(t *testing.T) {
	client, handler, tearDown := setupAPIMock()
	defer tearDown()

	handler.HandleFunc("/v1/domains", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(rw, "invalid method: "+req.Method, http.StatusBadRequest)
			return
		}

		file, err := os.Open("./fixtures/domains-GetAll.json")
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		defer func() { _ = file.Close() }()

		_, err = io.Copy(rw, file)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	data, err := client.Domains.GetAll(nil)
	require.NoError(t, err)

	expected := []Domain{
		{ID: 273301, Name: "aaa.wtf", TypeID: 1, Version: 9, Status: "ACTIVE"},
		{ID: 273302, Name: "bbb.wtf", TypeID: 1, Version: 9, Status: "ACTIVE"},
		{ID: 273303, Name: "ccc.wtf", TypeID: 1, Version: 9, Status: "ACTIVE"},
		{ID: 273304, Name: "ddd.wtf", TypeID: 1, Version: 9, Status: "ACTIVE"},
	}

	assert.Equal(t, expected, data)
}

func TestDomainService_Search(t *testing.T) {
	client, handler, tearDown := setupAPIMock()
	defer tearDown()

	handler.HandleFunc("/v1/domains/search", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(rw, "invalid method: "+req.Method, http.StatusBadRequest)
			return
		}

		file, err := os.Open("./fixtures/domains-Search.json")
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		defer func() { _ = file.Close() }()

		_, err = io.Copy(rw, file)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	data, err := client.Domains.Search(Exact, "lego.wtf")
	require.NoError(t, err)

	expected := []Domain{
		{ID: 273302, Name: "lego.wtf", TypeID: 1, Version: 9, Status: "ACTIVE"},
	}

	assert.Equal(t, expected, data)
}
