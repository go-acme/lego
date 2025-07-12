package internal

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTxtRecordService_Create(t *testing.T) {
	client := mockBuilder().
		Route("POST /v1/domains/12345/records/txt", servermock.ResponseFromFixture("records-Create.json"),
			servermock.CheckRequestJSONBody(`{"name":""}`)).
		Build(t)

	records, err := client.TxtRecords.Create(t.Context(), 12345, RecordRequest{})
	require.NoError(t, err)

	recordsJSON, err := json.Marshal(records)
	require.NoError(t, err)

	expectedContent, err := os.ReadFile("./fixtures/records-Create.json")
	require.NoError(t, err)

	assert.JSONEq(t, string(expectedContent), string(recordsJSON))
}

func TestTxtRecordService_GetAll(t *testing.T) {
	client := mockBuilder().
		Route("GET /v1/domains/12345/records/txt", servermock.ResponseFromFixture("records-GetAll.json")).
		Build(t)

	records, err := client.TxtRecords.GetAll(t.Context(), 12345)
	require.NoError(t, err)

	recordsJSON, err := json.Marshal(records)
	require.NoError(t, err)

	expectedContent, err := os.ReadFile("./fixtures/records-GetAll.json")
	require.NoError(t, err)

	assert.JSONEq(t, string(expectedContent), string(recordsJSON))
}

func TestTxtRecordService_Get(t *testing.T) {
	client := mockBuilder().
		Route("GET /v1/domains/12345/records/txt/6789", servermock.ResponseFromFixture("records-Get.json")).
		Build(t)

	record, err := client.TxtRecords.Get(t.Context(), 12345, 6789)
	require.NoError(t, err)

	expected := &Record{
		ID:           3557066,
		Type:         "TXT",
		RecordType:   "txt",
		Name:         "test",
		TTL:          300,
		RecordOption: "roundRobin",
		GtdRegion:    1,
		ParentID:     273302,
		Parent:       "domain",
		Source:       "Domain",
		ModifiedTS:   1580908547863,
		Value: []RecordValue{{
			Value: `"test"`,
		}},
		RoundRobin: []RecordValue{{
			Value: `"test"`,
		}},
	}
	assert.Equal(t, expected, record)
}

func TestTxtRecordService_Update(t *testing.T) {
	client := mockBuilder().
		Route("PUT /v1/domains/12345/records/txt/6789",
			servermock.RawStringResponse(`{"success":"Record  updated successfully"}`)).
		Build(t)

	msg, err := client.TxtRecords.Update(t.Context(), 12345, 6789, RecordRequest{})
	require.NoError(t, err)

	expected := &SuccessMessage{Success: "Record  updated successfully"}
	assert.Equal(t, expected, msg)
}

func TestTxtRecordService_Delete(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /v1/domains/12345/records/txt/6789",
			servermock.RawStringResponse(`{"success":"Record  deleted successfully"}`)).
		Build(t)

	msg, err := client.TxtRecords.Delete(t.Context(), 12345, 6789)
	require.NoError(t, err)

	expected := &SuccessMessage{Success: "Record  deleted successfully"}
	assert.Equal(t, expected, msg)
}

func TestTxtRecordService_Search(t *testing.T) {
	client := mockBuilder().
		Route("GET /v1/domains/12345/records/txt/search", servermock.ResponseFromFixture("records-Search.json")).
		Build(t)

	records, err := client.TxtRecords.Search(t.Context(), 12345, Exact, "test")
	require.NoError(t, err)

	recordsJSON, err := json.Marshal(records)
	require.NoError(t, err)

	expectedContent, err := os.ReadFile("./fixtures/records-Search.json")
	require.NoError(t, err)

	assert.JSONEq(t, string(expectedContent), string(recordsJSON))
}
