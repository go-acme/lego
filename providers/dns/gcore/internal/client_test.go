package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testToken          = "test"
	testRecordContent  = "acme"
	testRecordContent2 = "foo"
	testTTL            = 10
)

func setupTest(t *testing.T) (*http.ServeMux, *Client) {
	t.Helper()

	mux := http.NewServeMux()

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := NewClient(testToken)
	client.baseURL, _ = url.Parse(server.URL)

	return mux, client
}

func TestClient_GetZone(t *testing.T) {
	mux, client := setupTest(t)

	expected := Zone{Name: "example.com"}

	mux.Handle("/v2/zones/example.com", validationHandler{
		method: http.MethodGet,
		next:   handleJSONResponse(expected),
	})

	zone, err := client.GetZone(context.Background(), "example.com")
	require.NoError(t, err)

	assert.Equal(t, expected, zone)
}

func TestClient_GetZone_error(t *testing.T) {
	mux, client := setupTest(t)

	mux.Handle("/v2/zones/example.com", validationHandler{
		method: http.MethodGet,
		next:   handleAPIError(),
	})

	_, err := client.GetZone(context.Background(), "example.com")
	require.Error(t, err)
}

func TestClient_GetRRSet(t *testing.T) {
	mux, client := setupTest(t)

	expected := RRSet{
		TTL: testTTL,
		Records: []Records{
			{Content: []string{testRecordContent}},
		},
	}

	mux.Handle("/v2/zones/example.com/foo.example.com/TXT", validationHandler{
		method: http.MethodGet,
		next:   handleJSONResponse(expected),
	})

	rrSet, err := client.GetRRSet(context.Background(), "example.com", "foo.example.com")
	require.NoError(t, err)

	assert.Equal(t, expected, rrSet)
}

func TestClient_GetRRSet_error(t *testing.T) {
	mux, client := setupTest(t)

	mux.Handle("/v2/zones/example.com/foo.example.com/TXT", validationHandler{
		method: http.MethodGet,
		next:   handleAPIError(),
	})

	_, err := client.GetRRSet(context.Background(), "example.com", "foo.example.com")
	require.Error(t, err)
}

func TestClient_DeleteRRSet(t *testing.T) {
	mux, client := setupTest(t)

	mux.Handle("/v2/zones/test.example.com/my.test.example.com/"+txtRecordType,
		validationHandler{method: http.MethodDelete})

	err := client.DeleteRRSet(context.Background(), "test.example.com", "my.test.example.com.")
	require.NoError(t, err)
}

func TestClient_DeleteRRSet_error(t *testing.T) {
	mux, client := setupTest(t)

	mux.Handle("/v2/zones/test.example.com/my.test.example.com/"+txtRecordType, validationHandler{
		method: http.MethodDelete,
		next:   handleAPIError(),
	})

	err := client.DeleteRRSet(context.Background(), "test.example.com", "my.test.example.com.")
	require.NoError(t, err)
}

func TestClient_AddRRSet(t *testing.T) {
	testCases := []struct {
		desc          string
		zone          string
		recordName    string
		value         string
		handledDomain string
		handlers      map[string]http.Handler
		wantErr       bool
	}{
		{
			desc:       "success add",
			zone:       "test.example.com",
			recordName: "my.test.example.com",
			value:      testRecordContent,
			handlers: map[string]http.Handler{
				// createRRSet
				"/v2/zones/test.example.com/my.test.example.com/" + txtRecordType: validationHandler{
					method: http.MethodPost,
					next:   handleAddRRSet([]Records{{Content: []string{testRecordContent}}}),
				},
			},
		},
		{
			desc:       "success update",
			zone:       "test.example.com",
			recordName: "my.test.example.com",
			value:      testRecordContent,
			handlers: map[string]http.Handler{
				"/v2/zones/test.example.com/my.test.example.com/" + txtRecordType: http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
					switch req.Method {
					case http.MethodGet: // GetRRSet
						data := RRSet{
							TTL:     testTTL,
							Records: []Records{{Content: []string{testRecordContent2}}},
						}
						handleJSONResponse(data).ServeHTTP(rw, req)
					case http.MethodPut: // updateRRSet
						expected := []Records{
							{Content: []string{testRecordContent}},
							{Content: []string{testRecordContent2}},
						}
						handleAddRRSet(expected).ServeHTTP(rw, req)
					default:
						http.Error(rw, "wrong method", http.StatusMethodNotAllowed)
					}
				}),
			},
		},
		{
			desc:       "not in the zone",
			zone:       "test.example.com",
			recordName: "notfound.example.com",
			value:      testRecordContent,
			wantErr:    true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			mux, cl := setupTest(t)

			for pattern, handler := range test.handlers {
				mux.Handle(pattern, handler)
			}

			err := cl.AddRRSet(context.Background(), test.zone, test.recordName, test.value, testTTL)
			if test.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

type validationHandler struct {
	method string
	next   http.Handler
}

func (v validationHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if req.Header.Get("Authorization") != fmt.Sprintf("%s %s", tokenHeader, testToken) {
		rw.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(rw).Encode(APIError{Message: "token up for parsing was not passed through the context"})
		return
	}

	if req.Method != v.method {
		http.Error(rw, "wrong method", http.StatusMethodNotAllowed)
		return
	}

	if v.next != nil {
		v.next.ServeHTTP(rw, req)
	}
}

func handleAPIError() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(rw).Encode(APIError{Message: "oops"})
	}
}

func handleJSONResponse(data interface{}) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		err := json.NewEncoder(rw).Encode(data)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func handleAddRRSet(expected []Records) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		body := RRSet{}

		err := json.NewDecoder(req.Body).Decode(&body)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		if body.TTL != testTTL {
			http.Error(rw, "wrong ttl", http.StatusInternalServerError)
			return
		}

		if !reflect.DeepEqual(body.Records, expected) {
			http.Error(rw, "wrong resource records", http.StatusInternalServerError)
		}
	}
}
