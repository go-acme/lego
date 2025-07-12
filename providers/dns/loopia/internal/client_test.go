package internal

import (
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockBuilder(password string) *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client := NewClient("apiuser", password)
			client.BaseURL = server.URL + "/"

			return client, nil
		},
		servermock.CheckHeader().WithContentType("text/xml"),
	)
}

func TestClient_AddZoneRecord(t *testing.T) {
	testCases := []struct {
		desc     string
		password string
		domain   string
		request  string
		response string
		err      string
	}{
		{
			desc:     "auth ok",
			password: "goodpassword",
			domain:   exampleDomain,
			request:  addZoneRecordGoodAuth,
			response: responseOk,
		},
		{
			desc:     "auth error",
			password: "badpassword",
			domain:   exampleDomain,
			request:  addZoneRecordBadAuth,
			response: responseAuthError,
			err:      "authentication error",
		},
		{
			desc:     "unknown error",
			password: "goodpassword",
			domain:   "badexample.com",
			request:  addZoneRecordNonValidDomain,
			response: responseUnknownError,
			err:      `unknown error: "UNKNOWN_ERROR"`,
		},
		{
			desc:     "empty response",
			password: "goodpassword",
			domain:   "empty.com",
			request:  addZoneRecordEmptyResponse,
			response: "",
			err:      "unmarshal error: EOF",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			client := mockBuilder(test.password).
				Route("POST /",
					servermock.RawStringResponse(test.response),
					servermock.CheckRequestBody(test.request)).
				Build(t)

			err := client.AddTXTRecord(t.Context(), test.domain, exampleSubDomain, 123, "TXTrecord")
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
	testCases := []struct {
		desc     string
		password string
		domain   string
		request  string
		response string
		err      string
	}{
		{
			desc:     "auth ok",
			password: "goodpassword",
			domain:   exampleDomain,
			request:  removeSubdomainGoodAuth,
			response: responseOk,
		},
		{
			desc:     "auth error",
			password: "badpassword",
			domain:   exampleDomain,
			request:  removeSubdomainBadAuth,
			response: responseAuthError,
			err:      "authentication error",
		},
		{
			desc:     "unknown error",
			password: "goodpassword",
			domain:   "badexample.com",
			request:  removeSubdomainNonValidDomain,
			response: responseUnknownError,
			err:      `unknown error: "UNKNOWN_ERROR"`,
		},
		{
			desc:     "empty response",
			password: "goodpassword",
			domain:   "empty.com",
			request:  removeSubdomainEmptyResponse,
			response: "",
			err:      "unmarshal error: EOF",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			client := mockBuilder(test.password).
				Route("POST /",
					servermock.RawStringResponse(test.response),
					servermock.CheckRequestBody(test.request)).
				Build(t)

			err := client.RemoveSubdomain(t.Context(), test.domain, exampleSubDomain)
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
	testCases := []struct {
		desc     string
		password string
		domain   string
		request  string
		response string
		err      string
	}{
		{
			desc:     "auth ok",
			password: "goodpassword",
			domain:   exampleDomain,
			request:  removeRecordGoodAuth,
			response: responseOk,
		},
		{
			desc:     "auth error",
			password: "badpassword",
			domain:   exampleDomain,
			request:  removeRecordBadAuth,
			response: responseAuthError,
			err:      "authentication error",
		},
		{
			desc:     "uknown error",
			password: "goodpassword",
			domain:   "badexample.com",
			request:  removeRecordNonValidDomain,
			response: responseUnknownError,
			err:      `unknown error: "UNKNOWN_ERROR"`,
		},
		{
			desc:     "empty response",
			password: "goodpassword",
			domain:   "empty.com",
			request:  removeRecordEmptyResponse,
			response: "",
			err:      "unmarshal error: EOF",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			client := mockBuilder(test.password).
				Route("POST /",
					servermock.RawStringResponse(test.response),
					servermock.CheckRequestBody(test.request)).
				Build(t)

			err := client.RemoveTXTRecord(t.Context(), test.domain, exampleSubDomain, 12345678)
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
	client := mockBuilder("goodpassword").
		Route("POST /",
			servermock.RawStringResponse(getZoneRecordsResponse),
			servermock.CheckRequestBody(getZoneRecords)).
		Build(t)

	recordObjs, err := client.GetTXTRecords(t.Context(), exampleDomain, exampleSubDomain)
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
	assert.Equal(t, expected, recordObjs)
}

func TestClient_rpcCall_404(t *testing.T) {
	client := mockBuilder("apipassword").
		Route("POST /",
			servermock.RawStringResponse("<?xml version='1.0' encoding='UTF-8'?>").
				WithStatusCode(http.StatusNotFound)).
		Build(t)

	call := &methodCall{
		MethodName: "dummyMethod",
		Params: []param{
			paramString{Value: "test1"},
		},
	}

	err := client.rpcCall(t.Context(), call, &responseString{})
	require.EqualError(t, err, "unexpected status code: [status code: 404] body: <?xml version='1.0' encoding='UTF-8'?>")
}

func TestClient_rpcCall_RPCError(t *testing.T) {
	client := mockBuilder("apipassword").
		Route("POST /",
			servermock.RawStringResponse(responseRPCError)).
		Build(t)

	call := &methodCall{
		MethodName: "getDomains",
		Params: []param{
			paramString{Value: "test1"},
		},
	}

	err := client.rpcCall(t.Context(), call, &responseString{})
	require.EqualError(t, err, "RPC Error: (201) Method signature error: 42")
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
