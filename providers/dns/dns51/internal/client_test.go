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
			client, err := NewClient("key", "secret")
			if err != nil {
				return nil, err
			}

			client.BaseURL, _ = url.Parse(server.URL)
			client.HTTPClient = server.Client()

			client.clock = func() time.Time {
				return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
			}

			return client, nil
		},
		servermock.CheckHeader().
			WithJSONHeaders(),
	)
}

func TestClient_CreateRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /record/create",
			servermock.ResponseFromFixture("create_record.json"),
			servermock.CheckRequestJSONBodyFromFixture("create_record-request.json"),
		).
		Build(t)

	record := RecordRequest{
		DomainID: 123,
		Type:     "TXT",
		ViewID:   456,
		Host:     "_acme-challenge",
		Value:    "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		TTL:      120,
		Remark:   "Created by go-acme/lego",
	}

	actual, err := client.CreateRecord(t.Context(), record)
	require.NoError(t, err)

	expected := &RecordData{
		RecordRequest: RecordRequest{
			DomainID: 123,
			Type:     "TXT",
			ViewID:   456,
			Value:    "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
			TTL:      120,
			Remark:   "Created by go-acme/lego",
		},
		RecordID: 1333,
		Record:   "_acme-challenge.example.com",
	}

	assert.Equal(t, expected, actual)
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /record/remove",
			servermock.ResponseFromFixture("delete_record.json"),
			servermock.CheckRequestJSONBodyFromFixture("delete_record-request.json"),
		).
		Build(t)

	err := client.DeleteRecord(t.Context(), 123, 456)
	require.NoError(t, err)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /record/remove",
			servermock.ResponseFromFixture("error.json"),
		).
		Build(t)

	err := client.DeleteRecord(t.Context(), 123, 456)
	require.EqualError(t, err, "3: 非法参数")
}

func TestClient_ListDomains(t *testing.T) {
	client := mockBuilder().
		Route("POST /domain/list",
			servermock.ResponseFromFixture("list_domains.json"),
			servermock.CheckRequestJSONBodyFromFixture("list_domains-request.json"),
		).
		Build(t)

	request := DomainRequest{
		GroupID:  123,
		Page:     1,
		PageSize: 10,
	}

	response, err := client.ListDomains(t.Context(), request)
	require.NoError(t, err)

	expected := &DomainData{
		Data: []Domain{
			{GroupID: 123, DomainID: 456, Domain: "example.com"},
			{GroupID: 123, DomainID: 789, Domain: "example.org"},
		},
		Page:      1,
		PageSize:  10,
		PageCount: 1,
	}

	assert.Equal(t, expected, response)
}
