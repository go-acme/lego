package internal

import (
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

			client.baseURL, _ = url.Parse(server.URL)
			client.HTTPClient = server.Client()

			return client, nil
		},
		servermock.CheckHeader().WithJSONHeaders().
			WithAuthorization("Bearer secret"),
	)
}

func TestClient_CreateRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /domains/example.com/records",
			servermock.ResponseFromFixture("create_record.json"),
			servermock.CheckRequestJSONBodyFromFixture("create_record-request.json")).
		Build(t)

	record := Record{
		Type: "TXT",
		Name: "foo",
		Data: "txtTXTtxt",
		TTL:  300,
	}

	rec, err := client.CreateRecord(t.Context(), "example.com", record)
	require.NoError(t, err)

	expected := &Record{
		ID:   123,
		Type: "TXT",
		Name: "foo",
		Data: "txtTXTtxt",
		TTL:  300,
	}

	require.Equal(t, expected, rec)
}

func TestClient_CreateRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /domains/example.com/records",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusBadRequest)).
		Build(t)

	record := Record{
		Type: "TXT",
		Name: "foo",
		Data: "txtTXTtxt",
		TTL:  300,
	}

	_, err := client.CreateRecord(t.Context(), "example.com", record)
	require.EqualError(t, err, "400: type: title: detail: instance: property1: a")
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /domains/example.com/records/123",
			servermock.Noop().
				WithStatusCode(http.StatusNoContent)).
		Build(t)

	err := client.DeleteRecord(t.Context(), "example.com", 123)
	require.NoError(t, err)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /domains/example.com/records/123",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusBadRequest)).
		Build(t)

	err := client.DeleteRecord(t.Context(), "example.com", 123)
	require.EqualError(t, err, "400: type: title: detail: instance: property1: a")
}
