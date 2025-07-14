package internal

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/nrdcg/goacmedns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock.Builder[*HTTPStorage] {
	return servermock.NewBuilder[*HTTPStorage](
		func(server *httptest.Server) (*HTTPStorage, error) {
			storage, err := NewHTTPStorage(server.URL)
			if err != nil {
				return nil, err
			}

			storage.client = server.Client()

			return storage, nil
		},
		servermock.CheckHeader().WithJSONHeaders())
}

func TestHTTPStorage_Fetch(t *testing.T) {
	storage := mockBuilder().
		Route("GET /example.com", servermock.ResponseFromFixture("fetch.json")).
		Build(t)

	account, err := storage.Fetch(t.Context(), "example.com")
	require.NoError(t, err)

	expected := goacmedns.Account{
		FullDomain: "foo.example.com",
		SubDomain:  "foo",
		Username:   "user",
		Password:   "secret",
		ServerURL:  "https://example.com",
	}

	assert.Equal(t, expected, account)
}

func TestHTTPStorage_Fetch_error(t *testing.T) {
	storage := mockBuilder().
		Route("GET /example.com",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusInternalServerError)).
		Build(t)

	_, err := storage.Fetch(t.Context(), "example.com")
	require.Error(t, err)
}

func TestHTTPStorage_FetchAll(t *testing.T) {
	storage := mockBuilder().
		Route("GET /", servermock.ResponseFromFixture("fetch-all.json")).
		Build(t)

	account, err := storage.FetchAll(t.Context())
	require.NoError(t, err)

	expected := map[string]goacmedns.Account{
		"a": {
			FullDomain: "foo.example.com",
			SubDomain:  "foo",
			Username:   "user",
			Password:   "secret",
			ServerURL:  "https://example.com",
		},
		"b": {
			FullDomain: "bar.example.com",
			SubDomain:  "bar",
			Username:   "user",
			Password:   "secret",
			ServerURL:  "https://example.com",
		},
	}

	assert.Equal(t, expected, account)
}

func TestHTTPStorage_FetchAll_error(t *testing.T) {
	storage := mockBuilder().
		Route("GET /",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusInternalServerError)).
		Build(t)

	_, err := storage.FetchAll(t.Context())
	require.Error(t, err)
}

func TestHTTPStorage_Put(t *testing.T) {
	storage := mockBuilder().
		Route("POST /example.com", nil,
			servermock.CheckRequestJSONBodyFromFixture("request-body.json")).
		Build(t)

	account := goacmedns.Account{
		FullDomain: "foo.example.com",
		SubDomain:  "foo",
		Username:   "user",
		Password:   "secret",
		ServerURL:  "https://example.com",
	}

	err := storage.Put(t.Context(), "example.com", account)
	require.NoError(t, err)
}

func TestHTTPStorage_Put_error(t *testing.T) {
	storage := mockBuilder().
		Route("POST /example.com",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusInternalServerError)).
		Build(t)

	account := goacmedns.Account{
		FullDomain: "foo.example.com",
		SubDomain:  "foo",
		Username:   "user",
		Password:   "secret",
		ServerURL:  "https://example.com",
	}

	err := storage.Put(t.Context(), "example.com", account)
	require.Error(t, err)
}

func TestHTTPStorage_Put_CNAME_created(t *testing.T) {
	storage := mockBuilder().
		Route("POST /example.com",
			servermock.Noop().
				WithStatusCode(http.StatusCreated),
			servermock.CheckRequestJSONBodyFromFixture("request-body.json")).
		Build(t)

	account := goacmedns.Account{
		FullDomain: "foo.example.com",
		SubDomain:  "foo",
		Username:   "user",
		Password:   "secret",
		ServerURL:  "https://example.com",
	}

	err := storage.Put(t.Context(), "example.com", account)
	require.ErrorIs(t, err, ErrCNAMEAlreadyCreated)
}
