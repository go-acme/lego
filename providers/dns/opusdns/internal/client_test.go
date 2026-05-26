package internal

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
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
			With(AuthenticationHeader, "secret"),
	)
}

func TestClient_PatchRecords_upsert(t *testing.T) {
	client := mockBuilder().
		Route("PATCH /v1/dns/example.com/records",
			servermock.Noop(),
			servermock.CheckRequestJSONBodyFromFixture("patch_records-upsert-request.json"),
		).
		Build(t)

	ops := []RecordOperation{{
		Op: RecordOperationUpset,
		Record: Record{
			Name:  "_acme-challenge",
			Type:  "TXT",
			TTL:   120,
			RData: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		},
	}}

	err := client.PatchRecords(t.Context(), "example.com", ops)
	require.NoError(t, err)
}

func TestClient_PatchRecords_remove(t *testing.T) {
	client := mockBuilder().
		Route("PATCH /v1/dns/example.com/records",
			servermock.Noop(),
			servermock.CheckRequestJSONBodyFromFixture("patch_records-remove-request.json"),
		).
		Build(t)

	ops := []RecordOperation{{
		Op: RecordOperationRemove,
		Record: Record{
			Name:  "_acme-challenge",
			Type:  "TXT",
			TTL:   120,
			RData: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		},
	}}

	err := client.PatchRecords(t.Context(), "example.com", ops)
	require.NoError(t, err)
}

func TestClient_PatchRecords_error(t *testing.T) {
	client := mockBuilder().
		Route("PATCH /v1/dns/example.com/records",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusForbidden),
		).
		Build(t)

	ops := []RecordOperation{{
		Op: RecordOperationUpset,
		Record: Record{
			Name:  "_acme-challenge",
			Type:  "TXT",
			TTL:   120,
			RData: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		},
	}}

	err := client.PatchRecords(t.Context(), "example.com", ops)
	require.EqualError(t, err, "[status: 403] Forbidden: Insufficient permissions for this zone.")
}
