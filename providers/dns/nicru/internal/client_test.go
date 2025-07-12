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
			client, err := NewClient(server.Client())
			if err != nil {
				return nil, err
			}

			client.baseURL, _ = url.Parse(server.URL)

			return client, nil
		},
		servermock.CheckHeader().
			WithAccept("text/xml"),
	)
}

func TestClient_GetServices(t *testing.T) {
	client := mockBuilder().
		Route("GET /services", servermock.ResponseFromFixture("services_GET.xml")).
		Build(t)

	zones, err := client.GetServices(t.Context())
	require.NoError(t, err)

	expected := []Service{
		{
			Admin:        "123/NIC-REG",
			DomainsLimit: "12",
			DomainsNum:   "5",
			Enable:       "true",
			HasPrimary:   "false",
			Name:         "testservice",
			Payer:        "123/NIC-REG",
			Tariff:       "Secondary L",
		},
		{
			Admin:        "123/NIC-REG",
			DomainsLimit: "150",
			DomainsNum:   "10",
			Enable:       "true",
			HasPrimary:   "true",
			Name:         "myservice",
			Payer:        "123/NIC-REG",
			Tariff:       "DNS-master XXL",
			RRLimit:      "7500",
			RRNum:        "1000",
		},
	}

	assert.Equal(t, expected, zones)
}

func TestClient_ListZones(t *testing.T) {
	client := mockBuilder().
		Route("GET /zones", servermock.ResponseFromFixture("zones_all_GET.xml")).
		Build(t)

	zones, err := client.ListZones(t.Context())
	require.NoError(t, err)

	expected := []Zone{
		{
			Admin:      "123/NIC-REG",
			Enable:     "true",
			HasChanges: "false",
			HasPrimary: "true",
			ID:         "227645",
			IDNName:    "тест.рф",
			Name:       "xn—e1aybc.xn--p1ai",
			Payer:      "123/NIC-REG",
			Service:    "myservice",
		},
		{
			Admin:      "123/NIC-REG",
			Enable:     "true",
			HasChanges: "false",
			HasPrimary: "true",
			ID:         "227642",
			IDNName:    "example.ru",
			Name:       "example.ru",
			Payer:      "123/NIC-REG",
			Service:    "myservice",
		},
		{
			Admin:      "123/NIC-REG",
			Enable:     "true",
			HasChanges: "false",
			HasPrimary: "true",
			ID:         "227643",
			IDNName:    "test.su",
			Name:       "test.su",
			Payer:      "123/NIC-REG",
			Service:    "myservice",
		},
	}

	assert.Equal(t, expected, zones)
}

func TestClient_ListZones_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /zones", servermock.ResponseFromFixture("errors.xml")).
		Build(t)

	_, err := client.ListZones(t.Context())
	require.ErrorIs(t, err, Error{
		Text: "Access token expired or not found",
		Code: "4097",
	})
}

func TestClient_GetZonesByService(t *testing.T) {
	client := mockBuilder().
		Route("GET /services/test/zones",
			servermock.ResponseFromFixture("zones_GET.xml")).
		Build(t)

	zones, err := client.GetZonesByService(t.Context(), "test")
	require.NoError(t, err)

	expected := []Zone{
		{
			Admin:      "123/NIC-REG",
			Enable:     "true",
			HasChanges: "false",
			HasPrimary: "true",
			ID:         "227645",
			IDNName:    "тест.рф",
			Name:       "xn—e1aybc.xn--p1ai",
			Payer:      "123/NIC-REG",
			Service:    "myservice",
		},
		{
			Admin:      "123/NIC-REG",
			Enable:     "true",
			HasChanges: "false",
			HasPrimary: "true",
			ID:         "227642",
			IDNName:    "example.ru",
			Name:       "example.ru",
			Payer:      "123/NIC-REG",
			Service:    "myservice",
		},
		{
			Admin:      "123/NIC-REG",
			Enable:     "true",
			HasChanges: "false",
			HasPrimary: "true",
			ID:         "227643",
			IDNName:    "test.su",
			Name:       "test.su",
			Payer:      "123/NIC-REG",
			Service:    "myservice",
		},
	}

	assert.Equal(t, expected, zones)
}

func TestClient_GetZonesByService_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /services/test/zones",
			servermock.ResponseFromFixture("errors.xml")).
		Build(t)

	_, err := client.GetZonesByService(t.Context(), "test")
	require.ErrorIs(t, err, Error{
		Text: "Access token expired or not found",
		Code: "4097",
	})
}

func TestClient_GetRecords(t *testing.T) {
	client := mockBuilder().
		Route("GET /services/test/zones/example.com./records",
			servermock.ResponseFromFixture("records_GET.xml")).
		Build(t)

	records, err := client.GetRecords(t.Context(), "test", "example.com.")
	require.NoError(t, err)

	expected := []RR{
		{
			ID:      "210074",
			Name:    "@",
			IDNName: "@",
			TTL:     "",
			Type:    "SOA",
			SOA: &SOA{
				MName: &MName{
					Name:    "ns3-l2.nic.ru.",
					IDNName: "ns3-l2.nic.ru.",
				},
				RName: &RName{
					Name:    "dns.nic.ru.",
					IDNName: "dns.nic.ru.",
				},
				Serial:  "2011112002",
				Refresh: "1440",
				Retry:   "3600",
				Expire:  "2592000",
				Minimum: "600",
			},
		},
		{
			ID:      "210075",
			Name:    "@",
			IDNName: "@",
			Type:    "NS",
			NS: &NS{
				Name:    "ns3-l2.nic.ru.",
				IDNName: "ns3- l2.nic.ru.",
			},
		},
		{
			ID:      "210076",
			Name:    "@",
			IDNName: "@",
			Type:    "NS",
			NS: &NS{
				Name:    "ns4-l2.nic.ru.",
				IDNName: "ns4-l2.nic.ru.",
			},
		},
		{
			ID:      "210077",
			Name:    "@",
			IDNName: "@",
			Type:    "NS",
			NS: &NS{
				Name:    "ns8-l2.nic.ru.",
				IDNName: "ns8- l2.nic.ru.",
			},
		},
	}

	assert.Equal(t, expected, records)
}

func TestClient_GetRecords_error(t *testing.T) {
	client := mockBuilder().
		Route("GET /services/test/zones/example.com./records",
			servermock.ResponseFromFixture("errors.xml")).
		Build(t)

	_, err := client.GetRecords(t.Context(), "test", "example.com.")
	require.ErrorIs(t, err, Error{
		Text: "Access token expired or not found",
		Code: "4097",
	})
}

func TestClient_AddRecord(t *testing.T) {
	client := mockBuilder().
		Route("PUT /services/test/zones/example.com./records",
			servermock.ResponseFromFixture("records_PUT.xml"),
			servermock.CheckHeader().
				WithContentType("text/xml")).
		Build(t)

	rrs := []RR{
		{
			Name: "@",
			Type: "NS",
			NS:   &NS{Name: "ns4-l2.nic.ru."},
		},
		{
			Name: "@",
			Type: "NS",
			NS:   &NS{Name: "ns8-l2.nic.ru."},
		},
	}

	response, err := client.AddRecords(t.Context(), "test", "example.com.", rrs)
	require.NoError(t, err)

	expected := []Zone{
		{
			Admin:      "123/NIC-REG",
			HasChanges: "true",
			ID:         "228095",
			IDNName:    "test.ru",
			Name:       "test.ru",
			Service:    "testservice",
			RR: []RR{
				{
					ID:      "210076",
					Name:    "@",
					IDNName: "@",
					Type:    "NS",
					NS: &NS{
						Name:    "ns4-l2.nic.ru.",
						IDNName: "ns4-l2.nic.ru.",
					},
				},
				{
					ID:      "210077",
					Name:    "@",
					IDNName: "@",
					Type:    "NS",
					NS: &NS{
						Name:    "ns8-l2.nic.ru.",
						IDNName: "ns8-l2.nic.ru.",
					},
				},
			},
		},
	}

	assert.Equal(t, expected, response)
}

func TestClient_AddRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("PUT /services/test/zones/example.com./records",
			servermock.ResponseFromFixture("errors.xml"),
			servermock.CheckHeader().
				WithContentType("text/xml")).
		Build(t)

	rrs := []RR{
		{
			Name: "@",
			Type: "NS",
			NS:   &NS{Name: "ns4-l2.nic.ru."},
		},
		{
			Name: "@",
			Type: "NS",
			NS:   &NS{Name: "ns8-l2.nic.ru."},
		},
	}

	_, err := client.AddRecords(t.Context(), "test", "example.com.", rrs)
	require.ErrorIs(t, err, Error{
		Text: "Access token expired or not found",
		Code: "4097",
	})
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /services/test/zones/example.com./records/123",
			servermock.ResponseFromFixture("record_DELETE.xml")).
		Build(t)

	err := client.DeleteRecord(t.Context(), "test", "example.com.", "123")
	require.NoError(t, err)
}

func TestClient_DeleteRecord_error(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /services/test/zones/example.com./records/123",
			servermock.ResponseFromFixture("errors.xml")).
		Build(t)

	err := client.DeleteRecord(t.Context(), "test", "example.com.", "123")
	require.ErrorIs(t, err, Error{
		Text: "Access token expired or not found",
		Code: "4097",
	})
}

func TestClient_CommitZone(t *testing.T) {
	client := mockBuilder().
		Route("POST /services/test/zones/example.com./commit",
			servermock.ResponseFromFixture("commit_POST.xml")).
		Build(t)

	err := client.CommitZone(t.Context(), "test", "example.com.")
	require.NoError(t, err)
}

func TestClient_CommitZone_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /services/test/zones/example.com./commit",
			servermock.ResponseFromFixture("errors.xml")).
		Build(t)

	err := client.CommitZone(t.Context(), "test", "example.com.")
	require.ErrorIs(t, err, Error{
		Text: "Access token expired or not found",
		Code: "4097",
	})
}
