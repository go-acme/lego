package internal

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
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
			servermock.ResponseFromFixture("zone_get.json"),
		).
		Build(t)

	zone, err := client.GetZone(context.Background(), "example.com")
	require.NoError(t, err)

	expected := &Zone{
		Name:          "example.com",
		DomainConnect: true,
		Records:       slices.Concat([]Record{fakeARecord()}),
		URLForwards:   []URLForward{fakeURLForward()},
		MailForwards:  []MailForward{fakeMailForward()},
	}

	assert.Equal(t, expected, zone)
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
			servermock.CheckRequestJSONBodyFromFixture("zone_add.json"),
		).
		Build(t)

	record := Record{
		Type:  "TXT",
		Host:  "_acme-challenge",
		RData: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		TTL:   600,
	}

	zone := &Zone{
		Name:          "example.com",
		DomainConnect: true,
		Records:       []Record{fakeARecord(), record},
		URLForwards:   []URLForward{fakeURLForward()},
		MailForwards:  []MailForward{fakeMailForward()},
	}

	err := client.SaveZone(context.Background(), "example.com", zone)
	require.NoError(t, err)
}

func TestClient_SaveZone_emptyForwards(t *testing.T) {
	client := mockBuilder().
		Route("PUT /example.com",
			servermock.Noop().
				WithStatusCode(http.StatusNoContent),
			servermock.CheckRequestJSONBodyFromFixture("zone_add_empty_forwards.json"),
		).
		Build(t)

	record := Record{
		Type:  "TXT",
		Host:  "_acme-challenge",
		RData: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		TTL:   600,
	}

	zone := &Zone{
		Name:          "example.com",
		DomainConnect: true,
		Records:       slices.Concat([]Record{fakeARecord(), record}),
	}

	err := client.SaveZone(context.Background(), "example.com", zone)
	require.NoError(t, err)
}

func TestClient_SaveZone_error(t *testing.T) {
	client := mockBuilder().
		Route("PUT /example.com",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized),
		).
		Build(t)

	zone := &Zone{
		Name:          "example.com",
		DomainConnect: true,
		Records:       []Record{fakeARecord()},
		URLForwards:   []URLForward{fakeURLForward()},
		MailForwards:  []MailForward{fakeMailForward()},
	}

	err := client.SaveZone(context.Background(), "example.com", zone)
	require.Error(t, err)

	require.EqualError(t, err, "401: INVALID_API_KEY: Invalid API Key")
}

func TestClient_ValidateZone(t *testing.T) {
	client := mockBuilder().
		Route("POST /example.com/check",
			servermock.ResponseFromFixture("zone_add_validate_ok.json"),
			servermock.CheckRequestJSONBodyFromFixture("zone_add.json"),
		).
		Build(t)

	record := Record{
		Type:  "TXT",
		Host:  "_acme-challenge",
		RData: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		TTL:   600,
	}

	zone := &Zone{
		Name:          "example.com",
		DomainConnect: true,
		Records:       []Record{fakeARecord(), record},
		URLForwards:   []URLForward{fakeURLForward()},
		MailForwards:  []MailForward{fakeMailForward()},
	}

	zone, err := client.ValidateZone(context.Background(), "example.com", zone)
	require.NoError(t, err)

	expected := &Zone{
		Name:          "example.com",
		DomainConnect: true,
		Records:       []Record{fakeARecord(), record},
		URLForwards:   []URLForward{fakeURLForward()},
		MailForwards:  []MailForward{fakeMailForward()},
		Report:        &Report{IsValid: true},
	}

	assert.Equal(t, expected, zone)
}

func TestClient_ValidateZone_report(t *testing.T) {
	client := mockBuilder().
		Route("POST /example.com/check",
			servermock.ResponseFromFixture("zone_add_validate_ko.json"),
			servermock.CheckRequestJSONBodyFromFixture("zone_add.json"),
		).
		Build(t)

	record := Record{
		Type:  "TXT",
		Host:  "_acme-challenge",
		RData: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		TTL:   600,
	}

	zone := &Zone{
		Name:          "example.com",
		DomainConnect: true,
		Records:       []Record{fakeARecord(), record},
		URLForwards:   []URLForward{fakeURLForward()},
		MailForwards:  []MailForward{fakeMailForward()},
	}

	zone, err := client.ValidateZone(context.Background(), "example.com", zone)
	require.NoError(t, err)

	expected := &Zone{
		Name:          "example.com",
		DomainConnect: true,
		Records:       []Record{fakeARecord(), record},
		URLForwards:   []URLForward{fakeURLForward()},
		MailForwards:  []MailForward{fakeMailForward()},
		Report:        fakeReport(),
	}

	assert.EqualError(t, zone.Report, `record error (ERROR): "120" is not a valid TTL, URL forward error (ERROR): string, mail forward error (ERROR): string, zone error (ERROR): string`)

	assert.Equal(t, expected, zone)
}

func TestClient_ValidateZone_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /example.com/check",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusUnauthorized),
		).
		Build(t)

	zone := &Zone{
		Name:          "example.com",
		DomainConnect: true,
		Records:       []Record{fakeARecord()},
		URLForwards:   []URLForward{fakeURLForward()},
		MailForwards:  []MailForward{fakeMailForward()},
	}

	_, err := client.ValidateZone(context.Background(), "example.com", zone)
	require.Error(t, err)

	require.EqualError(t, err, "401: INVALID_API_KEY: Invalid API Key")
}

func fakeARecord() Record {
	return Record{
		ID:       1000,
		Type:     "A",
		Host:     "@",
		TTL:      600,
		RData:    "string",
		Updated:  ptr.Pointer(true),
		Locked:   ptr.Pointer(true),
		IsDynDNS: ptr.Pointer(true),
		Proxy:    "ON",
	}
}

func fakeURLForward() URLForward {
	return URLForward{
		ID:          2000,
		ForwardType: "FRAME",
		Host:        "string",
		URL:         "string",
		Title:       "string",
		Keywords:    "string",
		Description: "string",
		Updated:     ptr.Pointer(true),
	}
}

func fakeMailForward() MailForward {
	return MailForward{
		ID:          3000,
		Source:      "string",
		Destination: "string",
		Updated:     ptr.Pointer(true),
	}
}

func fakeReport() *Report {
	return &Report{
		IsValid: false,
		RecordErrors: []RecordError{{
			Messages: []string{`"120" is not a valid TTL`},
			Severity: "ERROR",
			Record:   fakeARecord(),
		}},
		URLForwardErrors: []URLForwardError{{
			Messages:   []string{"string"},
			Severity:   "ERROR",
			URLForward: fakeURLForward(),
		}},
		MailForwardErrors: []MailForwardError{{
			Messages:    []string{"string"},
			MailForward: fakeMailForward(),
			Severity:    "ERROR",
		}},
		ZoneErrors: []ZoneError{{
			Message:      "string",
			Severity:     "ERROR",
			Records:      []Record{fakeARecord()},
			URLForwards:  []URLForward{fakeURLForward()},
			MailForwards: []MailForward{fakeMailForward()},
		}},
	}
}
