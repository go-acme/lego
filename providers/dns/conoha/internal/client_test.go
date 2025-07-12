package internal

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockBuilder() *servermock.Builder[*Client] {
	return servermock.NewBuilder[*Client](
		func(server *httptest.Server) (*Client, error) {
			client, err := NewClient("tyo1", "secret")
			if err != nil {
				return nil, err
			}

			client.HTTPClient = server.Client()
			client.baseURL, _ = url.Parse(server.URL)

			return client, nil
		},
		servermock.CheckHeader().WithJSONHeaders().
			With("X-Auth-Token", "secret"))
}

func TestClient_GetDomainID(t *testing.T) {
	type expected struct {
		domainID string
		error    bool
	}

	testCases := []struct {
		desc       string
		domainName string
		response   string
		expected   expected
	}{
		{
			desc:       "success",
			domainName: "domain1.com.",
			response:   "domains_GET.json",
			expected:   expected{domainID: "09494b72-b65b-4297-9efb-187f65a0553e"},
		},
		{
			desc:       "non existing domain",
			domainName: "domain1.com.",
			response:   "empty.json",
			expected:   expected{error: true},
		},
		{
			desc:       "marshaling error",
			domainName: "domain1.com.",
			response:   "empty.json",
			expected:   expected{error: true},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			client := mockBuilder().
				Route("GET /v1/domains", servermock.ResponseFromFixture(test.response)).
				Build(t)

			domainID, err := client.GetDomainID(t.Context(), test.domainName)

			if test.expected.error {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expected.domainID, domainID)
			}
		})
	}
}

func TestClient_CreateRecord(t *testing.T) {
	testCases := []struct {
		desc    string
		handler http.HandlerFunc
		assert  require.ErrorAssertionFunc
	}{
		{
			desc: "success",
			handler: func(rw http.ResponseWriter, req *http.Request) {
				raw, err := io.ReadAll(req.Body)
				if err != nil {
					http.Error(rw, err.Error(), http.StatusBadRequest)
					return
				}
				defer func() { _ = req.Body.Close() }()

				if string(bytes.TrimSpace(raw)) != `{"name":"lego.com.","type":"TXT","data":"txtTXTtxt","ttl":300}` {
					http.Error(rw, fmt.Sprintf("invalid request body: %s", string(raw)), http.StatusBadRequest)
					return
				}

				file, err := os.Open(filepath.Join("fixtures", "domains-records_POST.json"))
				if err != nil {
					http.Error(rw, err.Error(), http.StatusInternalServerError)
					return
				}
				defer func() { _ = file.Close() }()

				_, _ = io.Copy(rw, file)
			},
			assert: require.NoError,
		},
		{
			desc: "bad request",
			handler: func(rw http.ResponseWriter, req *http.Request) {
				http.Error(rw, "OOPS", http.StatusBadRequest)
			},
			assert: require.Error,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			client := mockBuilder().
				Route("POST /v1/domains/lego/records", test.handler).
				Build(t)

			domainID := "lego"

			record := Record{
				Name: "lego.com.",
				Type: "TXT",
				Data: "txtTXTtxt",
				TTL:  300,
			}

			err := client.CreateRecord(t.Context(), domainID, record)
			test.assert(t, err)
		})
	}
}

func TestClient_GetRecordID(t *testing.T) {
	client := mockBuilder().
		Route("GET /v1/domains/89acac79-38e7-497d-807c-a011e1310438/records",
			servermock.ResponseFromFixture("domains-records_GET.json")).
		Build(t)

	recordID, err := client.GetRecordID(t.Context(), "89acac79-38e7-497d-807c-a011e1310438", "www.example.com.", "A", "15.185.172.153")
	require.NoError(t, err)

	assert.Equal(t, "2e32e609-3a4f-45ba-bdef-e50eacd345ad", recordID)
}

func TestClient_DeleteRecord(t *testing.T) {
	client := mockBuilder().
		Route("DELETE /v1/domains/89acac79-38e7-497d-807c-a011e1310438/records/2e32e609-3a4f-45ba-bdef-e50eacd345ad",
			servermock.ResponseFromFixture("domains-records_GET.json")).
		Build(t)

	err := client.DeleteRecord(t.Context(), "89acac79-38e7-497d-807c-a011e1310438", "2e32e609-3a4f-45ba-bdef-e50eacd345ad")
	require.NoError(t, err)
}
