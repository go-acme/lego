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
			client := NewClient(OAuthStaticAccessToken(server.Client(), "secret"))
			client.BaseURL, _ = url.Parse(server.URL)

			return client, nil
		},
		servermock.CheckHeader().WithJSONHeaders().
			WithAuthorization("Bearer secret"))
}

func TestClient_AddTxtRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /v2/domains/example.com/records",
			servermock.ResponseFromFixture("domains-records_POST.json").
				WithStatusCode(http.StatusCreated),
			servermock.CheckRequestJSONBody(`{"type":"TXT","name":"_acme-challenge.example.com.","data":"w6uP8Tcg6K2QR905Rms8iXTlksL6OD1KOWBxTK7wxPI","ttl":30}`)).
		Build(t)

	record := Record{
		Type: "TXT",
		Name: "_acme-challenge.example.com.",
		Data: "w6uP8Tcg6K2QR905Rms8iXTlksL6OD1KOWBxTK7wxPI",
		TTL:  30,
	}

	newRecord, err := client.AddTxtRecord(t.Context(), "example.com", record)
	require.NoError(t, err)

	expected := &TxtRecordResponse{DomainRecord: Record{
		ID:   1234567,
		Type: "TXT",
		Name: "_acme-challenge",
		Data: "w6uP8Tcg6K2QR905Rms8iXTlksL6OD1KOWBxTK7wxPI",
		TTL:  0,
	}}

	assert.Equal(t, expected, newRecord)
}

func TestClient_RemoveTxtRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /v2/domains/example.com/records/1234567",
			servermock.ResponseFromFixture("domains-records_POST.json").
				WithStatusCode(http.StatusNoContent)).
		Build(t)

	err := client.RemoveTxtRecord(t.Context(), "example.com", 1234567)
	require.NoError(t, err)
}
