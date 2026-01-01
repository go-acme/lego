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
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client, err := NewClient("user", "secret")
			if err != nil {
				return nil, err
			}

			client.BaseURL, _ = url.Parse(server.URL)
			client.HTTPClient = server.Client()

			return client, nil
		},
		servermock.CheckHeader().
			WithJSONHeaders().
			WithBasicAuth("user", "secret"),
	)
}

func TestClient_CreateRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /CreateRecord",
			servermock.ResponseFromFixture("createRecord.json"),
			servermock.CheckRequestJSONBodyFromFixture("createRecord-request.json"),
		).
		Build(t)

	record := Record{
		DomainID: 123,
		Name:     "_acme-challenge.example.com",
		Content:  "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		Type:     "TXT",
		TTL:      120,
		Active:   true,
	}

	result, err := client.CreateRecord(t.Context(), record)
	require.NoError(t, err)

	expected := &CommonResponse{
		Status: true,
		ID:     456,
		Error:  "",
	}

	assert.Equal(t, expected, result)
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /DeleteRecord",
			servermock.ResponseFromFixture("deleteRecord.json"),
			servermock.CheckQueryParameter().Strict().
				With("id", "456"),
		).
		Build(t)

	err := client.DeleteRecord(t.Context(), 456)
	require.NoError(t, err)
}

func TestClient_DeleteRecord_error400(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /DeleteRecord",
			servermock.ResponseFromFixture("error-400.json").
				WithStatusCode(http.StatusBadRequest),
		).
		Build(t)

	err := client.DeleteRecord(t.Context(), 456)
	require.EqualError(t, err,
		"400: One or more validation errors occurred., "+
			"type: https://www.rfc-editor.org/rfc/rfc7231#section-6.5.1, instance: /api/route, traceId: 0HMPNHL0JHL76:00000001, detail: detail, "+
			"code: Error reason: Error or field name: severity")
}

func TestClient_DeleteRecord_error500(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /DeleteRecord",
			servermock.ResponseFromFixture("error-500.json").
				WithStatusCode(http.StatusInternalServerError),
		).
		Build(t)

	err := client.DeleteRecord(t.Context(), 456)
	require.EqualError(t, err, "500: Internal Server Error!: Something unexpected has happened: See application log for stack trace.")
}

func TestClient_GetDomains(t *testing.T) {
	client := mockBuilder().
		Route("GET /GetDomains",
			servermock.ResponseFromFixture("getDomains.json"),
		).
		Build(t)

	domains, err := client.GetDomains(t.Context())
	require.NoError(t, err)

	expected := []Domain{
		{
			ID:         123,
			Name:       "example.com",
			OwnerEmail: "user@example.com",
			DefaultNs1: "ns1.r4ns.com",
			DefaultNs2: "ns2.r4ns.net",
			SoaRefresh: 10800,
			SoaExpiry:  604800,
			SoaRetry:   3600,
			SoaNx:      3600,
			NsTTL:      3600,
			Online:     true,
			Secure:     true,
			APIAccess:  true,
			Created:    time.Date(2024, time.December, 9, 12, 35, 2, 751157300, time.UTC),
			Updated:    time.Date(2025, time.December, 9, 10, 35, 2, 751158400, time.UTC),
			SubnetMask: -32768,
			Type:       "NATIVE",
		},
	}

	assert.Equal(t, expected, domains)
}
