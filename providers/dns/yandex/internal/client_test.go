package internal

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T) (*http.ServeMux, *Client) {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client, err := NewClient("lego")
	require.NoError(t, err)

	client.HTTPClient = server.Client()
	client.BaseURL = server.URL

	return mux, client
}

func TestAddRecord(t *testing.T) {
	testCases := []struct {
		desc        string
		handler     http.HandlerFunc
		data        Record
		expectError bool
	}{
		{
			desc: "success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "lego", r.Header.Get(pddTokenHeader))

				err := r.ParseForm()
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}

				assert.Equal(t, `content=txtTXTtxtTXTtxtTXT&domain=example.com&subdomain=foo&ttl=300&type=TXT`, r.PostForm.Encode())

				response := AddResponse{
					Domain: "example.com",
					Record: &Record{
						ID:        1,
						Type:      "TXT",
						Domain:    "example.com",
						SubDomain: "foo",
						FQDN:      "foo.example.com.",
						Content:   "txtTXTtxtTXTtxtTXT",
						TTL:       300,
					},
					Success: "ok",
				}

				err = json.NewEncoder(w).Encode(response)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			},
			data: Record{
				Domain:    "example.com",
				Type:      "TXT",
				Content:   "txtTXTtxtTXTtxtTXT",
				SubDomain: "foo",
				TTL:       300,
			},
		},
		{
			desc: "error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "lego", r.Header.Get(pddTokenHeader))

				err := r.ParseForm()
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}

				assert.Equal(t, `content=txtTXTtxtTXTtxtTXT&domain=example.com&subdomain=foo&ttl=300&type=TXT`, r.PostForm.Encode())

				response := AddResponse{
					Domain:  "example.com",
					Success: "error",
					Error:   "bad things",
				}

				err = json.NewEncoder(w).Encode(response)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			},
			data: Record{
				Domain:    "example.com",
				Type:      "TXT",
				Content:   "txtTXTtxtTXTtxtTXT",
				SubDomain: "foo",
				TTL:       300,
			},
			expectError: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			mux, client := setupTest(t)

			mux.HandleFunc("/add", test.handler)

			record, err := client.AddRecord(test.data)
			if test.expectError {
				require.Error(t, err)
				require.Nil(t, record)
			} else {
				require.NoError(t, err)
				require.NotNil(t, record)
			}
		})
	}
}

func TestRemoveRecord(t *testing.T) {
	testCases := []struct {
		desc        string
		handler     http.HandlerFunc
		data        Record
		expectError bool
	}{
		{
			desc: "success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "lego", r.Header.Get(pddTokenHeader))

				err := r.ParseForm()
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}

				assert.Equal(t, `domain=example.com&record_id=6`, r.PostForm.Encode())

				response := RemoveResponse{
					Domain:   "example.com",
					RecordID: 6,
					Success:  "ok",
				}

				err = json.NewEncoder(w).Encode(response)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			},
			data: Record{
				ID:     6,
				Domain: "example.com",
			},
		},
		{
			desc: "error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "lego", r.Header.Get(pddTokenHeader))

				err := r.ParseForm()
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}

				assert.Equal(t, `domain=example.com&record_id=6`, r.PostForm.Encode())

				response := RemoveResponse{
					Domain:   "example.com",
					RecordID: 6,
					Success:  "error",
					Error:    "bad things",
				}

				err = json.NewEncoder(w).Encode(response)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			},
			data: Record{
				ID:     6,
				Domain: "example.com",
			},
			expectError: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			mux, client := setupTest(t)

			mux.HandleFunc("/del", test.handler)

			id, err := client.RemoveRecord(test.data)
			if test.expectError {
				require.Error(t, err)
				require.Equal(t, 0, id)
			} else {
				require.NoError(t, err)
				require.Equal(t, 6, id)
			}
		})
	}
}

func TestGetRecords(t *testing.T) {
	testCases := []struct {
		desc        string
		handler     http.HandlerFunc
		domain      string
		expectError bool
	}{
		{
			desc: "success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, "lego", r.Header.Get(pddTokenHeader))

				assert.Equal(t, "domain=example.com", r.URL.RawQuery)

				response := ListResponse{
					Domain: "example.com",
					Records: []Record{
						{
							ID:        1,
							Type:      "TXT",
							Domain:    "example.com",
							SubDomain: "foo",
							FQDN:      "foo.example.com.",
							Content:   "txtTXTtxtTXTtxtTXT",
							TTL:       300,
						},
						{
							ID:        2,
							Type:      "NS",
							Domain:    "example.com",
							SubDomain: "foo",
							FQDN:      "foo.example.com.",
							Content:   "bar",
							TTL:       300,
						},
					},
					Success: "ok",
				}

				err := json.NewEncoder(w).Encode(response)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			},
			domain: "example.com",
		},
		{
			desc: "error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, "lego", r.Header.Get(pddTokenHeader))

				assert.Equal(t, "domain=example.com", r.URL.RawQuery)

				response := ListResponse{
					Domain:  "example.com",
					Success: "error",
					Error:   "bad things",
				}

				err := json.NewEncoder(w).Encode(response)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			},
			domain:      "example.com",
			expectError: true,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			mux, client := setupTest(t)

			mux.HandleFunc("/list", test.handler)

			records, err := client.GetRecords(test.domain)
			if test.expectError {
				require.Error(t, err)
				require.Empty(t, records)
			} else {
				require.NoError(t, err)
				require.Len(t, records, 2)
			}
		})
	}
}
