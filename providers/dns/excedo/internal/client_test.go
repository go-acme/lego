package internal

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client, err := NewClient(server.URL, "secret")
			if err != nil {
				return nil, err
			}

			client.HTTPClient = server.Client()

			return client, nil
		},
	)
}

func TestClient_AddRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /dns/addrecord/",
			servermock.ResponseFromFixture("addrecord.json"),
			servermock.CheckHeader().
				WithAuthorization("Bearer session-token"),
			servermock.CheckForm().Strict().
				With("content", "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY").
				With("domainName", "example.com").
				With("name", "_acme-challenge").
				With("ttl", "60").
				With("type", "TXT"),
		).
		Build(t)

	client.token = &ExpirableToken{
		Token:   "session-token",
		Expires: time.Now().Add(6 * time.Hour),
	}

	record := Record{
		DomainName: "example.com",
		Name:       "_acme-challenge",
		Type:       "TXT",
		Content:    "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		TTL:        "60",
	}

	recordID, err := client.AddRecord(t.Context(), record)
	require.NoError(t, err)

	assert.EqualValues(t, 19695822, recordID)
}

func TestClient_AddRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /dns/addrecord/",
			servermock.ResponseFromFixture("error.json"),
		).
		Build(t)

	client.token = &ExpirableToken{
		Token:   "session-token",
		Expires: time.Now().Add(6 * time.Hour),
	}

	record := Record{
		DomainName: "example.com",
		Name:       "_acme-challenge",
		Type:       "TXT",
		Content:    "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		TTL:        "60",
	}

	_, err := client.AddRecord(t.Context(), record)
	require.EqualError(t, err, "2003: Required parameter missing")
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /dns/deleterecord/",
			servermock.ResponseFromFixture("deleterecord.json"),
			servermock.CheckHeader().
				WithAuthorization("Bearer session-token"),
		).
		Build(t)

	client.token = &ExpirableToken{
		Token:   "session-token",
		Expires: time.Now().Add(6 * time.Hour),
	}

	err := client.DeleteRecord(t.Context(), "example.com", "19695822")
	require.NoError(t, err)
}

func TestClient_GetRecords(t *testing.T) {
	client := mockBuilder().
		Route("GET /dns/getrecords/",
			servermock.ResponseFromFixture("getrecords.json"),
			servermock.CheckHeader().
				WithAuthorization("Bearer session-token"),
			servermock.CheckQueryParameter().Strict().
				With("domainname", "example.com"),
		).
		Build(t)

	client.token = &ExpirableToken{
		Token:   "session-token",
		Expires: time.Now().Add(6 * time.Hour),
	}

	zones, err := client.GetRecords(t.Context(), "example.com")
	require.NoError(t, err)

	expected := map[string]Zone{
		"example.com": {
			DNSType: "type",
			Records: []Record{{
				RecordID: "1234",
				Name:     "_acme-challenge.example.com",
				Type:     "TXT",
				Content:  "txt-value",
				TTL:      "60",
			}},
		},
	}

	assert.Equal(t, expected, zones)
}
