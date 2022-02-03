package internal

import (
	"context"
	"net/http"
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
			client := NewClient("key", "secret", server.Client())

			client.BaseURL, _ = url.Parse(server.URL)

			return client, nil
		},
		servermock.CheckHeader().
			WithJSONHeaders().
			WithRegexp("Authorization", `hmac key:.+:.+:\d+`),
	)
}

func TestClient_GetRecords(t *testing.T) {
	client := mockBuilder().
		Route("GET /dns/example.com/records",
			servermock.ResponseFromFixture("GetRecords.json")).
		Build(t)

	records, err := client.GetRecords(context.Background(), "example.com", nil)
	require.NoError(t, err)

	expected := []Record{
		{
			ID:         "string",
			Type:       "string",
			RecordName: "string",
			Content:    "string",
			TTL:        3600,
			Priority:   10,
			Service:    "string",
			Weight:     0,
			Target:     "string",
			Protocol:   "TCP",
			Port:       0,
		},
	}
	assert.Equal(t, expected, records)
}

func TestClient_GetRecord(t *testing.T) {
	client := mockBuilder().
		Route("GET /dns/example.com/records/123",
			servermock.ResponseFromFixture("GetRecord.json")).
		Build(t)

	record, err := client.GetRecord(context.Background(), "example.com", "123")
	require.NoError(t, err)

	expected := &Record{
		ID:         "string",
		Type:       "string",
		RecordName: "string",
		Content:    "string",
		TTL:        3600,
		Priority:   10,
		Service:    "string",
		Weight:     0,
		Target:     "string",
		Protocol:   "TCP",
		Port:       0,
	}
	assert.Equal(t, expected, record)
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /dns/example.com/records/123",
			servermock.Noop().
				WithStatusCode(http.StatusNoContent)).
		Build(t)

	err := client.DeleteRecord(context.Background(), "example.com", "123")
	require.NoError(t, err)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /dns/example.com/records/123",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized)).
		Build(t)

	err := client.DeleteRecord(context.Background(), "example.com", "123")
	require.Error(t, err)
}

func TestClient_sign(t *testing.T) {
	client := NewClient("my_key", "my_secret", nil)

	endpoint, err := url.Parse("https://localhost/v2/domains")
	require.NoError(t, err)

	query := endpoint.Query()
	query.Set("skip", "0")
	query.Set("take", "10")
	endpoint.RawQuery = query.Encode()

	req := httptest.NewRequest(http.MethodGet, endpoint.String(), nil)

	sign, err := client.sign(req, nil)
	require.NoError(t, err)

	assert.Regexp(t, `hmac my_key:[^:]+:[a-zA-Z]{10}:\d{10}`, sign)
}
