package internal

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client := NewClient("user", "secret")
			client.baseURL = server.URL
			client.HTTPClient = server.Client()

			return client, nil
		},
		servermock.CheckHeader().
			WithContentTypeFromURLEncoded())
}

func TestClient_GetRecords(t *testing.T) {
	client := mockBuilder().
		Route("POST /",
			servermock.ResponseFromFixture(commandDNSRowsList+".json"),
			checkFormRequest(`{"request":{"user":"user","auth":"xxx","command":"dns-rows-list","data":{"domain":"example.com"}}}`)).
		Build(t)

	records, err := client.GetRecords(t.Context(), "example.com.")
	require.NoError(t, err)

	assert.Len(t, records, 4)

	expected := []DNSRow{
		{
			ID:   "911",
			TTL:  "1800",
			Type: "A",
			Data: "1.2.3.4",
		},
		{
			ID:   "913",
			TTL:  "1800",
			Type: "MX",
			Data: "1 mail1.wedos.net",
		},
		{
			ID:   "914",
			TTL:  "1800",
			Type: "MX",
			Data: "10 mailbackup.wedos.net",
		},
		{
			ID:   "912",
			Name: "*",
			TTL:  "1800",
			Type: "A",
			Data: "1.2.3.4",
		},
	}

	assert.Equal(t, expected, records)
}

func TestClient_AddRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /",
			servermock.ResponseFromFixture(commandDNSRowAdd+".json"),
			checkFormRequest(`{"request":{"user":"user","auth":"xxx","command":"dns-row-add","data":{"domain":"example.com","name":"foo","ttl":1800,"type":"TXT","rdata":"foobar"}}}`)).
		Build(t)

	record := DNSRow{
		ID:   "",
		Name: "foo",
		TTL:  "1800",
		Type: "TXT",
		Data: "foobar",
	}

	err := client.AddRecord(t.Context(), "example.com.", record)
	require.NoError(t, err)
}

func TestClient_AddRecord_update(t *testing.T) {
	client := mockBuilder().
		Route("POST /",
			servermock.ResponseFromFixture(commandDNSRowUpdate+".json"),
			checkFormRequest(`{"request":{"user":"user","auth":"xxx","command":"dns-row-update","data":{"row_id":"1","domain":"example.com","ttl":1800,"type":"TXT","rdata":"foobar"}}}`)).
		Build(t)

	record := DNSRow{
		ID:   "1",
		Name: "foo",
		TTL:  "1800",
		Type: "TXT",
		Data: "foobar",
	}

	err := client.AddRecord(t.Context(), "example.com.", record)
	require.NoError(t, err)
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /",
			servermock.ResponseFromFixture(commandDNSRowDelete+".json"),
			checkFormRequest(`{"request":{"user":"user","auth":"xxx","command":"dns-row-delete","data":{"row_id":"1","domain":"example.com","rdata":""}}}`)).
		Build(t)

	err := client.DeleteRecord(t.Context(), "example.com.", "1")
	require.NoError(t, err)
}

func TestClient_Commit(t *testing.T) {
	client := mockBuilder().
		Route("POST /",
			servermock.ResponseFromFixture(commandDNSDomainCommit+".json"),
			checkFormRequest(`{"request":{"user":"user","auth":"xxx","command":"dns-domain-commit","data":{"name":"example.com"}}}`)).
		Build(t)

	err := client.Commit(t.Context(), "example.com.")
	require.NoError(t, err)
}

func checkFormRequest(data string) servermock.LinkFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			err := req.ParseForm()
			if err != nil {
				http.Error(rw, err.Error(), http.StatusBadRequest)
				return
			}

			form := regexp.MustCompile(`"auth":"\w+",`).
				ReplaceAllString(req.PostForm.Get("request"), `"auth":"xxx",`)

			if form != data {
				http.Error(rw, fmt.Sprintf("invalid form data: %s", req.PostForm.Get("request")), http.StatusBadRequest)
				return
			}

			next.ServeHTTP(rw, req)
		})
	}
}
