package internal

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client, err := NewClient("example.com", "user", "secret")
			if err != nil {
				return nil, err
			}

			client.HTTPClient = server.Client()
			client.baseURL, _ = url.Parse(server.URL)

			return client, nil
		},
		servermock.CheckHeader().WithJSONHeaders().
			WithRegexp("Authorization", `Basic .+`).
			WithRegexp("Date", `\d+-\d+-\d+T\d{2}:\d{2}:\d{2}.*`).
			With("Accept-Language", "en_us"))
}

func TestClient_GetServices(t *testing.T) {
	client := mockBuilder().
		Route("GET /v1/user/self/service",
			servermock.ResponseFromFixture("services.json")).
		Build(t)

	services, err := client.GetServices(t.Context())
	require.NoError(t, err)

	expected := []Service{
		{
			ID:          1111,
			ServiceName: ".sk doména",
			Status:      "active",
			Name:        "mydomain.sk",
			CreateTime:  1374357600,
			ExpireTime:  1405914526,
			Price:       12.3,
		},
		{
			ID:          2222,
			ServiceName: "The Hosting",
			Status:      "active",
			Name:        "myname_1",
			CreateTime:  1400145443,
			ExpireTime:  1431702371,
			Price:       55.2,
		},
	}

	assert.Equal(t, expected, services)
}

func TestClient_GetServices_errors(t *testing.T) {
	client := mockBuilder().
		Route("GET /v1/user/self/service",
			servermock.ResponseFromFixture("error_v1.json").
				WithStatusCode(http.StatusUnauthorized)).
		Build(t)

	_, err := client.GetServices(t.Context())
	require.EqualError(t, err, "401: No username or password.")
}

func TestClient_GetRecords(t *testing.T) {
	client := mockBuilder().
		Route("GET /v2/service/aaa/dns/record",
			servermock.ResponseFromFixture("records.json")).
		Build(t)

	filter := RecordFilter{
		Name:    "example.com",
		Type:    []string{"TXT"},
		Content: "txt",
	}

	records, err := client.GetRecords(t.Context(), "aaa", filter)
	require.NoError(t, err)

	expected := []Record{{
		ID:       13,
		Name:     "string",
		Content:  "string",
		TTL:      120,
		Priority: 1,
		Port:     443,
		Weight:   50,
	}}

	assert.Equal(t, expected, records)
}

func TestClient_GetRecords_errors(t *testing.T) {
	client := mockBuilder().
		Route("GET /v2/service/aaa/dns/record",
			servermock.ResponseFromFixture("error_403.json").
				WithStatusCode(http.StatusForbidden)).
		Build(t)

	filter := RecordFilter{
		Name:    "example.com",
		Type:    []string{"TXT"},
		Content: "txt",
	}

	_, err := client.GetRecords(t.Context(), "aaa", filter)
	require.EqualError(t, err, "403: /errors/httpException: This action is unauthorized.")
}

func TestClient_CreateRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /v2/service/aaa/dns/record",
			servermock.Noop().
				WithStatusCode(http.StatusNoContent)).
		Build(t)

	err := client.CreateRecord(t.Context(), "aaa", Record{})
	require.NoError(t, err)
}

func TestClient_CreateRecord_errors(t *testing.T) {
	client := mockBuilder().
		Route("POST /v2/service/aaa/dns/record",
			servermock.ResponseFromFixture("error_403.json").
				WithStatusCode(http.StatusForbidden)).
		Build(t)

	err := client.CreateRecord(t.Context(), "aaa", Record{})
	require.EqualError(t, err, "403: /errors/httpException: This action is unauthorized.")
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /v2/service/aaa/dns/record/123",
			servermock.Noop().
				WithStatusCode(http.StatusNoContent)).
		Build(t)

	err := client.DeleteRecord(t.Context(), "aaa", "123")
	require.NoError(t, err)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /v2/service/aaa/dns/record/123",
			servermock.ResponseFromFixture("error_403.json").
				WithStatusCode(http.StatusForbidden)).
		Build(t)

	err := client.DeleteRecord(t.Context(), "aaa", "123")
	require.EqualError(t, err, "403: /errors/httpException: This action is unauthorized.")
}

func TestClient_sign(t *testing.T) {
	client, err := NewClient("example.com", "user", "secret")
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodGet, "/v1/user/self/service", nil)
	require.NoError(t, err)

	err = client.sign(req, time.Date(2025, 6, 28, 1, 2, 3, 4, time.UTC))
	require.NoError(t, err)

	username, password, ok := req.BasicAuth()
	require.True(t, ok)

	assert.Equal(t, "user", username)
	assert.Equal(t, "743e2257421b260ed561f3e7af4b035414636393", password)
}
