package internal

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T) (*Client, *http.ServeMux) {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client, err := NewClient("tyo1", "secret")
	require.NoError(t, err)

	client.HTTPClient = server.Client()
	client.baseURL, _ = url.Parse(server.URL)

	return client, mux
}

func writeFixtureHandler(method, filename string) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != method {
			http.Error(rw, fmt.Sprintf("unsupported method %s", req.Method), http.StatusBadRequest)
			return
		}

		writeFixture(rw, filename)
	}
}

func writeBodyHandler(method, content string) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != method {
			http.Error(rw, fmt.Sprintf("unsupported method %s", req.Method), http.StatusBadRequest)
			return
		}
		_, err := fmt.Fprint(rw, content)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func writeFixture(rw http.ResponseWriter, filename string) {
	file, err := os.Open(filepath.Join("fixtures", filename))
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	defer func() { _ = file.Close() }()

	_, _ = io.Copy(rw, file)
}

func TestClient_GetDomainID(t *testing.T) {
	type expected struct {
		domainID string
		error    bool
	}

	testCases := []struct {
		desc       string
		domainName string
		handler    http.HandlerFunc
		expected   expected
	}{
		{
			desc:       "success",
			domainName: "domain1.com.",
			handler:    writeFixtureHandler(http.MethodGet, "domains_GET.json"),
			expected:   expected{domainID: "09494b72-b65b-4297-9efb-187f65a0553e"},
		},
		{
			desc:       "non existing domain",
			domainName: "domain1.com.",
			handler:    writeBodyHandler(http.MethodGet, "{}"),
			expected:   expected{error: true},
		},
		{
			desc:       "marshaling error",
			domainName: "domain1.com.",
			handler:    writeBodyHandler(http.MethodGet, "[]"),
			expected:   expected{error: true},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			client, mux := setupTest(t)

			mux.Handle("/v1/domains", test.handler)

			domainID, err := client.GetDomainID(context.Background(), test.domainName)

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
				if req.Method != http.MethodPost {
					http.Error(rw, fmt.Sprintf("unsupported method %s", req.Method), http.StatusBadRequest)
					return
				}

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

				writeFixture(rw, "domains-records_POST.json")
			},
			assert: require.NoError,
		},
		{
			desc: "bad request",
			handler: func(rw http.ResponseWriter, req *http.Request) {
				if req.Method != http.MethodPost {
					http.Error(rw, fmt.Sprintf("unsupported method %s", req.Method), http.StatusBadRequest)
					return
				}

				http.Error(rw, "OOPS", http.StatusBadRequest)
			},
			assert: require.Error,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			client, mux := setupTest(t)

			mux.Handle("/v1/domains/lego/records", test.handler)

			domainID := "lego"

			record := Record{
				Name: "lego.com.",
				Type: "TXT",
				Data: "txtTXTtxt",
				TTL:  300,
			}

			err := client.CreateRecord(context.Background(), domainID, record)
			test.assert(t, err)
		})
	}
}

func TestClient_GetRecordID(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/v1/domains/89acac79-38e7-497d-807c-a011e1310438/records",
		writeFixtureHandler(http.MethodGet, "domains-records_GET.json"))

	recordID, err := client.GetRecordID(context.Background(), "89acac79-38e7-497d-807c-a011e1310438", "www.example.com.", "A", "15.185.172.153")
	require.NoError(t, err)

	assert.Equal(t, "2e32e609-3a4f-45ba-bdef-e50eacd345ad", recordID)
}

func TestClient_DeleteRecord(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/v1/domains/89acac79-38e7-497d-807c-a011e1310438/records/2e32e609-3a4f-45ba-bdef-e50eacd345ad", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodDelete {
			http.Error(rw, fmt.Sprintf("unsupported method %s", req.Method), http.StatusBadRequest)
			return
		}

		rw.WriteHeader(http.StatusOK)
	})

	err := client.DeleteRecord(context.Background(), "89acac79-38e7-497d-807c-a011e1310438", "2e32e609-3a4f-45ba-bdef-e50eacd345ad")
	require.NoError(t, err)
}
