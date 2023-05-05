package internal

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T) (*Client, *http.ServeMux) {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := NewClient("secret")
	client.baseURL, _ = url.Parse(server.URL)

	return client, mux
}

func mockHandler(method string, filename string) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != method {
			http.Error(rw, fmt.Sprintf("invalid method, got %s want %s", req.Method, method), http.StatusBadRequest)
			return
		}

		filename = "./fixtures/" + filename
		statusCode := http.StatusOK

		if req.Header.Get(authorizationHeader) != "token secret" {
			statusCode = http.StatusUnauthorized
			filename = "./fixtures/error.json"
		}

		rw.WriteHeader(statusCode)

		file, err := os.Open(filename)
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
	}
}

func TestClient_CreateDNSRecord(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/dns-records", mockHandler(http.MethodPost, "POST_dns-records.json"))

	record := DNSRecord{
		RecordType: "TXT",
		Name:       "_acme-challenge",
		Domain:     "example.com",
		Data:       "test",
		TTL:        300,
	}

	resp, err := client.CreateDNSRecord(context.Background(), record)
	require.NoError(t, err)

	expected := &CreateDNSRecordResponse{
		Data: struct {
			Type       string `json:"type"`
			ID         string `json:"id"`
			Attributes struct {
				Status string `json:"status"`
			} `json:"attributes"`
			Links struct {
				QueueJob  string `json:"queue-job"`
				DNSRecord string `json:"dns-record"`
			} `json:"links"`
		}{
			Type: "queue-job",
			ID:   "18181818",
			Attributes: struct {
				Status string `json:"status"`
			}{
				Status: "pending",
			},
			Links: struct {
				QueueJob  string `json:"queue-job"`
				DNSRecord string `json:"dns-record"`
			}{
				QueueJob:  "https://api.variomedia.de/queue-jobs/18181818",
				DNSRecord: "https://api.variomedia.de/dns-records/19191919",
			},
		},
	}

	assert.Equal(t, expected, resp)
}

func TestClient_DeleteDNSRecord(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/dns-records/test", mockHandler(http.MethodDelete, "DELETE_dns-records_pending.json"))

	resp, err := client.DeleteDNSRecord(context.Background(), "test")
	require.NoError(t, err)

	expected := &DeleteRecordResponse{
		Data: struct {
			ID         string `json:"id"`
			Type       string `json:"type"`
			Attributes struct {
				JobType string `json:"job_type"`
				Status  string `json:"status"`
			} `json:"attributes"`
			Links struct {
				Self   string `json:"self"`
				Object string `json:"object"`
			} `json:"links"`
		}{
			ID:   "303030",
			Type: "queue-job",
			Attributes: struct {
				JobType string `json:"job_type"`
				Status  string `json:"status"`
			}{
				Status: "pending",
			},
		},
	}

	assert.Equal(t, expected, resp)
}

func TestClient_GetJob(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/queue-jobs/test", mockHandler(http.MethodGet, "GET_queue-jobs.json"))

	resp, err := client.GetJob(context.Background(), "test")
	require.NoError(t, err)

	expected := &GetJobResponse{
		Data: struct {
			ID         string `json:"id"`
			Type       string `json:"type"`
			Attributes struct {
				JobType string `json:"job_type"`
				Status  string `json:"status"`
			} `json:"attributes"`
			Links struct {
				Self   string `json:"self"`
				Object string `json:"object"`
			} `json:"links"`
		}{
			ID:   "171717",
			Type: "queue-job",
			Attributes: struct {
				JobType string `json:"job_type"`
				Status  string `json:"status"`
			}{
				JobType: "dns-record",
				Status:  "done",
			},
			Links: struct {
				Self   string `json:"self"`
				Object string `json:"object"`
			}{
				Self:   "https://api.variomedia.de/queue-jobs/171717",
				Object: "https://api.variomedia.de/dns-records/212121",
			},
		},
	}

	assert.Equal(t, expected, resp)
}
