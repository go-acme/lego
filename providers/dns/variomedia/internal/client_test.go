package internal

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client := NewClient("secret")
			client.baseURL, _ = url.Parse(server.URL)
			client.HTTPClient = server.Client()

			return client, nil
		},
		servermock.CheckHeader().
			WithAccept("application/vnd.variomedia.v1+json").
			WithAuthorization("token secret"))
}

func TestClient_CreateDNSRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /dns-records",
			servermock.ResponseFromFixture("POST_dns-records.json"),
			servermock.CheckHeader().
				WithContentType("application/vnd.api+json"),
			servermock.CheckRequestJSONBody(`{"data":{"type":"dns-record","attributes":{"record_type":"TXT","name":"_acme-challenge","domain":"example.com","data":"test","ttl":300}}}`)).
		Build(t)

	record := DNSRecord{
		RecordType: "TXT",
		Name:       "_acme-challenge",
		Domain:     "example.com",
		Data:       "test",
		TTL:        300,
	}

	resp, err := client.CreateDNSRecord(t.Context(), record)
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
	client := mockBuilder().
		Route("DELETE /dns-records/test",
			servermock.ResponseFromFixture("DELETE_dns-records_pending.json")).
		Build(t)

	resp, err := client.DeleteDNSRecord(t.Context(), "test")
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
	client := mockBuilder().
		Route("GET /queue-jobs/test",
			servermock.ResponseFromFixture("GET_queue-jobs.json")).
		Build(t)

	resp, err := client.GetJob(t.Context(), "test")
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
