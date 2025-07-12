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

func setupClient(server *httptest.Server) (*Client, error) {
	client := NewClient("secret")
	client.HTTPClient = server.Client()
	client.baseURL, _ = url.Parse(server.URL)

	return client, nil
}

func TestClient_UpdateRecords_error(t *testing.T) {
	client := servermock.NewBuilder[*Client](setupClient, servermock.CheckHeader().WithJSONHeaders()).
		Route("PATCH /v1/acme/zones/example.org/rrsets",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnprocessableEntity)).
		Build(t)

	rrSet := []UpdateRRSet{{
		Name:       "acme.example.org.",
		ChangeType: "add",
		Type:       "TXT",
		Records:    []Record{{Content: `"my-acme-challenge"`}},
	}}

	resp, err := client.UpdateRecords(t.Context(), "example.org", rrSet)
	require.ErrorAs(t, err, new(*APIResponse))
	assert.Nil(t, resp)
}

func TestClient_UpdateRecords(t *testing.T) {
	client := servermock.NewBuilder[*Client](setupClient, servermock.CheckHeader().WithJSONHeaders()).
		Route("PATCH /v1/acme/zones/example.org/rrsets",
			servermock.ResponseFromFixture("rrsets-response.json")).
		Build(t)

	rrSet := []UpdateRRSet{{
		Name:       "acme.example.org.",
		ChangeType: "add",
		Type:       "TXT",
		Records:    []Record{{Content: `"my-acme-challenge"`}},
	}}

	resp, err := client.UpdateRecords(t.Context(), "example.org", rrSet)
	require.NoError(t, err)

	expected := &APIResponse{Status: "ok", Message: "RRsets updated"}

	assert.Equal(t, expected, resp)
}
