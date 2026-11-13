package internal

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client, err := NewClient()
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

func TestClient_GetZoneByName(t *testing.T) {
	client := mockBuilder().
		Route("POST /enum.api.v1.DnsService/GetZoneByName",
			servermock.ResponseFromFixture("GetZoneByName.json"),
			servermock.CheckRequestJSONBodyFromFixture("GetZoneByName-request.json"),
		).
		Build(t)

	in := &GetZoneByNameRequest{
		ProjectID: "projectA",
		Name:      "example.com",
	}

	result, err := client.GetZoneByName(mockContext(t), in)
	require.NoError(t, err)

	expected := &GetZoneByNameResponse{
		Zone: &DNSZone{
			ID:        "zoneA",
			ProjectID: "projectA",
			Name:      "example.com",
			Status:    "RESOURCE_STATUS_READY",
		},
	}

	assert.Equal(t, expected, result)
}

func TestClient_GetZoneByName_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /enum.api.v1.DnsService/GetZoneByName",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusNotFound),
		).
		Build(t)

	in := &GetZoneByNameRequest{
		ProjectID: "projectA",
		Name:      "example.com",
	}

	_, err := client.GetZoneByName(mockContext(t), in)
	require.EqualError(t, err, "not_found: Not found")
}

func TestClient_ListProjects(t *testing.T) {
	client := mockBuilder().
		Route("POST /enum.api.v1.ProjectService/ListProjects",
			servermock.ResponseFromFixture("ListProjects.json"),
			servermock.CheckRequestJSONBody(`{}`),
		).
		Build(t)

	result, err := client.ListProjects(mockContext(t))
	require.NoError(t, err)

	expected := &ListProjectsResponse{
		Projects: []*Project{{
			ID:          "projectA",
			OrgID:       "orgA",
			Name:        "foo",
			Description: "bar",
		}},
		Page: &PageResponse{
			NextPageToken: "azerty",
			TotalCount:    1,
		},
	}

	assert.Equal(t, expected, result)
}

func TestClient_AddRecordSetValue(t *testing.T) {
	client := mockBuilder().
		Route("POST /enum.api.v1.DnsService/AddRecordSetValue",
			servermock.ResponseFromFixture("AddRecordSetValue.json"),
			servermock.CheckRequestJSONBodyFromFixture("AddRecordSetValue-request.json"),
		).
		Build(t)

	in := &AddRecordSetValueRequest{
		ProjectID: "projectA",
		ZoneID:    "zoneA",
		Name:      "_acme-challenge.example.com",
		Type:      "TXT",
		TTL:       120,
		Value: &DNSRecordValue{
			Content: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		},
	}

	result, err := client.AddRecordSetValue(mockContext(t), in)
	require.NoError(t, err)

	expected := &RecordSetResponse{
		RecordSet: &DNSRecordSet{
			ID:     "rrsetA",
			ZoneID: "zoneA",
			Name:   "_acme-challenge.example.com",
			Type:   "TXT",
			TTL:    120,
			Records: []*DNSRecordValue{{
				Content: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
			}},
		},
	}

	assert.Equal(t, expected, result)
}

func TestClient_RemoveRecordSetValue(t *testing.T) {
	client := mockBuilder().
		Route("POST /enum.api.v1.DnsService/RemoveRecordSetValue",
			servermock.ResponseFromFixture("RemoveRecordSetValue.json"),
			servermock.CheckRequestJSONBodyFromFixture("RemoveRecordSetValue-request.json"),
		).
		Build(t)

	in := &RemoveRecordSetValueRequest{
		ProjectID: "projectA",
		ZoneID:    "zoneA",
		Name:      "_acme-challenge.example.com",
		Type:      "TXT",
		Content:   "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
	}

	result, err := client.RemoveRecordSetValue(mockContext(t), in)
	require.NoError(t, err)

	expected := &RecordSetResponse{
		RecordSet: &DNSRecordSet{
			ID:     "rrsetA",
			ZoneID: "zoneA",
			Name:   "_acme-challenge.example.com",
			Type:   "TXT",
			TTL:    120,
			Records: []*DNSRecordValue{{
				Content: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
			}},
		},
	}

	assert.Equal(t, expected, result)
}
