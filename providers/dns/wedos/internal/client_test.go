package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/go-acme/lego/v5/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client := NewClient("user", "secret")
			client.BaseURL = server.URL
			client.HTTPClient = server.Client()

			return client, nil
		},
		servermock.CheckHeader().
			WithContentTypeFromURLEncoded(),
	)
}

func TestClient_GetRecords(t *testing.T) {
	client := mockBuilder().
		Route("POST /",
			servermock.ResponseFromFixture(CommandDNSRowsList+".json"),
			checkFormRequest(CommandDNSRowsList+"_request.json"),
		).
		Build(t)

	records, err := client.GetRecords(t.Context(), "example.com.")
	require.NoError(t, err)

	assert.Len(t, records, 4)

	expected := []Record{
		{
			ID:   "911",
			TTL:  "300",
			Type: "A",
			Data: "1.2.3.4",
		},
		{
			ID:   "913",
			TTL:  "300",
			Type: "MX",
			Data: "1 mail1.wedos.net",
		},
		{
			ID:   "914",
			TTL:  "300",
			Type: "MX",
			Data: "10 mailbackup.wedos.net",
		},
		{
			ID:   "912",
			Name: "*",
			TTL:  "300",
			Type: "A",
			Data: "1.2.3.4",
		},
	}

	assert.Equal(t, expected, records)
}

func TestClient_AddRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /",
			servermock.ResponseFromFixture(CommandDNSRowAdd+".json"),
			checkFormRequest(CommandDNSRowAdd+"_request.json"),
		).
		Build(t)

	record := Record{
		Name: "_acme-challenge",
		TTL:  "300",
		Type: "TXT",
		Data: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
	}

	err := client.AddRecord(t.Context(), "example.com.", record)
	require.NoError(t, err)
}

func TestClient_AddRecord_update(t *testing.T) {
	client := mockBuilder().
		Route("POST /",
			servermock.ResponseFromFixture(CommandDNSRowUpdate+".json"),
			checkFormRequest(CommandDNSRowUpdate+"_request.json"),
		).
		Build(t)

	record := Record{
		ID:   "1",
		Name: "foo",
		TTL:  "300",
		Type: "TXT",
		Data: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
	}

	err := client.AddRecord(t.Context(), "example.com.", record)
	require.NoError(t, err)
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /",
			servermock.ResponseFromFixture(CommandDNSRowDelete+".json"),
			checkFormRequest(CommandDNSRowDelete+"_request.json"),
		).
		Build(t)

	err := client.DeleteRecord(t.Context(), "example.com.", "1")
	require.NoError(t, err)
}

func TestClient_Commit(t *testing.T) {
	client := mockBuilder().
		Route("POST /",
			servermock.ResponseFromFixture(CommandDNSDomainCommit+".json"),
			checkFormRequest(CommandDNSDomainCommit+"_request.json"),
		).
		Build(t)

	err := client.Commit(t.Context(), "example.com.")
	require.NoError(t, err)
}

func checkFormRequest(filename string) servermock.LinkFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			file, err := os.ReadFile(filepath.Join("fixtures", filename))
			if err != nil {
				http.Error(rw, err.Error(), http.StatusBadRequest)
				return
			}

			var expected bytes.Buffer

			err = json.Compact(&expected, file)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusBadRequest)
				return
			}

			err = req.ParseForm()
			if err != nil {
				http.Error(rw, err.Error(), http.StatusBadRequest)
				return
			}

			form := regexp.MustCompile(`"auth":"\w+",`).
				ReplaceAllString(req.PostForm.Get("request"), `"auth":"xxx",`)

			if form != expected.String() {
				log.Error("invalid form data", slog.String("form", form), slog.String("expected", expected.String()))
				http.Error(rw, fmt.Sprintf("invalid form data: %s", req.PostForm.Get("request")), http.StatusBadRequest)

				return
			}

			next.ServeHTTP(rw, req)
		})
	}
}
