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
			client, err := NewClient("secret", "shortname")
			if err != nil {
				return nil, err
			}

			client.HTTPClient = server.Client()
			client.baseURL, _ = url.Parse(server.URL)

			return client, nil
		},
		servermock.CheckHeader().WithJSONHeaders().
			WithAuthorization("APIToken secret"))
}

func TestClient_Create(t *testing.T) {
	client := mockBuilder().
		Route("POST /api/config/dns/namespaces/system/dns_zones/example.com/rrsets/groupA",
			servermock.ResponseFromFixture("create.json"),
			servermock.CheckRequestJSONBody(`{"dns_zone_name":"example.com","group_name":"groupA","rrset":{"description":"lego","ttl":60,"txt_record":{"name":"wwww","values":["txt"]}}}`)).
		Build(t)

	rrSet := RRSet{
		Description: "lego",
		TTL:         60,
		TXTRecord: &TXTRecord{
			Name:   "wwww",
			Values: []string{"txt"},
		},
	}

	result, err := client.CreateRRSet(t.Context(), "example.com", "groupA", rrSet)
	require.NoError(t, err)

	expected := &APIRRSet{
		DNSZoneName: "string",
		GroupName:   "string",
		RRSet: RRSet{
			Description: "string",
			TXTRecord: &TXTRecord{
				Name:   "string",
				Values: []string{"string"},
			},
		},
	}

	assert.Equal(t, expected, result)
}

func TestClient_Create_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /api/config/dns/namespaces/system/dns_zones/example.com/rrsets/groupA",
			servermock.Noop().WithStatusCode(http.StatusBadRequest)).
		Build(t)

	rrSet := RRSet{
		Description: "lego",
		TTL:         60,
		TXTRecord: &TXTRecord{
			Name:   "wwww",
			Values: []string{"txt"},
		},
	}

	_, err := client.CreateRRSet(t.Context(), "example.com", "groupA", rrSet)
	require.Error(t, err)
}

func TestClient_Get(t *testing.T) {
	client := mockBuilder().
		Route("GET /api/config/dns/namespaces/system/dns_zones/example.com/rrsets/groupA/www/TXT",
			servermock.ResponseFromFixture("get.json")).
		Build(t)

	result, err := client.GetRRSet(t.Context(), "example.com", "groupA", "www", "TXT")
	require.NoError(t, err)

	expected := &APIRRSet{
		DNSZoneName: "string",
		GroupName:   "string",
		Namespace:   "string",
		RecordName:  "string",
		Type:        "string",
		RRSet: RRSet{
			Description: "string",
			TXTRecord: &TXTRecord{
				Name:   "string",
				Values: []string{"string"},
			},
		},
	}

	assert.Equal(t, expected, result)
}

func TestClient_Get_not_found(t *testing.T) {
	client := mockBuilder().
		Route("GET /api/config/dns/namespaces/system/dns_zones/example.com/rrsets/groupA/www/TXT",
			servermock.ResponseFromFixture("error_404.json").WithStatusCode(http.StatusNotFound)).
		Build(t)

	result, err := client.GetRRSet(t.Context(), "example.com", "groupA", "www", "TXT")
	require.NoError(t, err)

	assert.Nil(t, result)
}

func TestClient_Get_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /api/config/dns/namespaces/system/dns_zones/example.com/rrsets/groupA/www/TXT",
			servermock.Noop().WithStatusCode(http.StatusBadRequest)).
		Build(t)

	_, err := client.GetRRSet(t.Context(), "example.com", "groupA", "www", "TXT")
	require.Error(t, err)
}

func TestClient_Delete(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /api/config/dns/namespaces/system/dns_zones/example.com/rrsets/groupA/www/TXT",
			servermock.ResponseFromFixture("get.json")).
		Build(t)

	result, err := client.DeleteRRSet(t.Context(), "example.com", "groupA", "www", "TXT")
	require.NoError(t, err)

	expected := &APIRRSet{
		DNSZoneName: "string",
		GroupName:   "string",
		Namespace:   "string",
		RecordName:  "string",
		Type:        "string",
		RRSet: RRSet{
			Description: "string",
			TXTRecord: &TXTRecord{
				Name:   "string",
				Values: []string{"string"},
			},
		},
	}

	assert.Equal(t, expected, result)
}

func TestClient_Delete_error(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /api/config/dns/namespaces/system/dns_zones/example.com/rrsets/groupA/www/TXT",
			servermock.Noop().WithStatusCode(http.StatusBadRequest)).
		Build(t)

	_, err := client.DeleteRRSet(t.Context(), "example.com", "groupA", "www", "TXT")
	require.Error(t, err)
}

func TestClient_Replace(t *testing.T) {
	client := mockBuilder().
		Route("PUT /api/config/dns/namespaces/system/dns_zones/example.com/rrsets/groupA/www/TXT",
			servermock.ResponseFromFixture("get.json"),
			servermock.CheckRequestJSONBody(`{"dns_zone_name":"example.com","group_name":"groupA","type":"TXT","rrset":{"description":"lego","ttl":60,"txt_record":{"name":"wwww","values":["txt"]}}}`)).
		Build(t)

	rrSet := RRSet{
		Description: "lego",
		TTL:         60,
		TXTRecord: &TXTRecord{
			Name:   "wwww",
			Values: []string{"txt"},
		},
	}

	result, err := client.ReplaceRRSet(t.Context(), "example.com", "groupA", "www", "TXT", rrSet)
	require.NoError(t, err)

	expected := &APIRRSet{
		DNSZoneName: "string",
		GroupName:   "string",
		Namespace:   "string",
		RecordName:  "string",
		Type:        "string",
		RRSet: RRSet{
			Description: "string",
			TXTRecord: &TXTRecord{
				Name:   "string",
				Values: []string{"string"},
			},
		},
	}

	assert.Equal(t, expected, result)
}

func TestClient_Replace_error(t *testing.T) {
	client := mockBuilder().
		Route("PUT /api/config/dns/namespaces/system/dns_zones/example.com/rrsets/groupA/www/TXT",
			servermock.Noop().WithStatusCode(http.StatusBadRequest)).
		Build(t)

	rrSet := RRSet{
		Description: "lego",
		TTL:         60,
		TXTRecord: &TXTRecord{
			Name:   "wwww",
			Values: []string{"txt"},
		},
	}

	_, err := client.ReplaceRRSet(t.Context(), "example.com", "groupA", "www", "TXT", rrSet)
	require.Error(t, err)
}
