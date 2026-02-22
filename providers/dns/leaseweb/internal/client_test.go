package internal

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v5/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client, err := NewClient("secret")
			if err != nil {
				return nil, err
			}

			client.BaseURL, _ = url.Parse(server.URL)
			client.HTTPClient = server.Client()

			return client, nil
		},
		servermock.CheckHeader().
			WithJSONHeaders().
			With(AuthHeader, "secret"),
	)
}

func TestClient_CreateRRSet(t *testing.T) {
	client := mockBuilder().
		Route("POST /domains/example.com/resourceRecordSets",
			servermock.ResponseFromFixture("createResourceRecordSet.json"),
			servermock.CheckRequestJSONBodyFromFixture("createResourceRecordSet-request.json"),
		).
		Build(t)

	rrset := RRSet{
		Content: []string{"ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY"},
		Name:    "_acme-challenge.example.com.",
		TTL:     300,
		Type:    "TXT",
	}

	result, err := client.CreateRRSet(t.Context(), "example.com", rrset)
	require.NoError(t, err)

	expected := &RRSet{
		Content:  []string{"ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY"},
		Name:     "_acme-challenge.example.com.",
		Editable: true,
		TTL:      300,
		Type:     "TXT",
	}

	assert.Equal(t, expected, result)
}

func TestClient_GetRRSet(t *testing.T) {
	client := mockBuilder().
		Route("GET /domains/example.com/resourceRecordSets/_acme-challenge.example.com./TXT",
			servermock.ResponseFromFixture("getResourceRecordSet.json"),
		).
		Build(t)

	result, err := client.GetRRSet(t.Context(), "example.com", "_acme-challenge.example.com.", "TXT")
	require.NoError(t, err)

	expected := &RRSet{
		Content:  []string{"foo", "Now36o-3BmlB623-0c1qCIUmgWVVmDJb88KGl24pqpo"},
		Name:     "_acme-challenge.example.com.",
		Editable: true,
		TTL:      3600,
		Type:     "TXT",
	}

	assert.Equal(t, expected, result)
}

func TestClient_GetRRSet_error_404(t *testing.T) {
	client := mockBuilder().
		Route("GET /domains/example.com/resourceRecordSets/_acme-challenge.example.com./TXT",
			servermock.ResponseFromFixture("error_404.json").
				WithStatusCode(http.StatusNotFound),
		).
		Build(t)

	_, err := client.GetRRSet(t.Context(), "example.com", "_acme-challenge.example.com.", "TXT")
	require.EqualError(t, err, "404: Resource not found (289346a1-3eaf-4da4-b707-62ef12eb08be)")

	target := &NotFoundError{}
	require.ErrorAs(t, err, &target)
}

func TestClient_UpdateRRSet(t *testing.T) {
	client := mockBuilder().
		Route("PUT /domains/example.com/resourceRecordSets/_acme-challenge.example.com./TXT",
			servermock.ResponseFromFixture("updateResourceRecordSet.json"),
			servermock.CheckRequestJSONBodyFromFixture("updateResourceRecordSet-request.json"),
		).
		Build(t)

	rrset := RRSet{
		Content: []string{"foo", "Now36o-3BmlB623-0c1qCIUmgWVVmDJb88KGl24pqpo", "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY"},
		Name:    "_acme-challenge.example.com.",
		TTL:     3600,
		Type:    "TXT",
	}

	result, err := client.UpdateRRSet(t.Context(), "example.com", rrset)
	require.NoError(t, err)

	expected := &RRSet{
		Content:  []string{"foo", "Now36o-3BmlB623-0c1qCIUmgWVVmDJb88KGl24pqpo", "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY"},
		Name:     "_acme-challenge.example.com.",
		Editable: true,
		TTL:      3600,
		Type:     "TXT",
	}

	assert.Equal(t, expected, result)
}

func TestClient_DeleteRRSet(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /domains/example.com/resourceRecordSets/_acme-challenge.example.com./TXT",
			servermock.Noop().
				WithStatusCode(http.StatusNoContent),
		).
		Build(t)

	err := client.DeleteRRSet(t.Context(), "example.com", "_acme-challenge.example.com.", "TXT")
	require.NoError(t, err)
}

func TestClient_DeleteRRSet_error(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /domains/example.com/resourceRecordSets/_acme-challenge.example.com./TXT",
			servermock.ResponseFromFixture("error_401.json").
				WithStatusCode(http.StatusUnauthorized),
		).
		Build(t)

	err := client.DeleteRRSet(t.Context(), "example.com", "_acme-challenge.example.com.", "TXT")
	require.EqualError(t, err, "401: You are not authorized to view this resource. (289346a1-3eaf-4da4-b707-62ef12eb08be)")
}
