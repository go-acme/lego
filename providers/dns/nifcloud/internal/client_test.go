package internal

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client, err := NewClient("A", "B")
			if err != nil {
				return nil, err
			}

			client.HTTPClient = server.Client()
			client.BaseURL, _ = url.Parse(server.URL)

			return client, nil
		},
		servermock.CheckHeader().
			WithRegexp("X-Nifty-Authorization", "NIFTY3-HTTPS NiftyAccessKeyId=A,Algorithm=HmacSHA1,Signature=.+"),
	)
}

func TestClient_ChangeResourceRecordSets(t *testing.T) {
	responseBody := `<?xml version="1.0" encoding="UTF-8"?>
<ChangeResourceRecordSetsResponse xmlns="https://route53.amazonaws.com/doc/2012-12-12/">
  <ChangeInfo>
    <Id>xxxxx</Id>
    <Status>INSYNC</Status>
    <SubmittedAt>2015-08-05T00:00:00.000Z</SubmittedAt>
  </ChangeInfo>
</ChangeResourceRecordSetsResponse>
`

	client := mockBuilder().
		Route("POST /", servermock.RawStringResponse(responseBody),
			servermock.CheckHeader().WithContentType("text/xml; charset=utf-8")).
		Build(t)

	res, err := client.ChangeResourceRecordSets(t.Context(), "example.com", ChangeResourceRecordSetsRequest{})
	require.NoError(t, err)

	assert.Equal(t, "xxxxx", res.ChangeInfo.ID)
	assert.Equal(t, "INSYNC", res.ChangeInfo.Status)
	assert.Equal(t, "2015-08-05T00:00:00.000Z", res.ChangeInfo.SubmittedAt)
}

func TestClient_ChangeResourceRecordSets_errors(t *testing.T) {
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
			expected:   "Sender(AuthFailed): The request signature we calculated does not match the signature you provided.",
		},
		{
			desc:         "response body error",
			responseBody: "foo",
			statusCode:   http.StatusOK,
			expected:     "unable to unmarshal response: [status code: 200] body: foo error: EOF",
		},
		{
			desc:         "error message error",
			responseBody: "foo",
			statusCode:   http.StatusInternalServerError,
			expected:     "unexpected status code: [status code: 500] body: foo",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			client := mockBuilder().
				Route("POST /",
					servermock.RawStringResponse(test.responseBody).
						WithStatusCode(test.statusCode),
					servermock.CheckHeader().
						WithContentType("text/xml; charset=utf-8")).
				Build(t)

			res, err := client.ChangeResourceRecordSets(t.Context(), "example.com", ChangeResourceRecordSetsRequest{})
			assert.Nil(t, res)
			assert.EqualError(t, err, test.expected)
		})
	}
}

func TestClient_GetChange(t *testing.T) {
	responseBody := `<?xml version="1.0" encoding="UTF-8"?>
<GetChangeResponse xmlns="https://route53.amazonaws.com/doc/2012-12-12/">
  <ChangeInfo>
    <Id>xxxxx</Id>
    <Status>INSYNC</Status>
    <SubmittedAt>2015-08-05T00:00:00.000Z</SubmittedAt>
  </ChangeInfo>
</GetChangeResponse>
`

	client := mockBuilder().
		Route("GET /", servermock.RawStringResponse(responseBody)).
		Build(t)

	res, err := client.GetChange(t.Context(), "12345")
	require.NoError(t, err)

	assert.Equal(t, "xxxxx", res.ChangeInfo.ID)
	assert.Equal(t, "INSYNC", res.ChangeInfo.Status)
	assert.Equal(t, "2015-08-05T00:00:00.000Z", res.ChangeInfo.SubmittedAt)
}

func TestClient_GetChange_errors(t *testing.T) {
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
			expected:   "Sender(AuthFailed): The request signature we calculated does not match the signature you provided.",
		},
		{
			desc:         "response body error",
			responseBody: "foo",
			statusCode:   http.StatusOK,
			expected:     "unable to unmarshal response: [status code: 200] body: foo error: EOF",
		},
		{
			desc:         "error message error",
			responseBody: "foo",
			statusCode:   http.StatusInternalServerError,
			expected:     "unexpected status code: [status code: 500] body: foo",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			client := mockBuilder().
				Route("GET /",
					servermock.RawStringResponse(test.responseBody).WithStatusCode(test.statusCode)).
				Build(t)

			res, err := client.GetChange(t.Context(), "12345")
			assert.Nil(t, res)
			assert.EqualError(t, err, test.expected)
		})
	}
}
