package internal

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTxtRecordService_Create(t *testing.T) {
	client, handler, tearDown := setupAPIMock()
	defer tearDown()

	handler.HandleFunc("/v1/domains/12345/records/txt", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(rw, "invalid method: "+req.Method, http.StatusBadRequest)
			return
		}

		file, err := os.Open("./fixtures/records-Create.json")
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		defer func() { _ = file.Close() }()

		_, err = io.Copy(rw, file)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	records, err := client.TxtRecords.Create(12345, RecordRequest{})
	require.NoError(t, err)

	recordsJSON, err := json.Marshal(records)
	require.NoError(t, err)

	expectedContent, err := os.ReadFile("./fixtures/records-Create.json")
	require.NoError(t, err)

	assert.JSONEq(t, string(expectedContent), string(recordsJSON))
}

func TestTxtRecordService_GetAll(t *testing.T) {
	client, handler, tearDown := setupAPIMock()
	defer tearDown()

	handler.HandleFunc("/v1/domains/12345/records/txt", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(rw, "invalid method: "+req.Method, http.StatusBadRequest)
			return
		}

		file, err := os.Open("./fixtures/records-GetAll.json")
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		defer func() { _ = file.Close() }()

		_, err = io.Copy(rw, file)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	records, err := client.TxtRecords.GetAll(12345)
	require.NoError(t, err)

	recordsJSON, err := json.Marshal(records)
	require.NoError(t, err)

	expectedContent, err := os.ReadFile("./fixtures/records-GetAll.json")
	require.NoError(t, err)

	assert.JSONEq(t, string(expectedContent), string(recordsJSON))
}

func TestTxtRecordService_Get(t *testing.T) {
	client, handler, tearDown := setupAPIMock()
	defer tearDown()

	handler.HandleFunc("/v1/domains/12345/records/txt/6789", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(rw, "invalid method: "+req.Method, http.StatusBadRequest)
			return
		}

		file, err := os.Open("./fixtures/records-Get.json")
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		defer func() { _ = file.Close() }()

		_, err = io.Copy(rw, file)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	record, err := client.TxtRecords.Get(12345, 6789)
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
	client, handler, tearDown := setupAPIMock()
	defer tearDown()

	handler.HandleFunc("/v1/domains/12345/records/txt/6789", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPut {
			http.Error(rw, "invalid method: "+req.Method, http.StatusBadRequest)
			return
		}

		_, err := rw.Write([]byte(`{"success":"Record  updated successfully"}`))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	msg, err := client.TxtRecords.Update(12345, 6789, RecordRequest{})
	require.NoError(t, err)

	expected := &SuccessMessage{Success: "Record  updated successfully"}
	assert.Equal(t, expected, msg)
}

func TestTxtRecordService_Delete(t *testing.T) {
	client, handler, tearDown := setupAPIMock()
	defer tearDown()

	handler.HandleFunc("/v1/domains/12345/records/txt/6789", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodDelete {
			http.Error(rw, "invalid method: "+req.Method, http.StatusBadRequest)
			return
		}

		_, err := rw.Write([]byte(`{"success":"Record  deleted successfully"}`))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	msg, err := client.TxtRecords.Delete(12345, 6789)
	require.NoError(t, err)

	expected := &SuccessMessage{Success: "Record  deleted successfully"}
	assert.Equal(t, expected, msg)
}

func TestTxtRecordService_Search(t *testing.T) {
	client, handler, tearDown := setupAPIMock()
	defer tearDown()

	handler.HandleFunc("/v1/domains/12345/records/txt/search", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(rw, "invalid method: "+req.Method, http.StatusBadRequest)
			return
		}

		file, err := os.Open("./fixtures/records-Search.json")
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		defer func() { _ = file.Close() }()

		_, err = io.Copy(rw, file)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	records, err := client.TxtRecords.Search(12345, Exact, "test")
	require.NoError(t, err)

	recordsJSON, err := json.Marshal(records)
	require.NoError(t, err)

	expectedContent, err := os.ReadFile("./fixtures/records-Search.json")
	require.NoError(t, err)

	assert.JSONEq(t, string(expectedContent), string(recordsJSON))
}
