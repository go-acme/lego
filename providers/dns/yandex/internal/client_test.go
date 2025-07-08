package internal

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/clientmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupClient(server *httptest.Server) (*Client, error) {
	client, err := NewClient("lego")
	if err != nil {
		return nil, err
	}

	client.HTTPClient = server.Client()
	client.baseURL, _ = url.Parse(server.URL)

	return client, nil
}

func TestAddRecord(t *testing.T) {
	client := clientmock.NewBuilder[*Client](setupClient).
		Route("POST /add",
			clientmock.ResponseFromFixture("add_record.json"),
			clientmock.CheckHeader().
				WithContentTypeFromURLEncoded(),
			clientmock.CheckForm().Strict().
				With("domain", "example.com").
				With("subdomain", "foo").
				With("ttl", "300").
				With("content", "txtTXTtxtTXTtxtTXT").
				With("type", "TXT")).
		Build(t)

	data := Record{
		Domain:    "example.com",
		Type:      "TXT",
		Content:   "txtTXTtxtTXTtxtTXT",
		SubDomain: "foo",
		TTL:       300,
	}

	record, err := client.AddRecord(t.Context(), data)
	require.NoError(t, err)
	require.NotNil(t, record)
}

func TestAddRecord_error(t *testing.T) {
	client := clientmock.NewBuilder[*Client](setupClient).
		Route("POST /add",
			clientmock.ResponseFromFixture("add_record_error.json"),
			clientmock.CheckHeader().
				WithContentTypeFromURLEncoded()).
		Build(t)

	data := Record{
		Domain:    "example.com",
		Type:      "TXT",
		Content:   "txtTXTtxtTXTtxtTXT",
		SubDomain: "foo",
		TTL:       300,
	}

	_, err := client.AddRecord(t.Context(), data)
	require.EqualError(t, err, "error during operation: error bad things")
}

func TestRemoveRecord(t *testing.T) {
	client := clientmock.NewBuilder[*Client](setupClient).
		Route("POST /del",
			clientmock.ResponseFromFixture("remove_record.json"),
			clientmock.CheckHeader().
				WithContentTypeFromURLEncoded(),
			clientmock.CheckForm().Strict().
				With("domain", "example.com").
				With("record_id", "6")).
		Build(t)

	data := Record{
		ID:     6,
		Domain: "example.com",
	}

	id, err := client.RemoveRecord(t.Context(), data)
	require.NoError(t, err)

	assert.Equal(t, 6, id)
}

func TestRemoveRecord_error(t *testing.T) {
	client := clientmock.NewBuilder[*Client](setupClient).
		Route("POST /del",
			clientmock.ResponseFromFixture("remove_record_error.json"),
			clientmock.CheckHeader().
				WithContentTypeFromURLEncoded()).
		Build(t)

	data := Record{
		ID:     6,
		Domain: "example.com",
	}

	_, err := client.RemoveRecord(t.Context(), data)
	require.EqualError(t, err, "error during operation: error bad things")
}

func TestGetRecords(t *testing.T) {
	client := clientmock.NewBuilder[*Client](setupClient).
		Route("GET /list",
			clientmock.ResponseFromFixture("get_records.json"),
			clientmock.CheckForm().Strict().
				With("domain", "example.com")).
		Build(t)

	records, err := client.GetRecords(t.Context(), "example.com")
	require.NoError(t, err)

	require.Len(t, records, 2)
}

func TestGetRecords_error(t *testing.T) {
	client := clientmock.NewBuilder[*Client](setupClient).
		Route("GET /list",
			clientmock.ResponseFromFixture("get_records_error.json")).
		Build(t)

	_, err := client.GetRecords(t.Context(), "example.com")
	require.EqualError(t, err, "error during operation: error bad things")
}
