package internal

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client, err := NewClient("key", "secret")
			if err != nil {
				return nil, err
			}

			client.HTTPClient = server.Client()
			client.BaseURL, _ = url.Parse(server.URL)

			return client, nil
		},
		servermock.CheckHeader().
			WithJSONHeaders().
			With("Authorization", "Bearer key").
			WithRegexp("X-Signature", `.+`),
	)
}

func TestClient_SetDNS(t *testing.T) {
	client := mockBuilder().
		Route("POST /restful/v2/domains/example.com/records",
			servermock.ResponseFromFixture("success.json"),
			servermock.CheckRequestJSONBodyFromFixture("set_dns-request.json"),
			servermock.CheckHeader().
				With("X-Signature", "5ajUPToqQdGz4QrXzXueT/QGWK94YbfX/tG8GRn8JRM="),
		).
		Build(t)

	payload := &SetDNSRequest{
		Subs: []SubRecord{{
			SubHost: "_acme-challenge",
			Record: Record{
				Type:   "txt",
				Value1: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
			},
		}},
		TTL:             120,
		AddDNSToCurrent: true,
	}

	err := client.SetDNS(t.Context(), "example.com", payload)
	require.NoError(t, err)
}

func TestClient_SetDNS_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /restful/v2/domains/example.com/records",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusForbidden),
		).
		Build(t)

	payload := &SetDNSRequest{
		Subs: []SubRecord{{
			SubHost: "_acme-challenge",
			Record: Record{
				Type:   "txt",
				Value1: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
			},
		}},
		TTL:             120,
		AddDNSToCurrent: true,
	}

	err := client.SetDNS(t.Context(), "example.com", payload)
	require.EqualError(t, err, "[code: 403] Forbidden: The domain doesn't have main dns.")
}

func TestClient_RemoveDNS(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /restful/v2/domains/example.com/records",
			servermock.ResponseFromFixture("success.json"),
			servermock.CheckRequestJSONBodyFromFixture("remove_dns-request.json"),
			servermock.CheckHeader().
				With("X-Signature", "ybrN00UL8nWclU1AmvyoIGb7uJURud5UPKFuOkGNtuI="),
		).
		Build(t)

	payload := &RemoveDNSRequest{
		Subs: []SubRecord{{
			SubHost: "_acme-challenge",
			Record: Record{
				Type:   "txt",
				Value1: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
			},
		}},
	}

	err := client.RemoveDNS(t.Context(), "example.com", payload)
	require.NoError(t, err)
}

func TestClient_RemoveDNS_error(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /restful/v2/domains/example.com/records",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusForbidden),
		).
		Build(t)

	payload := &RemoveDNSRequest{
		Subs: []SubRecord{{
			SubHost: "_acme-challenge",
			Record: Record{
				Type:   "txt",
				Value1: "ADw2sEd82DUgXcQ9hNBZThJs7zVJkR5v9JeSbAb9mZY",
			},
		}},
	}

	err := client.RemoveDNS(t.Context(), "example.com", payload)
	require.EqualError(t, err, "[code: 403] Forbidden: The domain doesn't have main dns.")
}
