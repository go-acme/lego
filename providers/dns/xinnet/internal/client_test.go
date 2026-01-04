package internal

import (
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			signer, err := NewSigner("9f3b6760952b41eda7ddad84d755b3944cbb7929277947c294b7a578128e1170", "agent12345")
			if err != nil {
				return nil, err
			}

			signer.clock = func() time.Time {
				return time.Date(2024, 6, 5, 8, 55, 45, 0, time.UTC)
			}

			client, err := NewClient(signer)
			if err != nil {
				return nil, err
			}

			client.BaseURL, _ = url.Parse(server.URL)
			client.HTTPClient = server.Client()

			return client, nil
		},
		servermock.CheckHeader().
			WithJSONHeaders(),
	)
}

func TestClient_CreateRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /api/dns/create/",
			servermock.ResponseFromFixture("create_record.json"),
			servermock.CheckRequestJSONBodyFromFixture("create_record-request.json"),
			servermock.CheckHeader().
				With("Timestamp", "20240605T085545Z").
				WithAuthorization("HMAC-SHA256 Access=agent12345, Signature=c72bfefb4841fc6fd4611ec880541c0281f07b157488cf8945a4857aa96cf4ab"),
		).
		Build(t)

	record := Record{
		DomainName: "example.com",
		RecordName: "_acme-challenge.example.com",
		Type:       "TXT",
		Value:      "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		Line:       "默认",
		TTL:        120,
	}

	recordID, err := client.CreateRecord(t.Context(), record)
	require.NoError(t, err)

	assert.EqualValues(t, 165812154, recordID)
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /api/dns/delete/",
			servermock.ResponseFromFixture("delete_record.json"),
			servermock.CheckRequestJSONBodyFromFixture("delete_record-request.json"),
			servermock.CheckHeader().
				With("Timestamp", "20240605T085545Z").
				WithAuthorization("HMAC-SHA256 Access=agent12345, Signature=892ce697628f9042ef43d94021ab68b936e9ef6e53eaa804fd8194e592c8af96"),
		).
		Build(t)

	err := client.DeleteRecord(t.Context(), "example.com", 165812154)
	require.NoError(t, err)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /api/dns/delete/",
			servermock.ResponseFromFixture("error.json")).
		Build(t)

	err := client.DeleteRecord(t.Context(), "example.com", 165812154)
	require.EqualError(t, err, "1: 已存在相同的解析记录 (pvl1VyyUQur3JB2vPfK)")
}
