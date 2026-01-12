package internal

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder(
		func(server *httptest.Server) (*Client, error) {
			client, err := NewClient("secret")
			if err != nil {
				return nil, err
			}

			client.HTTPClient = server.Client()
			client.BaseURL, _ = url.Parse(server.URL)

			return client, nil
		},
		servermock.CheckHeader().
			WithJSONHeaders().
			With("apikey", "secret"),
	)
}

func TestClient_AddRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /",
			servermock.ResponseFromFixture("success.json"),
			servermock.CheckRequestJSONBodyFromFixture("add_record-request.json"),
		).
		Build(t)

	record := Record{
		Type:    "TXT",
		Name:    "_acme-challenge",
		Content: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		TTL:     2,
	}

	err := client.AddRecord(context.Background(), "example.com", record)
	require.NoError(t, err)
}

func TestClient_AddRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusBadRequest),
		).
		Build(t)

	record := Record{
		Type:      "TXT",
		Name:      "_acme-challenge",
		Content:   "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		TTL:       480,
		Overwrite: true,
	}

	err := client.AddRecord(context.Background(), "example.com", record)
	require.Error(t, err)

	require.EqualError(t, err, "JSON Defined Record Type not Supported (code=6)")
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /",
			servermock.ResponseFromFixture("success.json"),
			servermock.CheckRequestJSONBodyFromFixture("delete_record-request.json"),
		).
		Build(t)

	record := Record{
		Type:    "TXT",
		Name:    "_acme-challenge",
		Content: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
	}

	err := client.DeleteRecord(context.Background(), "example.com", record)
	require.NoError(t, err)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusBadRequest),
		).
		Build(t)

	record := Record{
		Type:    "TXT",
		Name:    "foo",
		Content: "txtTXTtxt",
	}

	err := client.DeleteRecord(context.Background(), "example.com", record)

	require.Error(t, err)

	require.EqualError(t, err, "JSON Defined Record Type not Supported (code=6)")
}
