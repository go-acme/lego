package internal

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockBuilder(apiKey, pat string) *servermock.Builder[*Client] {
	checkHeaders := servermock.CheckHeader().WithJSONHeaders()

	if apiKey != "" {
		checkHeaders = checkHeaders.WithAuthorization("Apikey secret-apikey")
	} else {
		checkHeaders = checkHeaders.WithAuthorization("Bearer secret-pat")
	}

	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client := NewClient(apiKey, pat)
			client.BaseURL, _ = url.Parse(server.URL)
			client.HTTPClient = server.Client()

			return client, nil
		},
		checkHeaders,
	)
}

func TestClient_AddVerificationRecord(t *testing.T) {
	client := mockBuilder("secret-apikey", "").
		Route("POST /domains/example.com/verification/_acme-challenge/TXT",
			servermock.ResponseFromFixture("post_txt_verification_record.json"),
			servermock.CheckRequestJSONBodyFromFixture("post_txt_verification_record-request.json"),
		).
		Build(t)

	result, err := client.AddVerificationRecord(t.Context(), "example.com", "_acme-challenge", "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY")
	require.NoError(t, err)

	expected := &Record{
		TTL:    300,
		Values: []string{"value1", "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY"},
		Name:   "_acme-challenge",
		Type:   "TXT",
	}

	assert.Equal(t, expected, result)
}

func TestClient_GetTXTRecord(t *testing.T) {
	client := mockBuilder("secret-apikey", "").
		Route("GET /domains/example.com/records/_acme-challenge/TXT",
			servermock.ResponseFromFixture("get_txt_record.json"),
		).
		Build(t)

	result, err := client.GetTXTRecord(t.Context(), "example.com", "_acme-challenge")
	require.NoError(t, err)

	expected := &Record{
		TTL:    300,
		Values: []string{"value1"},
		Name:   "_acme-challenge",
		Type:   "TXT",
	}

	assert.Equal(t, expected, result)
}

func TestClient_OverwriteTXTRecord(t *testing.T) {
	client := mockBuilder("secret-apikey", "").
		Route("PUT /domains/example.com/records/_acme-challenge/TXT",
			servermock.ResponseFromFixture("api_response.json"),
			servermock.CheckRequestJSONBodyFromFixture("put_txt_record-request.json"),
		).
		Build(t)

	record := &Record{
		TTL: 300,
		Values: []string{
			"ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
			"value1",
		},
	}

	err := client.OverwriteTXTRecord(t.Context(), "example.com", "_acme-challenge", record)
	require.NoError(t, err)
}

func TestClient_DeleteTXTRecord(t *testing.T) {
	client := mockBuilder("", "secret-pat").
		Route("DELETE /domains/example.com/records/_acme-challenge/TXT",
			servermock.Noop().
				WithStatusCode(http.StatusNoContent),
		).
		Build(t)

	err := client.DeleteTXTRecord(t.Context(), "example.com", "_acme-challenge")
	require.NoError(t, err)
}

func TestClient_DeleteTXTRecord_error_400(t *testing.T) {
	client := mockBuilder("", "secret-pat").
		Route("DELETE /domains/example.com/records/_acme-challenge/TXT",
			servermock.ResponseFromFixture("error_400.json").
				WithStatusCode(http.StatusBadRequest),
		).
		Build(t)

	err := client.DeleteTXTRecord(t.Context(), "example.com", "_acme-challenge")
	require.EqualError(t, err, "(400) The server could not comply with the request since it is either malformed or otherwise incorrect.: Bad Request: HTTPBadRequest")
}

func TestClient_DeleteTXTRecord_error_401(t *testing.T) {
	client := mockBuilder("", "secret-pat").
		Route("DELETE /domains/example.com/records/_acme-challenge/TXT",
			servermock.ResponseFromFixture("error_401.json").
				WithStatusCode(http.StatusUnauthorized),
		).
		Build(t)

	err := client.DeleteTXTRecord(t.Context(), "example.com", "_acme-challenge")
	require.EqualError(t, err, "(401) You must provide an access token or an API Key.")
}
