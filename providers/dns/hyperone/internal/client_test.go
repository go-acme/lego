package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
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
	server := createTestServer(t, http.MethodGet, "/dns/loc123/project/proj123/zone/zone321/recordset", fromFile("recordset.json"))
	client := getTestClient(t, server.URL)

	recordset, err := client.FindRecordset("zone321", "SOA", "example.com.")
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

	server := createTestServer(t, http.MethodPost, "/dns/loc123/project/proj123/zone/zone123/recordset",
		hasReqBody(expectedReqBody), fromFile("createRecordset.json"))
	client := getTestClient(t, server.URL)

	rs, err := client.CreateRecordset("zone123", "TXT", "test.example.com.", "value", 3600)
	require.NoError(t, err)

	expected := &Recordset{RecordType: "TXT", Name: "test.example.com.", TTL: 3600, ID: "1234567890qwertyuiop"}
	assert.Equal(t, expected, rs)
}

func TestClient_DeleteRecordset(t *testing.T) {
	server := createTestServer(t, http.MethodDelete, "/dns/loc123/project/proj123/zone/zone321/recordset/rs322")
	client := getTestClient(t, server.URL)

	err := client.DeleteRecordset("zone321", "rs322")
	require.NoError(t, err)
}

func TestClient_GetRecords(t *testing.T) {
	server := createTestServer(t, http.MethodGet, "/dns/loc123/project/proj123/zone/321/recordset/322/record", fromFile("record.json"))
	client := getTestClient(t, server.URL)

	records, err := client.GetRecords("321", "322")
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

	server := createTestServer(t, http.MethodPost, "/dns/loc123/project/proj123/zone/z123/recordset/rs325/record",
		hasReqBody(expectedReqBody), fromFile("createRecord.json"))
	client := getTestClient(t, server.URL)

	rs, err := client.CreateRecord("z123", "rs325", "value")
	require.NoError(t, err)

	expected := &Record{ID: "123321qwerqwewqerq", Content: "value", Enabled: true}
	assert.Equal(t, expected, rs)
}

func TestClient_DeleteRecord(t *testing.T) {
	server := createTestServer(t, http.MethodDelete, "/dns/loc123/project/proj123/zone/321/recordset/322/record/323")
	client := getTestClient(t, server.URL)

	err := client.DeleteRecord("321", "322", "323")
	require.NoError(t, err)
}

func TestClient_FindZone(t *testing.T) {
	server := createTestServer(t, http.MethodGet, "/dns/loc123/project/proj123/zone", fromFile("zones.json"))
	client := getTestClient(t, server.URL)

	zone, err := client.FindZone("example.com")
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
	server := createTestServer(t, http.MethodGet, "/dns/loc123/project/proj123/zone", fromFile("zones.json"))
	client := getTestClient(t, server.URL)

	zones, err := client.GetZones()
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

type assertHandler func(http.ResponseWriter, *http.Request) (int, error)

func createTestServer(t *testing.T, method, path string, handlers ...assertHandler) *httptest.Server {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)

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

	t.Cleanup(server.Close)

	return server
}

func hasReqBody(v interface{}) assertHandler {
	return func(rw http.ResponseWriter, req *http.Request) (int, error) {
		reqBody, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return http.StatusBadRequest, err
		}

		marshal, err := json.Marshal(v)
		if err != nil {
			return http.StatusInternalServerError, err
		}

		if !bytes.Equal(marshal, reqBody) {
			return http.StatusBadRequest, fmt.Errorf("invalid request body, got: %s, expect: %s", string(reqBody), string(marshal))
		}

		return http.StatusOK, nil
	}
}

func fromFile(fixtureName string) assertHandler {
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

func getTestClient(t *testing.T, serverURL string) *Client {
	passport := &Passport{
		SubjectID: "/iam/project/proj123/sa/xxxxxxx",
	}

	client, err := NewClient(serverURL, "loc123", passport)
	require.NoError(t, err)

	client.signer = signerMock{}

	return client
}
