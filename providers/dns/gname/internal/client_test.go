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
			client, err := NewClient("app123", "secret")
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
			WithContentTypeFromURLEncoded(),
	)
}

func TestClient_AddRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /api/resolution/add",
			servermock.ResponseFromFixture("add_record.json"),
			servermock.CheckForm().Strict().
				With("ym", "example.com").
				With("zj", "_acme-challenge").
				With("lx", "TXT").
				With("ttl", "120").
				With("jlz", "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY").
				With("gntoken", "2A43EE8272F68DFC834DE8311B288C8F").
				With("gntime", "1767225600").
				With("appid", "app123"),
		).
		Build(t)

	record := Record{
		Domain:      "example.com",
		RecordType:  "TXT",
		HostRecord:  "_acme-challenge",
		RecordValue: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		TTL:         120,
	}

	recordID, err := client.AddRecord(t.Context(), record)
	require.NoError(t, err)

	assert.Equal(t, 1277, recordID)
}

func TestClient_AddRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /api/resolution/add",
			servermock.ResponseFromFixture("error.json"),
		).
		Build(t)

	record := Record{
		Domain:      "example.com",
		RecordType:  "TXT",
		HostRecord:  "_acme-challenge",
		RecordValue: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		TTL:         120,
	}

	_, err := client.AddRecord(t.Context(), record)
	require.EqualError(t, err, "-1: Sorry,APPID  is invalid")
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /api/resolution/delete",
			servermock.ResponseFromFixture("delete_record.json"),
			servermock.CheckForm().Strict().
				With("ym", "example.com").
				With("jxid", "123").
				With("gntoken", "42C4572FC0AAB891EFE101CB675F31AE").
				With("gntime", "1767225600").
				With("appid", "app123"),
		).
		Build(t)

	err := client.DeleteRecord(t.Context(), "example.com", 123)
	require.NoError(t, err)
}
