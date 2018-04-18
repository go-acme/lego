package nifcloud

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func runTestServer(responseBody string, statusCode int) *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		fmt.Fprintln(w, responseBody)
	}))
	return server
}

func TestChangeResourceRecordSets(t *testing.T) {
	responseBody := `<?xml version="1.0" encoding="UTF-8"?>
<ChangeResourceRecordSetsResponse xmlns="https://route53.amazonaws.com/doc/2012-12-12/">
  <ChangeInfo>
    <Id>xxxxx</Id>
    <Status>INSYNC</Status>
    <SubmittedAt>2015-08-05T00:00:00.000Z</SubmittedAt>
  </ChangeInfo>
</ChangeResourceRecordSetsResponse>
`
	server := runTestServer(responseBody, http.StatusOK)
	defer server.Close()

	client := &Client{
		endpoint: server.URL,
	}

	res, err := client.ChangeResourceRecordSets("example.com", ChangeResourceRecordSetsRequest{})
	require.NoError(t, err)
	assert.Equal(t, "xxxxx", res.ChangeInfo.ID)
	assert.Equal(t, "INSYNC", res.ChangeInfo.Status)
	assert.Equal(t, "2015-08-05T00:00:00.000Z", res.ChangeInfo.SubmittedAt)
}

func TestGetChange(t *testing.T) {
	responseBody := `<?xml version="1.0" encoding="UTF-8"?>
<GetChangeResponse xmlns="https://route53.amazonaws.com/doc/2012-12-12/">
  <ChangeInfo>
    <Id>xxxxx</Id>
    <Status>INSYNC</Status>
    <SubmittedAt>2015-08-05T00:00:00.000Z</SubmittedAt>
  </ChangeInfo>
</GetChangeResponse>
`
	server := runTestServer(responseBody, http.StatusOK)
	defer server.Close()

	client := &Client{
		endpoint: server.URL,
	}

	res, err := client.GetChange("12345")
	require.NoError(t, err)
	assert.Equal(t, "xxxxx", res.ChangeInfo.ID)
	assert.Equal(t, "INSYNC", res.ChangeInfo.Status)
	assert.Equal(t, "2015-08-05T00:00:00.000Z", res.ChangeInfo.SubmittedAt)
}

func TestErrorCase(t *testing.T) {
	responseBody := `<?xml version="1.0" encoding="UTF-8"?>
<ErrorResponse>
  <Error>
    <Type>Sender</Type>
    <Code>AuthFailed</Code>
    <Message>The request signature we calculated does not match the signature you provided.</Message>
  </Error>
</ErrorResponse>
`
	server := runTestServer(responseBody, http.StatusUnauthorized)
	defer server.Close()

	client := &Client{
		endpoint: server.URL,
	}

	res, err := client.GetChange("12345")
	assert.Nil(t, res)
	assert.Equal(t, "The request signature we calculated does not match the signature you provided.", err.Error())
}
