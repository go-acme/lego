package internal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setup(t *testing.T, handler func(http.ResponseWriter, *http.Request)) *Client {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	mux.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(rw, fmt.Sprintf("unsupported method: %s", req.Method), http.StatusMethodNotAllowed)
			return
		}

		token := req.Header.Get("Authorization")
		if token != "Njalla secret" {
			_, _ = rw.Write([]byte(`{"jsonrpc":"2.0", "Error": {"code": 403, "message": "Invalid token."}}`))
			return
		}

		if handler != nil {
			handler(rw, req)
		} else {
			_, _ = rw.Write([]byte(`{"jsonrpc":"2.0"}`))
		}
	})

	client := NewClient("secret")
	client.apiEndpoint = server.URL

	return client
}

func TestClient_AddRecord(t *testing.T) {
	client := setup(t, func(rw http.ResponseWriter, req *http.Request) {
		apiReq := struct {
			Method string `json:"method"`
			Params Record `json:"params"`
		}{}

		err := json.NewDecoder(req.Body).Decode(&apiReq)
		if err != nil {
			http.Error(rw, "failed to marshal test request body", http.StatusInternalServerError)
			return
		}

		apiReq.Params.ID = "123"

		resp := map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      "897",
			"result":  apiReq.Params,
		}

		err = json.NewEncoder(rw).Encode(resp)
		if err != nil {
			http.Error(rw, "failed to marshal test response", http.StatusInternalServerError)
			return
		}
	})

	record := Record{
		Content: "foobar",
		Domain:  "test",
		Name:    "example.com",
		TTL:     300,
		Type:    "TXT",
	}

	result, err := client.AddRecord(record)
	require.NoError(t, err)

	expected := &Record{
		ID:      "123",
		Content: "foobar",
		Domain:  "test",
		Name:    "example.com",
		TTL:     300,
		Type:    "TXT",
	}
	assert.Equal(t, expected, result)
}

func TestClient_AddRecord_error(t *testing.T) {
	client := setup(t, nil)
	client.token = "invalid"

	record := Record{
		Content: "test",
		Domain:  "test01",
		Name:    "example.com",
		TTL:     300,
		Type:    "TXT",
	}

	result, err := client.AddRecord(record)
	require.Error(t, err)

	assert.Nil(t, result)
}

func TestClient_ListRecords(t *testing.T) {
	client := setup(t, func(rw http.ResponseWriter, req *http.Request) {
		apiReq := struct {
			Method string `json:"method"`
			Params Record `json:"params"`
		}{}

		err := json.NewDecoder(req.Body).Decode(&apiReq)
		if err != nil {
			http.Error(rw, "failed to marshal test request body", http.StatusInternalServerError)
			return
		}

		resp := map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      "897",
			"result": Records{
				Records: []Record{
					{
						ID:      "1",
						Domain:  apiReq.Params.Domain,
						Content: "test",
						Name:    "test01",
						TTL:     300,
						Type:    "TXT",
					},
					{
						ID:      "2",
						Domain:  apiReq.Params.Domain,
						Content: "txtTxt",
						Name:    "test02",
						TTL:     120,
						Type:    "TXT",
					},
				},
			},
		}

		err = json.NewEncoder(rw).Encode(resp)
		if err != nil {
			http.Error(rw, "failed to marshal test response", http.StatusInternalServerError)
			return
		}
	})

	records, err := client.ListRecords("example.com")
	require.NoError(t, err)

	expected := []Record{
		{
			ID:      "1",
			Domain:  "example.com",
			Content: "test",
			Name:    "test01",
			TTL:     300,
			Type:    "TXT",
		},
		{
			ID:      "2",
			Domain:  "example.com",
			Content: "txtTxt",
			Name:    "test02",
			TTL:     120,
			Type:    "TXT",
		},
	}

	assert.Equal(t, expected, records)
}

func TestClient_ListRecords_error(t *testing.T) {
	client := setup(t, nil)
	client.token = "invalid"

	records, err := client.ListRecords("example.com")
	require.Error(t, err)

	assert.Empty(t, records)
}

func TestClient_RemoveRecord(t *testing.T) {
	client := setup(t, func(rw http.ResponseWriter, req *http.Request) {
		apiReq := struct {
			Method string `json:"method"`
			Params Record `json:"params"`
		}{}

		err := json.NewDecoder(req.Body).Decode(&apiReq)
		if err != nil {
			http.Error(rw, "failed to marshal test request body", http.StatusInternalServerError)
			return
		}

		if apiReq.Params.ID == "" {
			_, _ = rw.Write([]byte(`{"jsonrpc":"2.0", "Error": {"code": 400, "message": ""missing ID"}}`))
			return
		}

		if apiReq.Params.Domain == "" {
			_, _ = rw.Write([]byte(`{"jsonrpc":"2.0", "Error": {"code": 400, "message": ""missing domain"}}`))
			return
		}

		_, _ = rw.Write([]byte(`{"jsonrpc":"2.0"}`))
	})

	err := client.RemoveRecord("123", "example.com")
	require.NoError(t, err)
}

func TestClient_RemoveRecord_error(t *testing.T) {
	client := setup(t, nil)
	client.token = "invalid"

	err := client.RemoveRecord("123", "example.com")
	require.Error(t, err)
}
