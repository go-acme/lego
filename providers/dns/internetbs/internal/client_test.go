package internal

import (
	"fmt"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testBaseURL = "https://testapi.internet.bs"

const (
	testAPIKey   = "testapi"
	testPassword = "testpass"
)

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client := NewClient(testAPIKey, testPassword)
			client.baseURL, _ = url.Parse(server.URL)
			client.HTTPClient = server.Client()

			return client, nil
		},
		servermock.CheckHeader().
			WithContentTypeFromURLEncoded(),
	)
}

func TestClient_AddRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /Domain/DnsRecord/Add",
			servermock.ResponseFromFixture("Domain_DnsRecord_Add_SUCCESS.json"),
			servermock.CheckForm().Strict().
				With("fullrecordname", "www.example.com").
				With("ttl", "36000").
				With("type", "TXT").
				With("value", "xxx").
				With("password", testPassword).
				With("apiKey", testAPIKey).
				With("ResponseFormat", "JSON")).
		Build(t)

	query := RecordQuery{
		FullRecordName: "www.example.com",
		Type:           "TXT",
		Value:          "xxx",
		TTL:            36000,
	}

	err := client.AddRecord(t.Context(), query)
	require.NoError(t, err)
}

func TestClient_AddRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /Domain/DnsRecord/Add",
			servermock.ResponseFromFixture("Domain_DnsRecord_Add_FAILURE.json")).
		Build(t)

	query := RecordQuery{
		FullRecordName: "www.example.com.",
		Type:           "TXT",
		Value:          "xxx",
		TTL:            36000,
	}

	err := client.AddRecord(t.Context(), query)
	require.Error(t, err)
}

func TestClient_AddRecord_integration(t *testing.T) {
	env, ok := os.LookupEnv("INTERNET_BS_DEBUG")
	if !ok {
		t.Skip("skip integration test")
	}

	client := NewClient(testAPIKey, testPassword)
	client.baseURL, _ = url.Parse(testBaseURL)
	client.debug, _ = strconv.ParseBool(env)

	query := RecordQuery{
		FullRecordName: "www.example.com",
		Type:           "TXT",
		Value:          "xxx",
		TTL:            36000,
	}

	err := client.AddRecord(t.Context(), query)
	require.NoError(t, err)

	query = RecordQuery{
		FullRecordName: "www.example.com",
		Type:           "TXT",
		Value:          "yyy",
		TTL:            36000,
	}

	err = client.AddRecord(t.Context(), query)
	require.NoError(t, err)
}

func TestClient_RemoveRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /Domain/DnsRecord/Remove",
			servermock.ResponseFromFixture("Domain_DnsRecord_Remove_SUCCESS.json"),
			servermock.CheckForm().Strict().
				With("fullrecordname", "www.example.com").
				With("type", "TXT").
				With("password", testPassword).
				With("apiKey", testAPIKey).
				With("ResponseFormat", "JSON")).
		Build(t)

	query := RecordQuery{
		FullRecordName: "www.example.com",
		Type:           "TXT",
		Value:          "",
	}
	err := client.RemoveRecord(t.Context(), query)
	require.NoError(t, err)
}

func TestClient_RemoveRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /Domain/DnsRecord/Remove",
			servermock.ResponseFromFixture("Domain_DnsRecord_Remove_FAILURE.json")).
		Build(t)

	query := RecordQuery{
		FullRecordName: "www.example.com.",
		Type:           "TXT",
		Value:          "",
	}
	err := client.RemoveRecord(t.Context(), query)
	require.Error(t, err)
}

func TestClient_RemoveRecord_integration(t *testing.T) {
	env, ok := os.LookupEnv("INTERNET_BS_DEBUG")
	if !ok {
		t.Skip("skip integration test")
	}

	client := NewClient(testAPIKey, testPassword)
	client.baseURL, _ = url.Parse(testBaseURL)
	client.debug, _ = strconv.ParseBool(env)

	query := RecordQuery{
		FullRecordName: "www.example.com",
		Type:           "TXT",
		Value:          "",
	}

	err := client.RemoveRecord(t.Context(), query)
	require.NoError(t, err)
}

func TestClient_ListRecords(t *testing.T) {
	client := mockBuilder().
		Route("POST /Domain/DnsRecord/List",
			servermock.ResponseFromFixture("Domain_DnsRecord_List_SUCCESS.json"),
			servermock.CheckForm().Strict().
				With("Domain", "example.com").
				With("password", testPassword).
				With("apiKey", testAPIKey).
				With("ResponseFormat", "JSON")).
		Build(t)

	query := ListRecordQuery{
		Domain: "example.com",
	}

	records, err := client.ListRecords(t.Context(), query)
	require.NoError(t, err)

	expected := []Record{
		{
			Name:  "example.com",
			Value: "ns-hongkong.internet.bs",
			TTL:   3600,
			Type:  "NS",
		},
		{
			Name:  "example.com",
			Value: "ns-toronto.internet.bs",
			TTL:   3600,
			Type:  "NS",
		},
		{
			Name:  "example.com",
			Value: "ns-london.internet.bs",
			TTL:   3600,
			Type:  "NS",
		},
		{
			Name:  "test.example.com",
			Value: "example1.com",
			TTL:   3600,
			Type:  "CNAME",
		},
		{
			Name:  "www.example.com",
			Value: "xxx",
			TTL:   36000,
			Type:  "TXT",
		},
		{
			Name:  "www.example.com",
			Value: "yyy",
			TTL:   36000,
			Type:  "TXT",
		},
	}

	assert.Equal(t, expected, records)
}

func TestClient_ListRecords_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /Domain/DnsRecord/List",
			servermock.ResponseFromFixture("Domain_DnsRecord_List_FAILURE.json")).
		Build(t)

	query := ListRecordQuery{
		Domain: "www.example.com",
	}

	_, err := client.ListRecords(t.Context(), query)
	require.Error(t, err)
}

func TestClient_ListRecords_integration(t *testing.T) {
	env, ok := os.LookupEnv("INTERNET_BS_DEBUG")
	if !ok {
		t.Skip("skip integration test")
	}

	client := NewClient(testAPIKey, testPassword)
	client.baseURL, _ = url.Parse(testBaseURL)
	client.debug, _ = strconv.ParseBool(env)

	query := ListRecordQuery{
		Domain: "example.com",
	}

	records, err := client.ListRecords(t.Context(), query)
	require.NoError(t, err)

	for _, record := range records {
		fmt.Println(record)
	}
}
