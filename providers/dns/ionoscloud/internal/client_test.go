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

func TestClient_RetrieveZones(t *testing.T) {
	client := mockBuilder().
		Route("GET /zones",
			servermock.ResponseFromFixture("zones.json"),
			servermock.CheckQueryParameter().Strict().
				With("filter.zoneName", "example.com")).
		Build(t)

	zones, err := client.RetrieveZones(t.Context(), "example.com")
	require.NoError(t, err)

	expected := []Zone{{
		ID:   "e74d0d15-f567-4b7b-9069-26ee1f93bae3",
		Type: "zone",
		Metadata: ZoneMetadata{
			CreatedDate:          time.Date(2022, time.August, 21, 15, 52, 53, 0, time.UTC),
			CreatedBy:            "ionos:iam:cloud:31960002:users/87f9a82e-b28d-49ed-9d04-fba2c0459cd3",
			CreatedByUserID:      "87f9a82e-b28d-49ed-9d04-fba2c0459cd3",
			LastModifiedDate:     time.Date(2022, time.August, 21, 15, 52, 53, 0, time.UTC),
			LastModifiedBy:       "ionos:iam:cloud:31960002:users/87f9a82e-b28d-49ed-9d04-fba2c0459cd3",
			LastModifiedByUserID: "63cef532-26fe-4a64-a4e0-de7c8a506c90",
			ResourceURN:          "ionos:<product>:<location>:<contract>:<resource-path>",
			State:                "PROVISIONING",
			Nameservers:          []string{"ns-ic.ui-dns.com", "ns-ic.ui-dns.de", "ns-ic.ui-dns.org", "ns-ic.ui-dns.biz"},
		},
		Properties: ZoneProperties{
			ZoneName:    "example.com",
			Description: "The hosted zone is used for example.com",
			Enabled:     true,
		},
	}}

	assert.Equal(t, expected, zones)
}

func TestClient_RetrieveZones_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /zones",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized)).
		Build(t)

	_, err := client.RetrieveZones(t.Context(), "example.com")
	require.EqualError(t, err, "401: paas-auth-1: Unauthorized, wrong or no api key provided to process this request")
}

func TestClient_CreateRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /zones/abc/records",
			servermock.ResponseFromFixture("create_record.json"),
			servermock.CheckRequestJSONBodyFromFixture("create_record-request.json")).
		Build(t)

	record := RecordProperties{
		Name:    "_acme-challenge",
		Type:    "TXT",
		Content: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		TTL:     120,
	}

	result, err := client.CreateRecord(t.Context(), "abc", record)
	require.NoError(t, err)

	expected := &RecordResponse{
		ID:   "90d81ac0-3a30-44d4-95a5-12959effa6ee",
		Type: "record",
		Metadata: RecordMetadata{
			CreatedDate:          time.Date(2022, time.August, 21, 15, 52, 53, 0, time.UTC),
			CreatedBy:            "ionos:iam:cloud:31960002:users/87f9a82e-b28d-49ed-9d04-fba2c0459cd3",
			CreatedByUserID:      "87f9a82e-b28d-49ed-9d04-fba2c0459cd3",
			LastModifiedDate:     time.Date(2022, time.August, 21, 15, 52, 53, 0, time.UTC),
			LastModifiedBy:       "ionos:iam:cloud:31960002:users/87f9a82e-b28d-49ed-9d04-fba2c0459cd3",
			LastModifiedByUserID: "63cef532-26fe-4a64-a4e0-de7c8a506c90",
			ResourceURN:          "ionos:<product>:<location>:<contract>:<resource-path>",
			State:                "PROVISIONING",
			Fqdn:                 "app.example.com",
			ZoneID:               "a363f30c-4c0c-4552-9a07-298d87f219bf",
		},
		Properties: RecordProperties{
			Name:     "app",
			Type:     "A",
			Content:  "1.2.3.4",
			TTL:      3600,
			Priority: 3600,
			Enabled:  true,
		},
	}

	assert.Equal(t, expected, result)
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /zones/abc/records/def",
			servermock.Noop().
				WithStatusCode(http.StatusAccepted)).
		Build(t)

	err := client.DeleteRecord(t.Context(), "abc", "def")
	require.NoError(t, err)
}
