package loopia

import (
	"encoding/xml"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddZoneRecord(t *testing.T) {
	serverResponses := map[string]string{
		addZoneRecordGoodAuth:       responseOk,
		addZoneRecordBadAuth:        responseAuthError,
		addZoneRecordNonValidDomain: responseUnknownError,
		addZoneRecordEmptyResponse:  "",
	}

	fakeServer := createFakeServer(t, serverResponses)
	defer fakeServer.Close()

	tt := []struct {
		name     string
		password string
		domain   string
		err      string
	}{
		{name: "auth ok", password: "goodpassword", domain: exampleDomain},
		{name: "auth error", password: "badpassword", domain: exampleDomain, err: "Authentication Error"},
		{name: "unknown error", password: "goodpassword", domain: "badexample.com", err: "Unknown Error: 'UNKNOWN_ERROR'"},
		{name: "empty response", password: "goodpassword", domain: "empty.com", err: "unmarshal error: EOF"},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			client := NewClient("apiuser", tc.password)
			client.BaseURL = fakeServer.URL + "/"
			client.HTTPClient = fakeServer.Client()

			err := client.addTXTRecord(tc.domain, acmeChallenge, 123, "TXTrecord")
			if len(tc.err) == 0 {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.EqualError(t, err, tc.err)
			}
		})
	}
}

func TestRpcCall404(t *testing.T) {
	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := ioutil.ReadAll(r.Body)
		require.NoError(t, err)
		w.WriteHeader(http.StatusNotFound)
		_, err = io.Copy(w, strings.NewReader(string("<?xml version='1.0' encoding='UTF-8'?>")))
		require.NoError(t, err)
	}))
	defer fakeServer.Close()

	call := &methodCall{
		MethodName: "dummyMethod",
		Params: []param{
			paramString{Value: "test1"},
		},
	}

	client := NewClient("apiuser", "apipassword")
	client.BaseURL = fakeServer.URL + "/"
	client.HTTPClient = fakeServer.Client()

	err := client.rpcCall(call, &responseString{})
	assert.EqualError(t, err, "HTTP Post Error: 404")
}

func TestRpcCallRPCError(t *testing.T) {
	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := ioutil.ReadAll(r.Body)
		require.NoError(t, err)
		_, err = io.Copy(w, strings.NewReader(string(responseRPCError)))
		require.NoError(t, err)
	}))
	defer fakeServer.Close()

	call := &methodCall{
		MethodName: "getDomains",
		Params: []param{
			paramString{Value: "test1"},
		},
	}

	client := NewClient("apiuser", "apipassword")
	client.BaseURL = fakeServer.URL + "/"
	client.HTTPClient = fakeServer.Client()

	err := client.rpcCall(call, &responseString{})
	assert.EqualError(t, err, "Loopia DNS: RPC Error: (201) Method signature error: 42")
}

func TestRemoveSubdomain(t *testing.T) {
	serverResponses := map[string]string{
		removeSubdomainGoodAuth:       responseOk,
		removeSubdomainBadAuth:        responseAuthError,
		removeSubdomainNonValidDomain: responseUnknownError,
		removeSubdomainEmptyResponse:  "",
	}
	fakeServer := createFakeServer(t, serverResponses)
	defer fakeServer.Close()

	tt := []struct {
		name     string
		password string
		domain   string
		err      string
	}{
		{name: "auth ok", password: "goodpassword", domain: exampleDomain},
		{name: "auth error", password: "badpassword", domain: exampleDomain, err: "Authentication Error"},
		{name: "uknown error", password: "goodpassword", domain: "badexample.com", err: "Unknown Error: 'UNKNOWN_ERROR'"},
		{name: "empty response", password: "goodpassword", domain: "empty.com", err: "unmarshal error: EOF"},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			client := NewClient("apiuser", tc.password)
			client.BaseURL = fakeServer.URL + "/"
			client.HTTPClient = fakeServer.Client()
			err := client.removeSubdomain(tc.domain, acmeChallenge)
			if len(tc.err) == 0 {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.EqualError(t, err, tc.err)
			}
		})
	}
}

func TestRemoveZoneRecord(t *testing.T) {
	serverResponses := map[string]string{
		removeRecordGoodAuth:       responseOk,
		removeRecordBadAuth:        responseAuthError,
		removeRecordNonValidDomain: responseUnknownError,
		removeRecordEmptyResponse:  "",
	}
	fakeServer := createFakeServer(t, serverResponses)
	defer fakeServer.Close()

	tt := []struct {
		name     string
		password string
		domain   string
		err      string
	}{
		{name: "auth ok", password: "goodpassword", domain: exampleDomain},
		{name: "auth error", password: "badpassword", domain: exampleDomain, err: "Authentication Error"},
		{name: "uknown error", password: "goodpassword", domain: "badexample.com", err: "Unknown Error: 'UNKNOWN_ERROR'"},
		{name: "empty response", password: "goodpassword", domain: "empty.com", err: "unmarshal error: EOF"},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			client := NewClient("apiuser", tc.password)
			client.BaseURL = fakeServer.URL + "/"
			client.HTTPClient = fakeServer.Client()
			err := client.removeTXTRecord(tc.domain, acmeChallenge, 12345678)
			if len(tc.err) == 0 {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.EqualError(t, err, tc.err)
			}
		})
	}
}

func TestGetZoneRecord(t *testing.T) {
	serverResponses := map[string]string{
		getZoneRecords: getZoneRecordsResponse,
	}
	fakeServer := createFakeServer(t, serverResponses)
	defer fakeServer.Close()

	client := NewClient("apiuser", "goodpassword")
	client.BaseURL = fakeServer.URL + "/"
	client.HTTPClient = fakeServer.Client()
	recordObjs, err := client.getTXTRecords(exampleDomain, acmeChallenge)
	assert.NoError(t, err)

	expected := []recordObj{
		{
			Type:     "TXT",
			TTL:      300,
			Priority: 0,
			Rdata:    exampleRdata,
			RecordID: 12345678,
		},
	}
	assert.EqualValues(t, expected, recordObjs)
}

func TestUnmarshallFaultyRecordObject(t *testing.T) {
	tt := []struct {
		name string
		xml  string
	}{
		{name: "faulty name", xml: "<name>name<name>"},
		{name: "faulty string", xml: "<value><string>foo<string></value>"},
		{name: "faulty int", xml: "<value><int>1<int></value>"},
	}

	resp := &recordObj{}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			err := xml.Unmarshal([]byte(tc.xml), resp)
			require.Error(t, err)
		})
	}
}

func createFakeServer(t *testing.T, serverResponses map[string]string) *httptest.Server {
	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "text/xml", r.Header.Get("Content-Type"), "invalid content type")
		req, err := ioutil.ReadAll(r.Body)
		require.NoError(t, err)
		resp, ok := serverResponses[string(req)]
		require.True(t, ok)
		_, err = io.Copy(w, strings.NewReader(resp))
		require.NoError(t, err)
	}))
	return fakeServer
}
