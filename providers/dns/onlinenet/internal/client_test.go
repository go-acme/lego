package internal

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

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
			WithAuthorization("Bearer secret"),
	)
}

func TestClient_GetZoneVersion(t *testing.T) {
	client := mockBuilder().
		Route("GET /domain/example.com/version/active",
			servermock.ResponseFromFixture("get_zone_version.json"),
		).
		Build(t)

	zoneVersion, err := client.GetZoneVersion(t.Context(), "example.com", "active")
	require.NoError(t, err)

	expected := &ZoneVersion{
		UUIDRef:      "6200f3ee-5648-4d69-a846-03ac88121660",
		Name:         "example.com",
		CreationDate: "2026-04-04",
		Active:       false,
		Zone:         &Reference{Ref: "c3d5c0bb-5094-4974-b54b-16175207ac91"},
		Domain:       &Reference{Ref: "a7362c44-867a-4dea-89ad-1a1d5efc8e7a"},
	}

	assert.Equal(t, expected, zoneVersion)
}

func TestClient_CreateZoneVersion(t *testing.T) {
	client := mockBuilder().
		Route("POST /domain/example.com/version",
			servermock.ResponseFromFixture("create_zone_version.json").
				WithStatusCode(http.StatusCreated),
			servermock.CheckForm().Strict().
				With("name", "lego"),
		).
		Build(t)

	zoneVersion, err := client.CreateZoneVersion(t.Context(), "example.com", "lego")
	require.NoError(t, err)

	expected := &ZoneVersion{
		UUIDRef:      "9335be4a-063c-43d6-a393-8bd5d7c78f07",
		Name:         "example.com",
		CreationDate: "2026-04-04",
		Active:       false,
		Zone:         &Reference{Ref: "c3d5c0bb-5094-4974-b54b-16175207ac91"},
		Domain:       &Reference{Ref: "a7362c44-867a-4dea-89ad-1a1d5efc8e7a"},
	}

	assert.Equal(t, expected, zoneVersion)
}

func TestClient_DeleteZoneVersion(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /domain/example.com/version/9335be4a-063c-43d6-a393-8bd5d7c78f07",
			servermock.Noop().
				WithStatusCode(http.StatusNoContent),
		).
		Build(t)

	err := client.DeleteZoneVersion(t.Context(), "example.com", "9335be4a-063c-43d6-a393-8bd5d7c78f07")
	require.NoError(t, err)
}

func TestClient_EditActiveZoneVersion(t *testing.T) {
	client := mockBuilder().
		Route("PATCH /domain/example.com/version/active",
			servermock.Noop().
				WithStatusCode(http.StatusCreated),
			servermock.CheckRequestJSONBodyFromFixture("edit_active_zone-request_add.json"),
		).
		Build(t)

	operation := ResourceRecordOperation{
		Name:       "_acme-challenge",
		Type:       "TXT",
		ChangeType: ChangeTypeAdd,
		Records: []Record{{
			Name: "_acme-challenge",
			Type: "TXT",
			TTL:  120,
			Data: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		}},
	}

	err := client.EditActiveZoneVersion(t.Context(), "example.com", []ResourceRecordOperation{operation})
	require.NoError(t, err)
}

func TestClient_EnableZoneVersion(t *testing.T) {
	client := mockBuilder().
		Route("PATCH /domain/example.com/version/9335be4a-063c-43d6-a393-8bd5d7c78f07/enable",
			servermock.Noop().
				WithStatusCode(http.StatusNoContent),
		).
		Build(t)

	err := client.EnableZoneVersion(t.Context(), "example.com", "9335be4a-063c-43d6-a393-8bd5d7c78f07")
	require.NoError(t, err)
}

func TestClient_GetActiveZone(t *testing.T) {
	client := mockBuilder().
		Route("GET /domain/example.com/zone",
			servermock.ResponseFromFixture("get_active_zone.json"),
		).
		Build(t)

	result, err := client.GetActiveZone(t.Context(), "example.com")
	require.NoError(t, err)

	expected := []ResourceRecord{
		{
			ID:   465,
			Name: "@",
			Type: "A",
			Aux:  0,
			TTL:  300,
			Data: "192.168.0.1",
			Domain: &Domain{
				ID:       2,
				Name:     "example.com",
				DNSSec:   true,
				External: true,
				Versions: &Reference{Ref: "string"},
				Zone:     &Reference{Ref: "string"},
			},
		},
		{
			ID:   123,
			Name: "_acme-challenge",
			Type: "TXT",
			Aux:  0,
			TTL:  120,
			Data: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
			Domain: &Domain{
				ID:       2,
				Name:     "example.com",
				DNSSec:   true,
				External: true,
				Versions: &Reference{Ref: "string"},
				Zone:     &Reference{Ref: "string"},
			},
		},
	}

	assert.Equal(t, expected, result)
}

func TestClient_CreateResourceRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /domain/example.com/version/9335be4a-063c-43d6-a393-8bd5d7c78f07/zone",
			servermock.ResponseFromFixture("create_resource_record.json").
				WithStatusCode(http.StatusCreated),
			servermock.CheckForm().Strict().
				With("name", "_acme-challenge").
				With("type", "TXT").
				With("ttl", "120").
				With("data", "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY").
				With("priority", "0"),
		).
		Build(t)

	record := RecordRequest{
		Name: "_acme-challenge",
		Type: "TXT",
		TTL:  120,
		Data: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
	}

	result, err := client.CreateResourceRecord(t.Context(), "example.com", "9335be4a-063c-43d6-a393-8bd5d7c78f07", record)
	require.NoError(t, err)

	expected := &ResourceRecord{
		ID:   123,
		Name: "_acme-challenge",
		Type: "TXT",
		TTL:  120,
		Data: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		Domain: &Domain{
			ID:       2,
			Name:     "example.com",
			DNSSec:   true,
			External: true,
			Versions: &Reference{Ref: "string"},
			Zone:     &Reference{Ref: "string"},
		},
	}

	assert.Equal(t, expected, result)
}

func TestClient_DeleteResourceRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /domain/example.com/version/9335be4a-063c-43d6-a393-8bd5d7c78f07/zone/123",
			servermock.Noop().
				WithStatusCode(http.StatusNoContent),
		).
		Build(t)

	err := client.DeleteResourceRecord(t.Context(), "example.com", "9335be4a-063c-43d6-a393-8bd5d7c78f07", "123")
	require.NoError(t, err)
}
