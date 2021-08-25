package internal

import (
	"encoding/json"
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

	client := NewClient("", "")
	client.BaseURL = svr.URL

	return client, handler, svr.Close
}

func TestClient_GetRecords(t *testing.T) {
	client, handler, tearDown := setupAPIMock()
	defer tearDown()

	handler.HandleFunc("/lego.wtf", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(rw, "invalid method: "+req.Method, http.StatusBadRequest)
			return
		}

		file, err := os.Open("./fixtures/records-01.json")
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

	records, err := client.GetRecords("lego.wtf")
	require.NoError(t, err)

	recordsJSON, err := json.Marshal(records)
	require.NoError(t, err)

	expectedContent, err := os.ReadFile("./fixtures/records-01.json")
	require.NoError(t, err)

	assert.JSONEq(t, string(expectedContent), string(recordsJSON))
}

func TestClient_GetRecords_error(t *testing.T) {
	client, handler, tearDown := setupAPIMock()
	defer tearDown()

	handler.HandleFunc("/lego.wtf", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(rw, "invalid method: "+req.Method, http.StatusBadRequest)
			return
		}

		err := json.NewEncoder(rw).Encode(Message{ErrorMsg: "authentication failed"})
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	records, err := client.GetRecords("lego.wtf")
	require.Error(t, err)

	assert.Nil(t, records)
}

func TestClient_CreateUpdateRecord(t *testing.T) {
	client, handler, tearDown := setupAPIMock()
	defer tearDown()

	handler.HandleFunc("/lego.wtf", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(rw, "invalid method: "+req.Method, http.StatusBadRequest)
			return
		}

		content, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		expectedRequest := `{"name":"_acme-challenge.www","type":"TXT","ttl":30,"content":["aaa","bbb"]}`

		if !assert.JSONEq(t, expectedRequest, string(content)) {
			http.Error(rw, "invalid content", http.StatusBadRequest)
			return
		}

		err = json.NewEncoder(rw).Encode(Message{Message: "ok"})
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	record := Record{
		Name:    "_acme-challenge.www",
		Type:    "TXT",
		TTL:     30,
		Content: Value{"aaa", "bbb"},
	}

	msg, err := client.CreateUpdateRecord("lego.wtf", record)
	require.NoError(t, err)

	expected := &Message{Message: "ok"}
	assert.Equal(t, expected, msg)
}

func TestClient_CreateUpdateRecord_error(t *testing.T) {
	client, handler, tearDown := setupAPIMock()
	defer tearDown()

	handler.HandleFunc("/lego.wtf", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(rw, "invalid method: "+req.Method, http.StatusBadRequest)
			return
		}

		err := json.NewEncoder(rw).Encode(Message{ErrorMsg: "parameter type must be cname, txt, tlsa, caa, a or aaaa"})
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	record := Record{
		Name: "_acme-challenge.www",
	}

	msg, err := client.CreateUpdateRecord("lego.wtf", record)
	require.Error(t, err)

	assert.Nil(t, msg)
}

func TestClient_DeleteRecord(t *testing.T) {
	client, handler, tearDown := setupAPIMock()
	defer tearDown()

	handler.HandleFunc("/lego.wtf", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodDelete {
			http.Error(rw, "invalid method: "+req.Method, http.StatusBadRequest)
			return
		}

		content, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		expectedRequest := `{"name":"_acme-challenge.www","type":"TXT"}`

		if !assert.JSONEq(t, expectedRequest, string(content)) {
			http.Error(rw, "invalid content", http.StatusBadRequest)
			return
		}

		err = json.NewEncoder(rw).Encode(Message{Message: "ok"})
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	record := Record{
		Name: "_acme-challenge.www",
		Type: "TXT",
	}

	msg, err := client.DeleteRecord("lego.wtf", record)
	require.NoError(t, err)

	expected := &Message{Message: "ok"}
	assert.Equal(t, expected, msg)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client, handler, tearDown := setupAPIMock()
	defer tearDown()

	handler.HandleFunc("/lego.wtf", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodDelete {
			http.Error(rw, "invalid method: "+req.Method, http.StatusBadRequest)
			return
		}

		err := json.NewEncoder(rw).Encode(Message{ErrorMsg: "parameter type must be cname, txt, tlsa, caa, a or aaaa"})
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	record := Record{
		Name: "_acme-challenge.www",
	}

	msg, err := client.DeleteRecord("lego.wtf", record)
	require.Error(t, err)

	assert.Nil(t, msg)
}
