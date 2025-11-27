package internal

import (
	"encoding/json"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupClient(server *httptest.Server) (*Client, error) {
	client := NewClient("secret")
	client.HTTPClient = server.Client()
	client.BaseURL, _ = url.Parse(server.URL)

	return client, nil
}

func TestClient_ListZoneConfigs(t *testing.T) {
	client := servermock.NewBuilder[*Client](setupClient).
		Route("POST /zoneConfigsFind",
			servermock.ResponseFromFixture("zoneConfigsFind.json"),
			servermock.CheckRequestJSONBodyFromFixture("zoneConfigsFind-request.json")).
		Build(t)

	zonesFind := ZoneConfigsFindRequest{
		Filter: Filter{Field: "zoneName", Value: "example.com"},
		Limit:  1,
		Page:   1,
	}

	zoneResponse, err := client.ListZoneConfigs(t.Context(), zonesFind)
	require.NoError(t, err)

	expected := &ZoneResponse{
		Limit:        10,
		Page:         1,
		TotalEntries: 15,
		TotalPages:   2,
		Type:         "FindZoneConfigsResult",
		Data: []ZoneConfig{{
			ID:                    "123",
			AccountID:             "456",
			Status:                "s",
			Name:                  "n",
			NameUnicode:           "u",
			MasterIP:              "m",
			Type:                  "t",
			EMailAddress:          "e",
			ZoneTransferWhitelist: []string{"a", "b"},
			LastChangeDate:        "l",
			DNSServerGroupID:      "g",
			DNSSecMode:            "m",
			SOAValues: &SOAValues{
				Refresh:     1,
				Retry:       2,
				Expire:      3,
				TTL:         4,
				NegativeTTL: 5,
			},
			TemplateValues: json.RawMessage(nil),
		}},
	}

	assert.Equal(t, expected, zoneResponse)
}

func TestClient_ListZoneConfigs_error(t *testing.T) {
	client := servermock.NewBuilder[*Client](setupClient).
		Route("POST /zoneConfigsFind",
			servermock.ResponseFromFixture("zoneConfigsFind_error.json")).
		Build(t)

	zonesFind := ZoneConfigsFindRequest{
		Filter: Filter{Field: "zoneName", Value: "example.com"},
		Limit:  1,
		Page:   1,
	}

	_, err := client.ListZoneConfigs(t.Context(), zonesFind)
	require.Error(t, err)
}

func TestClient_UpdateZone(t *testing.T) {
	client := servermock.NewBuilder[*Client](setupClient).
		Route("POST /zoneUpdate",
			servermock.ResponseFromFixture("zoneUpdate.json"),
			servermock.CheckRequestJSONBodyFromFixture("zoneUpdate-request.json")).
		Build(t)

	request := ZoneUpdateRequest{
		ZoneConfig: ZoneConfig{
			ID:                    "123",
			AccountID:             "456",
			Status:                "s",
			Name:                  "n",
			NameUnicode:           "u",
			MasterIP:              "m",
			Type:                  "t",
			EMailAddress:          "e",
			ZoneTransferWhitelist: []string{"a", "b"},
			LastChangeDate:        "l",
			DNSServerGroupID:      "g",
			DNSSecMode:            "m",
			SOAValues: &SOAValues{
				Refresh:     1,
				Retry:       2,
				Expire:      3,
				TTL:         4,
				NegativeTTL: 5,
			},
		},
		RecordsToDelete: []DNSRecord{{
			Type:    "TXT",
			Name:    "_acme-challenge.example.com",
			Content: `"txt"`,
		}},
	}

	response, err := client.UpdateZone(t.Context(), request)
	require.NoError(t, err)

	expected := &Zone{
		Records: []DNSRecord{{
			ID:               "123",
			ZoneID:           "456",
			RecordTemplateID: "789",
			Name:             "n",
			Type:             "TXT",
			Content:          "txt",
			TTL:              120,
			Priority:         5,
			LastChangeDate:   "d",
		}},
		ZoneConfig: ZoneConfig{
			ID:                    "123",
			AccountID:             "456",
			Status:                "s",
			Name:                  "n",
			NameUnicode:           "u",
			MasterIP:              "m",
			Type:                  "t",
			EMailAddress:          "e",
			ZoneTransferWhitelist: []string{"a", "b"},
			LastChangeDate:        "l",
			DNSServerGroupID:      "g",
			DNSSecMode:            "m",
			SOAValues: &SOAValues{
				Refresh:     1,
				Retry:       2,
				Expire:      3,
				TTL:         4,
				NegativeTTL: 5,
			},
		},
	}

	assert.Equal(t, expected, response)
}

func TestClient_UpdateZone_error(t *testing.T) {
	client := servermock.NewBuilder[*Client](setupClient).
		Route("POST /zoneUpdate",
			servermock.ResponseFromFixture("zoneUpdate_error.json")).
		Build(t)

	request := ZoneUpdateRequest{
		ZoneConfig: ZoneConfig{
			ID:                    "123",
			AccountID:             "456",
			Status:                "s",
			Name:                  "n",
			NameUnicode:           "u",
			MasterIP:              "m",
			Type:                  "t",
			EMailAddress:          "e",
			ZoneTransferWhitelist: []string{"a", "b"},
			LastChangeDate:        "l",
			DNSServerGroupID:      "g",
			DNSSecMode:            "m",
			SOAValues: &SOAValues{
				Refresh:     1,
				Retry:       2,
				Expire:      3,
				TTL:         4,
				NegativeTTL: 5,
			},
		},
		RecordsToDelete: []DNSRecord{{
			Type:    "TXT",
			Name:    "_acme-challenge.example.com",
			Content: `"txt"`,
		}},
	}

	_, err := client.UpdateZone(t.Context(), request)
	require.Error(t, err)
}
