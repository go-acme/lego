package internal

import (
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type signerMock struct{}

func (s signerMock) GetJWT() (string, error) {
	return "", nil
}

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			passport := &Passport{
				SubjectID: "/iam/project/proj123/sa/xxxxxxx",
			}

			client, err := NewClient(server.URL, "loc123", passport)
			if err != nil {
				return nil, err
			}

			client.HTTPClient = server.Client()
			client.signer = signerMock{}

			return client, nil
		},
		servermock.CheckHeader().WithJSONHeaders().
			WithAuthorization("Bearer"))
}

func TestClient_FindRecordset(t *testing.T) {
	client := mockBuilder().
		Route("GET /dns/loc123/project/proj123/zone/zone321/recordset",
			servermock.ResponseFromFixture("recordset.json")).
		Build(t)

	recordset, err := client.FindRecordset(t.Context(), "zone321", "SOA", "example.com.")
	require.NoError(t, err)

	expected := &Recordset{
		ID:         "123456789abcd",
		Name:       "example.com.",
		RecordType: "SOA",
		TTL:        1800,
	}

	assert.Equal(t, expected, recordset)
}

func TestClient_CreateRecordset(t *testing.T) {
	expectedReqBody := Recordset{
		RecordType: "TXT",
		Name:       "test.example.com.",
		TTL:        3600,
		Record:     &Record{Content: "value"},
	}

	client := mockBuilder().
		Route("POST /dns/loc123/project/proj123/zone/zone123/recordset",
			servermock.ResponseFromFixture("createRecordset.json"),
			servermock.CheckRequestJSONBodyFromStruct(expectedReqBody)).
		Build(t)

	rs, err := client.CreateRecordset(t.Context(), "zone123", "TXT", "test.example.com.", "value", 3600)
	require.NoError(t, err)

	expected := &Recordset{RecordType: "TXT", Name: "test.example.com.", TTL: 3600, ID: "1234567890qwertyuiop"}
	assert.Equal(t, expected, rs)
}

func TestClient_DeleteRecordset(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /dns/loc123/project/proj123/zone/zone321/recordset/rs322", nil).
		Build(t)

	err := client.DeleteRecordset(t.Context(), "zone321", "rs322")
	require.NoError(t, err)
}

func TestClient_GetRecords(t *testing.T) {
	client := mockBuilder().
		Route("GET /dns/loc123/project/proj123/zone/321/recordset/322/record",
			servermock.ResponseFromFixture("record.json")).
		Build(t)

	records, err := client.GetRecords(t.Context(), "321", "322")
	require.NoError(t, err)

	expected := []Record{
		{
			ID:      "135128352183572dd",
			Content: "pns.hyperone.com. hostmaster.hyperone.com. 1 15 180 1209600 1800",
			Enabled: true,
		},
	}

	assert.Equal(t, expected, records)
}

func TestClient_CreateRecord(t *testing.T) {
	expectedReqBody := Record{
		Content: "value",
	}

	client := mockBuilder().
		Route("POST /dns/loc123/project/proj123/zone/z123/recordset/rs325/record",
			servermock.ResponseFromFixture("createRecord.json"),
			servermock.CheckRequestJSONBodyFromStruct(expectedReqBody)).
		Build(t)

	rs, err := client.CreateRecord(t.Context(), "z123", "rs325", "value")
	require.NoError(t, err)

	expected := &Record{ID: "123321qwerqwewqerq", Content: "value", Enabled: true}
	assert.Equal(t, expected, rs)
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /dns/loc123/project/proj123/zone/321/recordset/322/record/323",
			servermock.ResponseFromFixture("createRecord.json")).
		Build(t)

	err := client.DeleteRecord(t.Context(), "321", "322", "323")
	require.NoError(t, err)
}

func TestClient_FindZone(t *testing.T) {
	client := mockBuilder().
		Route("GET /dns/loc123/project/proj123/zone",
			servermock.ResponseFromFixture("zones.json")).
		Build(t)

	zone, err := client.FindZone(t.Context(), "example.com")
	require.NoError(t, err)

	expected := &Zone{
		ID:      "zoneB",
		Name:    "example.com",
		DNSName: "example.com",
		FQDN:    "example.com.",
		URI:     "",
	}

	assert.Equal(t, expected, zone)
}

func TestClient_GetZones(t *testing.T) {
	client := mockBuilder().
		Route("GET /dns/loc123/project/proj123/zone",
			servermock.ResponseFromFixture("zones.json")).
		Build(t)

	zones, err := client.GetZones(t.Context())
	require.NoError(t, err)

	expected := []Zone{
		{
			ID:      "zoneA",
			Name:    "example.org",
			DNSName: "example.org",
			FQDN:    "example.org.",
			URI:     "",
		},
		{
			ID:      "zoneB",
			Name:    "example.com",
			DNSName: "example.com",
			FQDN:    "example.com.",
			URI:     "",
		},
	}

	assert.Equal(t, expected, zones)
}
