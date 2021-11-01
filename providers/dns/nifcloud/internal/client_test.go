package internal

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T, responseBody string, statusCode int) *Client {
	t.Helper()

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(statusCode)
		_, _ = fmt.Fprintln(w, responseBody)
	})

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	client, err := NewClient("A", "B")
	require.NoError(t, err)

	client.HTTPClient = server.Client()
	client.BaseURL = server.URL

	return client
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

	client := setupTest(t, responseBody, http.StatusOK)

	res, err := client.ChangeResourceRecordSets("example.com", ChangeResourceRecordSetsRequest{})
	require.NoError(t, err)

	assert.Equal(t, "xxxxx", res.ChangeInfo.ID)
	assert.Equal(t, "INSYNC", res.ChangeInfo.Status)
	assert.Equal(t, "2015-08-05T00:00:00.000Z", res.ChangeInfo.SubmittedAt)
}

func TestChangeResourceRecordSetsErrors(t *testing.T) {
	testCases := []struct {
		desc         string
		responseBody string
		statusCode   int
		expected     string
	}{
		{
			desc: "API error",
			responseBody: `<?xml version="1.0" encoding="UTF-8"?>
<ErrorResponse>
  <Error>
    <Type>Sender</Type>
    <Code>AuthFailed</Code>
    <Message>The request signature we calculated does not match the signature you provided.</Message>
  </Error>
</ErrorResponse>
`,
			statusCode: http.StatusUnauthorized,
			expected:   "an error occurred: The request signature we calculated does not match the signature you provided.",
		},
		{
			desc:         "response body error",
			responseBody: "foo",
			statusCode:   http.StatusOK,
			expected:     "an error occurred while unmarshaling the response body to XML: EOF",
		},
		{
			desc:         "error message error",
			responseBody: "foo",
			statusCode:   http.StatusInternalServerError,
			expected:     "an error occurred while unmarshaling the error body to XML: EOF",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			client := setupTest(t, test.responseBody, test.statusCode)

			res, err := client.ChangeResourceRecordSets("example.com", ChangeResourceRecordSetsRequest{})
			assert.Nil(t, res)
			assert.EqualError(t, err, test.expected)
		})
	}
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

	client := setupTest(t, responseBody, http.StatusOK)

	res, err := client.GetChange("12345")
	require.NoError(t, err)

	assert.Equal(t, "xxxxx", res.ChangeInfo.ID)
	assert.Equal(t, "INSYNC", res.ChangeInfo.Status)
	assert.Equal(t, "2015-08-05T00:00:00.000Z", res.ChangeInfo.SubmittedAt)
}

func TestGetChangeErrors(t *testing.T) {
	testCases := []struct {
		desc         string
		responseBody string
		statusCode   int
		expected     string
	}{
		{
			desc: "API error",
			responseBody: `<?xml version="1.0" encoding="UTF-8"?>
<ErrorResponse>
  <Error>
    <Type>Sender</Type>
    <Code>AuthFailed</Code>
    <Message>The request signature we calculated does not match the signature you provided.</Message>
  </Error>
</ErrorResponse>
`,
			statusCode: http.StatusUnauthorized,
			expected:   "an error occurred: The request signature we calculated does not match the signature you provided.",
		},
		{
			desc:         "response body error",
			responseBody: "foo",
			statusCode:   http.StatusOK,
			expected:     "an error occurred while unmarshaling the response body to XML: EOF",
		},
		{
			desc:         "error message error",
			responseBody: "foo",
			statusCode:   http.StatusInternalServerError,
			expected:     "an error occurred while unmarshaling the error body to XML: EOF",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			client := setupTest(t, test.responseBody, test.statusCode)

			res, err := client.GetChange("12345")
			assert.Nil(t, res)
			assert.EqualError(t, err, test.expected)
		})
	}
}
