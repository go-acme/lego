package internal

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_AddZoneRecord(t *testing.T) {
	serverResponses := map[string]string{
		addZoneRecordGoodAuth:       responseOk,
		addZoneRecordBadAuth:        responseAuthError,
		addZoneRecordNonValidDomain: responseUnknownError,
		addZoneRecordEmptyResponse:  "",
	}

	serverURL := createFakeServer(t, serverResponses)

	testCases := []struct {
		desc     string
		password string
		domain   string
		err      string
	}{
		{
			desc:     "auth ok",
			password: "goodpassword",
			domain:   exampleDomain,
		},
		{
			desc:     "auth error",
			password: "badpassword",
			domain:   exampleDomain,
			err:      "authentication error",
		},
		{
			desc:     "unknown error",
			password: "goodpassword",
			domain:   "badexample.com",
			err:      `unknown error: "UNKNOWN_ERROR"`,
		},
		{
			desc:     "empty response",
			password: "goodpassword",
			domain:   "empty.com",
			err:      "unmarshal error: EOF",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			client := NewClient("apiuser", test.password)
			client.BaseURL = serverURL + "/"

			err := client.AddTXTRecord(context.Background(), test.domain, exampleSubDomain, 123, "TXTrecord")
			if test.err == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.EqualError(t, err, test.err)
			}
		})
	}
}

func TestClient_RemoveSubdomain(t *testing.T) {
	serverResponses := map[string]string{
		removeSubdomainGoodAuth:       responseOk,
		removeSubdomainBadAuth:        responseAuthError,
		removeSubdomainNonValidDomain: responseUnknownError,
		removeSubdomainEmptyResponse:  "",
	}

	serverURL := createFakeServer(t, serverResponses)

	testCases := []struct {
		desc     string
		password string
		domain   string
		err      string
	}{
		{
			desc:     "auth ok",
			password: "goodpassword",
			domain:   exampleDomain,
		},
		{
			desc:     "auth error",
			password: "badpassword",
			domain:   exampleDomain,
			err:      "authentication error",
		},
		{
			desc:     "unknown error",
			password: "goodpassword",
			domain:   "badexample.com",
			err:      `unknown error: "UNKNOWN_ERROR"`,
		},
		{
			desc:     "empty response",
			password: "goodpassword",
			domain:   "empty.com",
			err:      "unmarshal error: EOF",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			client := NewClient("apiuser", test.password)
			client.BaseURL = serverURL + "/"

			err := client.RemoveSubdomain(context.Background(), test.domain, exampleSubDomain)
			if test.err == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.EqualError(t, err, test.err)
			}
		})
	}
}

func TestClient_RemoveZoneRecord(t *testing.T) {
	serverResponses := map[string]string{
		removeRecordGoodAuth:       responseOk,
		removeRecordBadAuth:        responseAuthError,
		removeRecordNonValidDomain: responseUnknownError,
		removeRecordEmptyResponse:  "",
	}

	serverURL := createFakeServer(t, serverResponses)

	testCases := []struct {
		desc     string
		password string
		domain   string
		err      string
	}{
		{
			desc:     "auth ok",
			password: "goodpassword",
			domain:   exampleDomain,
		},
		{
			desc:     "auth error",
			password: "badpassword",
			domain:   exampleDomain,
			err:      "authentication error",
		},
		{
			desc:     "uknown error",
			password: "goodpassword",
			domain:   "badexample.com",
			err:      `unknown error: "UNKNOWN_ERROR"`,
		},
		{
			desc:     "empty response",
			password: "goodpassword",
			domain:   "empty.com",
			err:      "unmarshal error: EOF",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			client := NewClient("apiuser", test.password)
			client.BaseURL = serverURL + "/"

			err := client.RemoveTXTRecord(context.Background(), test.domain, exampleSubDomain, 12345678)
			if test.err == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.EqualError(t, err, test.err)
			}
		})
	}
}

func TestClient_GetZoneRecord(t *testing.T) {
	serverResponses := map[string]string{
		getZoneRecords: getZoneRecordsResponse,
	}

	serverURL := createFakeServer(t, serverResponses)

	client := NewClient("apiuser", "goodpassword")
	client.BaseURL = serverURL + "/"

	recordObjs, err := client.GetTXTRecords(context.Background(), exampleDomain, exampleSubDomain)
	require.NoError(t, err)

	expected := []RecordObj{
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

func TestClient_rpcCall_404(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusNotFound)

		_, err = fmt.Fprint(w, "<?xml version='1.0' encoding='UTF-8'?>")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}))

	t.Cleanup(server.Close)

	call := &methodCall{
		MethodName: "dummyMethod",
		Params: []param{
			paramString{Value: "test1"},
		},
	}

	client := NewClient("apiuser", "apipassword")
	client.BaseURL = server.URL + "/"

	err := client.rpcCall(context.Background(), call, &responseString{})
	assert.EqualError(t, err, "unexpected status code: [status code: 404] body: <?xml version='1.0' encoding='UTF-8'?>")
}

func TestClient_rpcCall_RPCError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		_, err = fmt.Fprint(w, responseRPCError)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}))

	t.Cleanup(server.Close)

	call := &methodCall{
		MethodName: "getDomains",
		Params: []param{
			paramString{Value: "test1"},
		},
	}

	client := NewClient("apiuser", "apipassword")
	client.BaseURL = server.URL + "/"

	err := client.rpcCall(context.Background(), call, &responseString{})
	assert.EqualError(t, err, "RPC Error: (201) Method signature error: 42")
}

func TestUnmarshallFaultyRecordObject(t *testing.T) {
	testCases := []struct {
		desc string
		xml  string
	}{
		{
			desc: "faulty name",
			xml:  "<name>name<name>",
		},
		{
			desc: "faulty string",
			xml:  "<value><string>foo<string></value>",
		},
		{
			desc: "faulty int",
			xml:  "<value><int>1<int></value>",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			resp := &RecordObj{}

			err := xml.Unmarshal([]byte(test.xml), resp)
			require.Error(t, err)
		})
	}
}

func createFakeServer(t *testing.T, serverResponses map[string]string) string {
	t.Helper()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "text/xml" {
			http.Error(w, fmt.Sprintf("invalid content type: %s", r.Header.Get("Content-Type")), http.StatusBadRequest)
			return
		}

		req, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		resp, ok := serverResponses[string(req)]
		if !ok {
			http.Error(w, "no response for request", http.StatusBadRequest)
			return
		}

		_, err = fmt.Fprint(w, resp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	return server.URL
}
