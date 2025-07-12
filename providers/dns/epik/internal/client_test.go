package internal

import (
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
			client := NewClient("secret")
			client.HTTPClient = server.Client()
			client.baseURL, _ = url.Parse(server.URL)

			return client, nil
		},
		servermock.CheckHeader().WithJSONHeaders(),
	)
}

func TestClient_GetDNSRecords(t *testing.T) {
	client := mockBuilder().
		Route("GET /domains/example.com/records",
			servermock.ResponseFromFixture("getDnsRecord.json"),
			servermock.CheckQueryParameter().Strict().
				With("SIGNATURE", "secret")).
		Build(t)

	records, err := client.GetDNSRecords(t.Context(), "example.com")
	require.NoError(t, err)

	expected := []Record{
		{
			ID:   "abc123",
			Name: "www",
			Type: "CAA",
			Data: "1 issue letsencrypt.org",
			AUX:  0,
			TTL:  300,
		},
		{
			ID:   "abc123",
			Name: "www",
			Type: "A",
			Data: "192.64.147.249",
			AUX:  0,
			TTL:  300,
		},
		{
			ID:   "abc123",
			Name: "*",
			Type: "A",
			Data: "192.64.147.249",
			AUX:  0,
			TTL:  300,
		},
		{
			ID:   "abc123",
			Type: "CAA",
			Data: "0 issue trust-provider.com",
			AUX:  0,
			TTL:  300,
		},
		{
			ID:   "abc123",
			Type: "CAA",
			Data: "1 issue letsencrypt.org",
			AUX:  0,
			TTL:  300,
		},
		{
			ID:   "abc123",
			Type: "A",
			Data: "192.64.147.249",
			AUX:  0,
			TTL:  300,
		},
	}

	assert.Equal(t, expected, records)
}

func TestClient_GetDNSRecords_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /domains/example.com/records",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized),
			servermock.CheckQueryParameter().Strict().
				With("SIGNATURE", "secret")).
		Build(t)

	_, err := client.GetDNSRecords(t.Context(), "example.com")
	require.Error(t, err)
}

func TestClient_CreateHostRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /domains/example.com/records",
			servermock.ResponseFromFixture("createHostRecord.json"),
			servermock.CheckQueryParameter().Strict().
				With("SIGNATURE", "secret")).
		Build(t)

	record := RecordRequest{
		Host: "www2",
		Type: "A",
		Data: "192.64.147.249",
		Aux:  0,
		TTL:  300,
	}

	data, err := client.CreateHostRecord(t.Context(), "example.com", record)
	require.NoError(t, err)

	expected := &Data{
		Code:    1000,
		Message: "Command completed successfully.",
	}

	assert.Equal(t, expected, data)
}

func TestClient_CreateHostRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /domains/example.com/records",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized),
			servermock.CheckQueryParameter().Strict().
				With("SIGNATURE", "secret")).
		Build(t)

	record := RecordRequest{
		Host: "www2",
		Type: "A",
		Data: "192.64.147.249",
		Aux:  0,
		TTL:  300,
	}

	_, err := client.CreateHostRecord(t.Context(), "example.com", record)
	require.Error(t, err)
}

func TestClient_RemoveHostRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /domains/example.com/records",
			servermock.ResponseFromFixture("removeHostRecord.json"),
			servermock.CheckQueryParameter().Strict().
				With("ID", "abc123").
				With("SIGNATURE", "secret")).
		Build(t)

	data, err := client.RemoveHostRecord(t.Context(), "example.com", "abc123")
	require.NoError(t, err)

	expected := &Data{
		Code:    1000,
		Message: "Command completed successfully.",
	}

	assert.Equal(t, expected, data)
}

func TestClient_RemoveHostRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /domains/example.com/records",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized)).
		Build(t)

	_, err := client.RemoveHostRecord(t.Context(), "example.com", "abc123")
	require.Error(t, err)
}
