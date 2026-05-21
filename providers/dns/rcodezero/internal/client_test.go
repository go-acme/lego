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

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client := NewClient("secret")
			client.HTTPClient = server.Client()
			client.BaseURL, _ = url.Parse(server.URL)

			return client, nil
		},
		servermock.CheckHeader().
			WithJSONHeaders(),
	)
}

func TestClient_UpdateRecords(t *testing.T) {
	client := mockBuilder().
		Route("PATCH /v1/acme/zones/example.com/rrsets",
			servermock.ResponseFromFixture("update_rrsets.json"),
			servermock.CheckRequestJSONBodyFromFixture("update_rrsets_add-request.json"),
		).
		Build(t)

	rrSet := []UpdateRRSet{{
		Name:       "acme.example.com.",
		ChangeType: "add",
		Type:       "TXT",
		Records:    []Record{{Content: `"ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY"`}},
		TTL:        120,
	}}

	resp, err := client.UpdateRecords(t.Context(), "example.com", rrSet)
	require.NoError(t, err)

	expected := &APIResponse{Status: "ok", Message: "RRsets updated"}

	assert.Equal(t, expected, resp)
}

func TestClient_UpdateRecords_error(t *testing.T) {
	client := mockBuilder().
		Route("PATCH /v1/acme/zones/example.com/rrsets",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnprocessableEntity),
		).
		Build(t)

	rrSet := []UpdateRRSet{{
		Name:       "acme.example.com.",
		ChangeType: "add",
		Type:       "TXT",
		Records:    []Record{{Content: `"ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY"`}},
	}}

	resp, err := client.UpdateRecords(t.Context(), "example.com", rrSet)
	require.ErrorAs(t, err, new(*APIResponse))
	assert.Nil(t, resp)
}
