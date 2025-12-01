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
			client := NewClient("user", "secret", 123)
			client.HTTPClient = server.Client()
			client.BaseURL, _ = url.Parse(server.URL)

			return client, nil
		},
		servermock.CheckHeader().
			WithBasicAuth("user", "secret").
			WithJSONHeaders())
}

func TestClient_AddRecords(t *testing.T) {
	client := mockBuilder().
		Route("POST /zone/example.com/_stream",
			servermock.ResponseFromFixture("add_record.json"),
			servermock.CheckRequestJSONBodyFromFixture("add_record-request.json"),
			servermock.CheckHeader().
				With("X-Domainrobot-Context", "123")).
		Build(t)

	records := []*ResourceRecord{{
		Name:  "example.com",
		TTL:   600,
		Type:  "TXT",
		Value: "txtTXTtxt",
	}}

	resp, err := client.AddRecords(t.Context(), "example.com", records)
	require.NoError(t, err)

	expected := &DataZoneResponse{
		STID: "20251121-appf4923-126284",
		CTID: "",
		Messages: []ResponseMessage{
			{
				Text: "string",
				Messages: []string{
					"string",
				},
				Objects: []GenericObject{
					{
						Type:  "string",
						Value: "string",
					},
				},
				Code:   "string",
				Status: "SUCCESS",
			},
		},
		Status: &ResponseStatus{
			Code: "S0301",
			Text: "Zone was updated successfully on the name server.",
			Type: "SUCCESS",
		},
		Object: nil,
		Data: []Zone{
			{
				Name: "example.com",
				ResourceRecords: []ResourceRecord{
					{
						Name:  "example.com",
						TTL:   120,
						Type:  "TXT",
						Value: "txt",
						Pref:  1,
					},
				},
				Action:            "xxx",
				VirtualNameServer: "yyy",
			},
		},
	}

	assert.Equal(t, expected, resp)
}

func TestClient_AddRecords_error(t *testing.T) {
	client := mockBuilder().
		Route("POST /zone/example.com/_stream",
			servermock.ResponseFromFixture("error.json").
				WithStatusCode(http.StatusBadRequest)).
		Build(t)

	records := []*ResourceRecord{{
		Name:  "example.com",
		TTL:   600,
		Type:  "TXT",
		Value: "txtTXTtxt",
	}}

	_, err := client.AddRecords(t.Context(), "example.com", records)
	require.EqualError(t, err, `STID: 20251121-appf4923-126284, status: code: E0202002, text: Zone konnte auf dem Nameserver nicht aktualisiert werden., type: ERROR, message: code: EF02022, text: Der Zusatzeintrag wurde doppelt eingetragen., status: ERROR, object: OURDOMAIN.TLD@nsa7.schlundtech.de/rr[17]: _acme-challenge.www.whoami.int.OURDOMAIN.TLD TXT "rK2SJb_ZcrYefbfCKU6jZEANfEAJeOtSh1Fv8hkUoVc"`)
}

func TestClient_RemoveRecords(t *testing.T) {
	client := mockBuilder().
		Route("POST /zone/example.com/_stream",
			servermock.ResponseFromFixture("remove_record.json"),
			servermock.CheckRequestJSONBodyFromFixture("remove_record-request.json"),
			servermock.CheckHeader().
				With("X-Domainrobot-Context", "123")).
		Build(t)

	records := []*ResourceRecord{{
		Name:  "example.com",
		TTL:   600,
		Type:  "TXT",
		Value: "txtTXTtxt",
	}}

	resp, err := client.RemoveRecords(t.Context(), "example.com", records)
	require.NoError(t, err)

	expected := &DataZoneResponse{
		STID: "20251121-appf4923-126284",
		CTID: "",
		Messages: []ResponseMessage{
			{
				Text: "string",
				Messages: []string{
					"string",
				},
				Objects: []GenericObject{
					{
						Type:  "string",
						Value: "string",
					},
				},
				Code:   "string",
				Status: "SUCCESS",
			},
		},
		Status: &ResponseStatus{
			Code: "S0301",
			Text: "Zone was updated successfully on the name server.",
			Type: "SUCCESS",
		},
		Object: nil,
		Data: []Zone{
			{
				Name: "example.com",
				ResourceRecords: []ResourceRecord{
					{
						Name:  "example.com",
						TTL:   120,
						Type:  "TXT",
						Value: "txt",
						Pref:  1,
					},
				},
				Action:            "xxx",
				VirtualNameServer: "yyy",
			},
		},
	}

	assert.Equal(t, expected, resp)
}
