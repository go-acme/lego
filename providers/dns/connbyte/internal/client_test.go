package internal

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client, err := NewClient("secret")
			if err != nil {
				return nil, err
			}

			client.BaseURL, _ = url.Parse(server.URL)
			client.HTTPClient = server.Client()

			return client, nil
		},
		servermock.CheckHeader().
			WithJSONHeaders().
			WithAuthorization("Bearer secret"),
	)
}

func TestClient_ListZones(t *testing.T) {
	client := mockBuilder().
		Route("POST /ns/zone/list",
			servermock.ResponseFromFixture("zone_list.json"),
		).
		Build(t)

	result, err := client.ListZones(t.Context())
	require.NoError(t, err)

	expected := []ZoneListResponse{{
		Zone: Zone{
			ID:        "01DXF6DT00ZZZHQPP3JFJS8VEE",
			ProjectID: "01DXF6DT00ZZZHQPP3JFJS8VEE",
			Zone:      "example.com",
			CreatedAt: time.Date(2026, time.April, 5, 21, 43, 40, 0, time.UTC),
		},
	}}

	require.Equal(t, expected, result)
}

func TestClient_ListZones_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /ns/zone/list",
			servermock.ResponseFromFixture("error-bad_request.json").
				WithStatusCode(http.StatusBadRequest),
		).
		Build(t)

	_, err := client.ListZones(t.Context())

	require.EqualError(t, err, "400: Bad Request Transferred data are incorrect."+
		" (01DXF6DT009CJANMTD1S57HQWQ 1777992035 /ns/zone/list)"+
		" [authorization: 'Authorization' must be set.]"+
		" [projectId: Field must be set. Field cannot be empty.]")
}

func TestClient_CreateRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /ns/record/create",
			servermock.ResponseFromFixture("record_create.json"),
			servermock.CheckRequestJSONBodyFromFixture("record_create-request.json"),
		).
		Build(t)

	record := Record{
		ZoneID:  "01DXF6DT00ZZZHQPP3JFJS8VEE",
		Name:    "_acme-challenge",
		Type:    "TXT",
		Content: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		TTL:     120,
	}

	result, err := client.CreateRecord(t.Context(), record)
	require.NoError(t, err)

	expected := &RecordCreateResponse{
		ID:     435,
		Record: "_acme-challenge.example.com",
	}

	require.Equal(t, expected, result)
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /ns/record/delete",
			servermock.ResponseFromFixture("record_delete.json"),
			servermock.CheckRequestJSONBodyFromFixture("record_delete-request.json"),
		).
		Build(t)

	err := client.DeleteRecord(t.Context(), "01DXF6DT00ZZZHQPP3JFJS8VEE", 435)
	require.NoError(t, err)
}
