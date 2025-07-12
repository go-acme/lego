package internal

import (
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/stubrouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *stubrouter.Builder[*Client] {
	return stubrouter.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client, err := NewClient(server.URL, "secret")
			if err != nil {
				return nil, err
			}

			client.HTTPClient = server.Client()

			return client, nil
		},
		stubrouter.CheckHeader().WithContentTypeFromURLEncoded())
}

func TestClient_AddRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /api/zones/records/add",
			stubrouter.ResponseFromFixture("add-record.json"),
			stubrouter.CheckForm().Strict().
				With("domain", "_acme-challenge.example.com").
				With("text", "txtTXTtxt").
				With("type", "TXT").
				With("token", "secret")).
		Build(t)

	record := Record{
		Domain: "_acme-challenge.example.com",
		Type:   "TXT",
		Text:   "txtTXTtxt",
	}

	newRecord, err := client.AddRecord(t.Context(), record)
	require.NoError(t, err)

	expected := &Record{Name: "example.com", Type: "A"}

	assert.Equal(t, expected, newRecord)
}

func TestClient_AddRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /api/zones/records/add",
			stubrouter.ResponseFromFixture("error.json")).
		Build(t)

	record := Record{
		Domain: "_acme-challenge.example.com",
		Type:   "TXT",
		Text:   "txtTXTtxt",
	}

	_, err := client.AddRecord(t.Context(), record)
	require.Error(t, err)

	assert.EqualError(t, err, "Status: error, ErrorMessage: error message, StackTrace: application stack trace, InnerErrorMessage: inner exception message")
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /api/zones/records/delete",
			stubrouter.ResponseFromFixture("delete-record.json"),
			stubrouter.CheckForm().Strict().
				With("domain", "_acme-challenge.example.com").
				With("text", "txtTXTtxt").
				With("type", "TXT").
				With("token", "secret")).
		Build(t)

	record := Record{
		Domain: "_acme-challenge.example.com",
		Type:   "TXT",
		Text:   "txtTXTtxt",
	}

	err := client.DeleteRecord(t.Context(), record)
	require.NoError(t, err)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /api/zones/records/delete",
			stubrouter.ResponseFromFixture("error.json")).
		Build(t)

	record := Record{
		Domain: "_acme-challenge.example.com",
		Type:   "TXT",
		Text:   "txtTXTtxt",
	}

	err := client.DeleteRecord(t.Context(), record)
	require.Error(t, err)

	assert.EqualError(t, err, "Status: error, ErrorMessage: error message, StackTrace: application stack trace, InnerErrorMessage: inner exception message")
}
