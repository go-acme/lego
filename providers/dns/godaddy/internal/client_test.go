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
			client := NewClient("key", "secret")
			client.HTTPClient = server.Client()
			client.baseURL, _ = url.Parse(server.URL)

			return client, nil
		},
		servermock.CheckHeader().WithJSONHeaders().
			WithAuthorization("sso-key key:secret"))
}

func TestClient_GetRecords(t *testing.T) {
	client := mockBuilder().
		Route("GET /v1/domains/example.com/records/TXT/", servermock.ResponseFromFixture("getrecords.json")).
		Build(t)

	records, err := client.GetRecords(t.Context(), "example.com", "TXT", "")
	require.NoError(t, err)

	expected := []DNSRecord{
		{Name: "_acme-challenge", Type: "TXT", Data: " ", TTL: 600},
		{Name: "_acme-challenge.example", Type: "TXT", Data: "6rrai7-jm7l3PxL4WGmGoS6VMeefSHx24r-qCvUIOxU", TTL: 600},
		{Name: "_acme-challenge.example", Type: "TXT", Data: "8Axt-PXQvjOVD2oi2YXqyyn8U5CDcC8P-BphlCxk3Ek", TTL: 600},
		{Name: "_acme-challenge.lego", Type: "TXT", Data: " ", TTL: 600},
		{Name: "_acme-challenge.lego", Type: "TXT", Data: "0Ad60wO_yxxJPFPb1deir6lQ37FPLeA02YCobo7Qm8A", TTL: 600},
		{Name: "_acme-challenge.lego", Type: "TXT", Data: "acme", TTL: 600},
	}

	assert.Equal(t, expected, records)
}

func TestClient_GetRecords_errors(t *testing.T) {
	client := mockBuilder().
		Route("GET /v1/domains/example.com/records/TXT/",
			servermock.ResponseFromFixture("errors.json").WithStatusCode(http.StatusUnprocessableEntity)).
		Build(t)

	records, err := client.GetRecords(t.Context(), "example.com", "TXT", "")
	require.EqualError(t, err, "[status code: 422] INVALID_BODY: Request body doesn't fulfill schema, see details in `fields`")
	assert.Nil(t, records)
}

func TestClient_UpdateTxtRecords(t *testing.T) {
	client := mockBuilder().
		Route("PUT /v1/domains/example.com/records/TXT/lego", nil,
			servermock.CheckRequestJSONBodyFromFile("update_records-request.json")).
		Build(t)

	records := []DNSRecord{
		{Name: "_acme-challenge", Type: "TXT", Data: " ", TTL: 600},
		{Name: "_acme-challenge.example", Type: "TXT", Data: "6rrai7-jm7l3PxL4WGmGoS6VMeefSHx24r-qCvUIOxU", TTL: 600},
		{Name: "_acme-challenge.example", Type: "TXT", Data: "8Axt-PXQvjOVD2oi2YXqyyn8U5CDcC8P-BphlCxk3Ek", TTL: 600},
		{Name: "_acme-challenge.lego", Type: "TXT", Data: " ", TTL: 600},
		{Name: "_acme-challenge.lego", Type: "TXT", Data: "0Ad60wO_yxxJPFPb1deir6lQ37FPLeA02YCobo7Qm8A", TTL: 600},
		{Name: "_acme-challenge.lego", Type: "TXT", Data: "acme", TTL: 600},
	}

	err := client.UpdateTxtRecords(t.Context(), records, "example.com", "lego")
	require.NoError(t, err)
}

func TestClient_UpdateTxtRecords_errors(t *testing.T) {
	client := mockBuilder().
		Route("PUT /v1/domains/example.com/records/TXT/lego",
			servermock.ResponseFromFixture("errors.json").WithStatusCode(http.StatusUnprocessableEntity),
			servermock.CheckRequestJSONBodyFromFile("update_records-request.json")).
		Build(t)

	records := []DNSRecord{
		{Name: "_acme-challenge", Type: "TXT", Data: " ", TTL: 600},
		{Name: "_acme-challenge.example", Type: "TXT", Data: "6rrai7-jm7l3PxL4WGmGoS6VMeefSHx24r-qCvUIOxU", TTL: 600},
		{Name: "_acme-challenge.example", Type: "TXT", Data: "8Axt-PXQvjOVD2oi2YXqyyn8U5CDcC8P-BphlCxk3Ek", TTL: 600},
		{Name: "_acme-challenge.lego", Type: "TXT", Data: " ", TTL: 600},
		{Name: "_acme-challenge.lego", Type: "TXT", Data: "0Ad60wO_yxxJPFPb1deir6lQ37FPLeA02YCobo7Qm8A", TTL: 600},
		{Name: "_acme-challenge.lego", Type: "TXT", Data: "acme", TTL: 600},
	}

	err := client.UpdateTxtRecords(t.Context(), records, "example.com", "lego")
	require.EqualError(t, err, "[status code: 422] INVALID_BODY: Request body doesn't fulfill schema, see details in `fields`")
}

func TestClient_DeleteTxtRecords(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /v1/domains/example.com/records/TXT/foo",
			servermock.Noop().WithStatusCode(http.StatusNoContent)).
		Build(t)

	err := client.DeleteTxtRecords(t.Context(), "example.com", "foo")
	require.NoError(t, err)
}

func TestClient_DeleteTxtRecords_errors(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /v1/domains/example.com/records/TXT/foo",
			servermock.ResponseFromFixture("error-extended.json").WithStatusCode(http.StatusConflict)).
		Build(t)

	err := client.DeleteTxtRecords(t.Context(), "example.com", "foo")
	require.EqualError(t, err, "[status code: 409] ACCESS_DENIED: Authenticated user is not allowed access [test: content (path=/foo) (pathRelated=/bar)]")
}
