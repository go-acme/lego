package internal

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder(
		func(server *httptest.Server) (*Client, error) {
			client, err := NewClient(OAuthStaticAccessToken(server.Client(), "secret"))
			if err != nil {
				return nil, err
			}

			client.baseURL, _ = url.Parse(server.URL)

			return client, nil
		},
		servermock.CheckHeader().
			WithJSONHeaders().
			With("Authorization", "Bearer secret"),
	)
}

func TestClient_RetrieveRecords(t *testing.T) {
	client := mockBuilder().
		Route("GET /address/foobar/dns",
			servermock.ResponseFromFixture("retrieve-records.json")).
		Build(t)

	records, err := client.RetrieveRecords(t.Context(), "foobar")
	require.NoError(t, err)

	expected := []Record{
		{
			ID:        "2857074",
			Name:      "foobar",
			Content:   "",
			TTL:       3600,
			Priority:  0,
			Type:      "A",
			CreatedAt: time.Date(2022, time.November, 26, 4, 30, 13, 0, time.UTC),
			UpdatedAt: time.Date(2022, time.November, 26, 4, 31, 33, 0, time.UTC),
		},
		{
			ID:        "2857075",
			Name:      "cname.foobar",
			Content:   "",
			TTL:       3600,
			Priority:  0,
			Type:      "CNAME",
			CreatedAt: time.Date(2022, time.November, 26, 4, 34, 24, 0, time.UTC),
			UpdatedAt: time.Date(2022, time.November, 26, 4, 34, 24, 0, time.UTC),
		},
	}

	assert.Equal(t, expected, records)
}

func TestClient_RetrieveRecords_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /address/foobar/dns",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized)).
		Build(t)

	_, err := client.RetrieveRecords(t.Context(), "foobar")
	require.Error(t, err)

	require.EqualError(t, err, "You need to authenticate with an API key. https://api.omg.lol")
}

func TestClient_CreateRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /address/foobar/dns",
			servermock.ResponseFromFixture("create-record.json"),
			servermock.CheckRequestJSONBodyFromFixture("create-record-request.json")).
		Build(t)

	record := Record{
		Name:    "foobar",
		Content: "txt",
		TTL:     60,
		Type:    "TXT",
	}

	newRecord, err := client.CreateRecord(t.Context(), "foobar", record)
	require.NoError(t, err)

	expected := &Record{
		ID:        "2857074",
		Name:      "foobar",
		Content:   "10.0.0.1",
		TTL:       3600,
		Priority:  0,
		Type:      "A",
		CreatedAt: time.Date(2022, time.November, 26, 4, 30, 13, 0, time.UTC),
		UpdatedAt: time.Date(2022, time.November, 26, 4, 30, 13, 0, time.UTC),
	}

	assert.Equal(t, expected, newRecord)
}

func TestClient_CreateRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /address/foobar/dns",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized)).
		Build(t)

	record := Record{
		Name:    "foobar",
		Content: "txt",
		TTL:     60,
		Type:    "TXT",
	}

	_, err := client.CreateRecord(t.Context(), "foobar", record)
	require.Error(t, err)

	require.EqualError(t, err, "You need to authenticate with an API key. https://api.omg.lol")
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /address/foobar/dns/123",
			servermock.ResponseFromFixture("delete-record.json")).
		Build(t)

	err := client.DeleteRecord(t.Context(), "foobar", "123")
	require.NoError(t, err)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /address/foobar/dns/123",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized)).
		Build(t)

	err := client.DeleteRecord(t.Context(), "foobar", "123")
	require.Error(t, err)

	require.EqualError(t, err, "You need to authenticate with an API key. https://api.omg.lol")
}
