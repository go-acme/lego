package internal

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/go-acme/lego/v4/providers/dns/internal/ptr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client, err := NewClient("abc", "secret")
			if err != nil {
				return nil, err
			}

			client.HTTPClient = server.Client()
			client.BaseURL, _ = url.Parse(server.URL)

			return client, nil
		},
		servermock.CheckHeader().
			WithJSONHeaders().
			With(HeaderAppID, "abc").
			With(HeaderAPIKey, "secret"),
	)
}

func TestClient_GetZone(t *testing.T) {
	client := mockBuilder().
		Route("GET /example.com",
			servermock.ResponseFromFixture("get_zone.json"),
		).
		Build(t)

	zone, err := client.GetZone(context.Background(), "example.com")
	require.NoError(t, err)

	assert.Equal(t, expectedZone(), zone)
}

func TestClient_GetZone_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /example.com",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized),
		).
		Build(t)

	_, err := client.GetZone(context.Background(), "example.com")
	require.Error(t, err)

	require.EqualError(t, err, "401: INVALID_API_KEY: Invalid API Key")
}

func TestClient_SaveZone(t *testing.T) {
	client := mockBuilder().
		Route("PUT /example.com",
			servermock.Noop().
				WithStatusCode(http.StatusNoContent),
			servermock.CheckRequestJSONBodyFromFixture("zones_add.json"),
		).
		Build(t)

	record := Record{
		Type:  "TXT",
		Host:  "_acme-challenge",
		RData: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		TTL:   600,
	}

	err := client.SaveZone(context.Background(), "example.com", expectedZone(record))
	require.NoError(t, err)
}

func TestClient_SaveZone_error(t *testing.T) {
	client := mockBuilder().
		Route("PUT /example.com",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized),
		).
		Build(t)

	err := client.SaveZone(context.Background(), "example.com", expectedZone())
	require.Error(t, err)

	require.EqualError(t, err, "401: INVALID_API_KEY: Invalid API Key")
}

func TestClient_ValidateZone(t *testing.T) {
	client := mockBuilder().
		Route("POST /example.com/check",
			servermock.ResponseFromFixture("zones_add.json"),
			servermock.CheckRequestJSONBodyFromFixture("zones_add.json"),
		).
		Build(t)

	record := Record{
		Type:  "TXT",
		Host:  "_acme-challenge",
		RData: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		TTL:   600,
	}

	zone, err := client.ValidateZone(context.Background(), "example.com", expectedZone(record))
	require.NoError(t, err)

	assert.Equal(t, expectedZone(record), zone)
}

func TestClient_ValidateZone_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /example.com/check",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized),
		).
		Build(t)

	_, err := client.ValidateZone(context.Background(), "example.com", expectedZone())
	require.Error(t, err)

	require.EqualError(t, err, "401: INVALID_API_KEY: Invalid API Key")
}

func expectedZone(records ...Record) *Zone {
	rs := []Record{{
		Type:     "A",
		Host:     "string",
		RData:    "string",
		Updated:  true,
		Locked:   true,
		IsDynDNS: ptr.Pointer(true),
		Proxy:    "ON",
	}}

	rs = append(rs, records...)

	return &Zone{
		Name:          "string",
		DomainConnect: true,
		Records:       rs,
		URLForwards: []URLForward{{
			ForwardType: "FRAME",
			Host:        "string",
			URL:         "string",
			Title:       "string",
			Keywords:    "string",
			Description: "string",
			Updated:     true,
		}},
		MailForwards: []MailForward{{
			Source:      "string",
			Destination: "string",
			Updated:     true,
		}},
		Report: &Report{
			IsValid: true,
			URLForwardErrors: []URLForwardError{{
				Messages: []string{"string"},
				Severity: "ERROR",
				URLForward: &URLForward{
					ForwardType: "FRAME",
					Host:        "string",
					URL:         "string",
					Title:       "string",
					Keywords:    "string",
					Description: "string",
					Updated:     true,
				},
			}},
			MailForwardErrors: []MailForwardError{{
				Messages: []string{"string"},
				MailForward: &MailForward{
					Source:      "string",
					Destination: "string",
					Updated:     true,
				},
				Severity: "ERROR",
			}},
			ZoneErrors: []ZoneError{{
				Message:  "string",
				Severity: "ERROR",
				Records: []Record{{
					Type:     "A",
					Host:     "string",
					RData:    "string",
					Updated:  true,
					Locked:   true,
					IsDynDNS: ptr.Pointer(true),
					Proxy:    "ON",
				}},
				URLForwards: []URLForward{{
					ForwardType: "FRAME",
					Host:        "string",
					URL:         "string",
					Title:       "string",
					Keywords:    "string",
					Description: "string",
					Updated:     true,
				}},
				MailForwards: []MailForward{{
					Source:      "string",
					Destination: "string",
					Updated:     true,
				}},
			}},
		},
	}
}
