package internal

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client, _ := NewClient(server.URL, "user", "secret")
			client.HTTPClient = server.Client()

			return client, nil
		},
		servermock.CheckHeader().
			WithContentTypeFromURLEncoded())
}

func newAPIError(reason string, a ...any) APIError {
	return APIError{
		Message: "Cannot View Dns Record",
		Result:  fmt.Sprintf(reason, a...),
	}
}

func TestClient_SetRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /CMD_API_DNS_CONTROL", nil,
			servermock.CheckQueryParameter().Strict().
				With("domain", "example.com").
				With("json", "yes"),
			servermock.CheckForm().UsePostForm().Strict().
				With("action", "add").
				With("name", "foo").
				With("type", "TXT").
				With("value", "txtTXTtxt").
				With("ttl", "123"),
		).
		Build(t)

	record := Record{
		Name:  "foo",
		Type:  "TXT",
		Value: "txtTXTtxt",
		TTL:   123,
	}

	err := client.SetRecord(t.Context(), "example.com", record)
	require.NoError(t, err)
}

func TestClient_SetRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /CMD_API_DNS_CONTROL",
			servermock.JSONEncode(newAPIError("OOPS")).
				WithStatusCode(http.StatusInternalServerError)).
		Build(t)

	record := Record{
		Name:  "foo",
		Type:  "TXT",
		Value: "txtTXTtxt",
		TTL:   123,
	}

	err := client.SetRecord(t.Context(), "example.com", record)
	require.EqualError(t, err, "[status code 500] Cannot View Dns Record: OOPS")
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /CMD_API_DNS_CONTROL", nil,
			servermock.CheckQueryParameter().Strict().
				With("domain", "example.com").
				With("json", "yes"),
			servermock.CheckForm().UsePostForm().Strict().
				With("action", "delete").
				With("name", "foo").
				With("type", "TXT").
				With("value", "txtTXTtxt"),
		).
		Build(t)

	record := Record{
		Name:  "foo",
		Type:  "TXT",
		Value: "txtTXTtxt",
	}

	err := client.DeleteRecord(t.Context(), "example.com", record)
	require.NoError(t, err)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /CMD_API_DNS_CONTROL",
			servermock.JSONEncode(newAPIError("OOPS")).
				WithStatusCode(http.StatusInternalServerError)).
		Build(t)

	record := Record{
		Name:  "foo",
		Type:  "TXT",
		Value: "txtTXTtxt",
	}

	err := client.DeleteRecord(t.Context(), "example.com", record)
	require.EqualError(t, err, "[status code 500] Cannot View Dns Record: OOPS")
}
