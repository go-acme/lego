package internal

import (
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
			WithJSONHeaders().
			WithAuthorization("Bearer secret"),
	)
}

func TestClient_GetDomains(t *testing.T) {
	client := mockBuilder().
		Route("GET /dns",
			servermock.ResponseFromFixture("domainlist.json"),
		).
		Build(t)

	domains, err := client.GetDomains(t.Context())
	require.NoError(t, err)

	expected := []Domain{
		{
			ID:         1,
			Name:       "example.com",
			TechName:   "example_com",
			State:      "New",
			IsDelegate: true,
			LinkedRecords: []Record{{
				ID:                   1,
				Type:                 "A",
				IP:                   "1.1.1.1",
				HostName:             "@",
				TTL:                  3600,
				CanonicalDescription: "example.com 3600 IN A 1.1.1.1",
			}},
		},
		{
			ID:         2,
			Name:       "test2.example",
			TechName:   "test2_example",
			State:      "Active",
			IsDelegate: true,
			LinkedRecords: []Record{{
				ID:                   2,
				Type:                 "A",
				IP:                   "1.1.1.2",
				HostName:             "@",
				TTL:                  3600,
				CanonicalDescription: "test2.example 3600 IN A 1.1.1.2",
			}},
		},
	}

	assert.Equal(t, expected, domains)
}

func TestClient_CreateTXTRecord(t *testing.T) {
	client := mockBuilder().
		Route("POST /dns/recordtxt",
			servermock.ResponseFromFixture("create_record_txt.json"),
			servermock.CheckRequestJSONBodyFromFixture("create_record_txt-request.json"),
		).
		Build(t)

	ctrr := CreateTXTRecordRequest{
		DomainID: "1",
		Name:     "_acme-challenge",
		TTL:      "300",
		Text:     "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
	}

	domain, err := client.CreateTXTRecord(t.Context(), ctrr)
	require.NoError(t, err)

	expected := &Record{
		ID:                   9,
		Type:                 "TXT",
		HostName:             "_acme-challenge.test.example.",
		Text:                 "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
		TTL:                  300,
		CanonicalDescription: "_acme-challenge.test.example. 3600 IN TXT ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
	}

	assert.Equal(t, expected, domain)
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /dns/1/9",
			servermock.Noop(),
		).
		Build(t)

	err := client.DeleteRecord(t.Context(), 1, 9)
	require.NoError(t, err)
}
