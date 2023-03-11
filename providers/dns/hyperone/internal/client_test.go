package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type signerMock struct{}

func (s signerMock) GetJWT() (string, error) {
	return "", nil
}

func TestClient_FindRecordset(t *testing.T) {
	client := setupTest(t, http.MethodGet, "/dns/loc123/project/proj123/zone/zone321/recordset", respFromFile("recordset.json"))

	recordset, err := client.FindRecordset(context.Background(), "zone321", "SOA", "example.com.")
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

	client := setupTest(t, http.MethodPost, "/dns/loc123/project/proj123/zone/zone123/recordset",
		hasReqBody(expectedReqBody), respFromFile("createRecordset.json"))

	rs, err := client.CreateRecordset(context.Background(), "zone123", "TXT", "test.example.com.", "value", 3600)
	require.NoError(t, err)

	expected := &Recordset{RecordType: "TXT", Name: "test.example.com.", TTL: 3600, ID: "1234567890qwertyuiop"}
	assert.Equal(t, expected, rs)
}

func TestClient_DeleteRecordset(t *testing.T) {
	client := setupTest(t, http.MethodDelete, "/dns/loc123/project/proj123/zone/zone321/recordset/rs322")

	err := client.DeleteRecordset(context.Background(), "zone321", "rs322")
	require.NoError(t, err)
}

func TestClient_GetRecords(t *testing.T) {
	client := setupTest(t, http.MethodGet, "/dns/loc123/project/proj123/zone/321/recordset/322/record", respFromFile("record.json"))

	records, err := client.GetRecords(context.Background(), "321", "322")
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

	client := setupTest(t, http.MethodPost, "/dns/loc123/project/proj123/zone/z123/recordset/rs325/record",
		hasReqBody(expectedReqBody), respFromFile("createRecord.json"))

	rs, err := client.CreateRecord(context.Background(), "z123", "rs325", "value")
	require.NoError(t, err)

	expected := &Record{ID: "123321qwerqwewqerq", Content: "value", Enabled: true}
	assert.Equal(t, expected, rs)
}

func TestClient_DeleteRecord(t *testing.T) {
	client := setupTest(t, http.MethodDelete, "/dns/loc123/project/proj123/zone/321/recordset/322/record/323")

	err := client.DeleteRecord(context.Background(), "321", "322", "323")
	require.NoError(t, err)
}

func TestClient_FindZone(t *testing.T) {
	client := setupTest(t, http.MethodGet, "/dns/loc123/project/proj123/zone", respFromFile("zones.json"))

	zone, err := client.FindZone(context.Background(), "example.com")
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
	client := setupTest(t, http.MethodGet, "/dns/loc123/project/proj123/zone", respFromFile("zones.json"))

	zones, err := client.GetZones(context.Background())
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

func setupTest(t *testing.T, method, path string, handlers ...assertHandler) *Client {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	mux.Handle(path, http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != method {
			http.Error(rw, fmt.Sprintf("unsupported method: %s", req.Method), http.StatusMethodNotAllowed)
			return
		}

		if len(handlers) != 0 {
			for _, handler := range handlers {
				code, err := handler(rw, req)
				if err != nil {
					http.Error(rw, err.Error(), code)
					return
				}
			}
		}
	}))

	passport := &Passport{
		SubjectID: "/iam/project/proj123/sa/xxxxxxx",
	}

	client, err := NewClient(server.URL, "loc123", passport)
	require.NoError(t, err)

	client.signer = signerMock{}

	return client
}

type assertHandler func(http.ResponseWriter, *http.Request) (int, error)

func hasReqBody(v interface{}) assertHandler {
	return func(rw http.ResponseWriter, req *http.Request) (int, error) {
		reqBody, err := io.ReadAll(req.Body)
		if err != nil {
			return http.StatusBadRequest, err
		}

		marshal, err := json.Marshal(v)
		if err != nil {
			return http.StatusInternalServerError, err
		}

		if !bytes.Equal(marshal, bytes.TrimSpace(reqBody)) {
			return http.StatusBadRequest, fmt.Errorf("invalid request body, got: %s, expect: %s", string(reqBody), string(marshal))
		}

		return http.StatusOK, nil
	}
}

func respFromFile(fixtureName string) assertHandler {
	return func(rw http.ResponseWriter, req *http.Request) (int, error) {
		file, err := os.Open(filepath.Join(".", "fixtures", fixtureName))
		if err != nil {
			return http.StatusInternalServerError, err
		}

		_, err = io.Copy(rw, file)
		if err != nil {
			return http.StatusInternalServerError, err
		}

		return http.StatusOK, nil
	}
}
