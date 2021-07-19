package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

const (
	testToken = "test"
)

func clientForTest() (*http.ServeMux, *Client, func()) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	client := NewClient(testToken, func(client *Client) {
		client.BaseURL = server.URL
	})

	return mux, client, server.Close
}

func Test_extractAllZones(t *testing.T) {
	type args struct {
		fqdn string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "success",
			args: args{
				fqdn: "_acme-challenge.my.test.domain.com.",
			},
			want: []string{"my.test.domain.com", "test.domain.com", "domain.com"},
		},
		{
			name: "empty",
			args: args{
				fqdn: "_acme-challenge.com.",
			},
			want: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractAllZones(tt.args.fqdn); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractAllZones() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewClient(t *testing.T) {
	type args struct {
		token string
		opts  []ClientOpt
	}
	tests := []struct {
		name string
		args args
		want Client
	}{
		{
			name: "without opts",
			args: args{
				token: "1",
				opts:  nil,
			},
			want: Client{
				HTTPClient: &http.Client{},
				BaseURL:    defaultBaseURL,
				Token:      "1",
			},
		},
		{
			name: "with opts",
			args: args{
				token: "1",
				opts: []ClientOpt{
					func(client *Client) {
						client.BaseURL = "2"
					},
					func(client *Client) {
						client.HTTPClient.Timeout = time.Second
					},
				},
			},
			want: Client{
				HTTPClient: &http.Client{
					Timeout: time.Second,
				},
				BaseURL: "2",
				Token:   "1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewClient(tt.args.token, tt.args.opts...); !reflect.DeepEqual(*got, tt.want) {
				t.Errorf("NewClient() = %+v, want %+v", *got, tt.want)
			}
		})
	}
}

func validRequest(w http.ResponseWriter, r *http.Request, waitedMethod string, body interface{}) (ok bool) {
	if r.Header.Get("Authorization") != fmt.Sprintf("%s %s", tokenHeader, testToken) {
		http.Error(w, "wrong token", http.StatusForbidden)
		return false
	}
	if r.Method != waitedMethod {
		http.Error(w, "wrong method", http.StatusForbidden)
		return false
	}
	if body == nil {
		return true
	}
	defer func() { _ = r.Body.Close() }()
	err := json.NewDecoder(r.Body).Decode(body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return false
	}
	return true
}

func sendResponse(w http.ResponseWriter, resp interface{}) {
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func enrichMuxWithFindZone(mux *http.ServeMux, path, result string) {
	mux.HandleFunc(
		path,
		func(w http.ResponseWriter, r *http.Request) {
			if !validRequest(w, r, http.MethodGet, nil) {
				return
			}
			sendResponse(w, getZoneResponse{Name: result})
		})
}

func enrichMuxWithAddZoneRecord(mux *http.ServeMux, path string) {
	mux.HandleFunc(
		path,
		func(w http.ResponseWriter, r *http.Request) {
			body := addRecordRequest{}
			if !validRequest(w, r, http.MethodPost, &body) {
				return
			}
			if body.TTL != 10 {
				http.Error(w, "wrong ttl", http.StatusInternalServerError)
				return
			}
			if !reflect.DeepEqual(body.ResourceRecords, []resourceRecord{{Content: []string{"acme"}}}) {
				http.Error(w, "wrong resource records", http.StatusInternalServerError)
				return
			}
		})
}

func TestClient_AddTXTRecord(t *testing.T) {
	type args struct {
		ctx   context.Context
		fqdn  string
		value string
		ttl   int
	}
	tests := []struct {
		name         string
		clientGetter func() (*Client, func())
		args         args
		wantErr      bool
	}{
		{
			name: "success",
			clientGetter: func() (*Client, func()) {
				mux, cl, cancel := clientForTest()
				enrichMuxWithFindZone(mux, "/v2/zones/test.domain.com", "test.domain.com")
				enrichMuxWithAddZoneRecord(mux,
					"/v2/zones/test.domain.com/_acme-challenge.my.test.domain.com/"+recordType)
				return cl, cancel
			},
			args: args{
				ctx:   context.Background(),
				fqdn:  "_acme-challenge.my.test.domain.com.",
				value: "acme",
				ttl:   10,
			},
			wantErr: false,
		},
		{
			name: "no zone",
			clientGetter: func() (*Client, func()) {
				mux, cl, cancel := clientForTest()
				enrichMuxWithFindZone(mux, "/v2/zones/not.found.com", "not.found.com")
				enrichMuxWithAddZoneRecord(mux,
					"/v2/zones/test.domain.com/_acme-challenge.my.test.domain.com/"+recordType)
				return cl, cancel
			},
			args: args{
				ctx:   context.Background(),
				fqdn:  "_acme-challenge.my.test.domain.com.",
				value: "acme",
				ttl:   10,
			},
			wantErr: true,
		},
		{
			name: "no add",
			clientGetter: func() (*Client, func()) {
				mux, cl, cancel := clientForTest()
				enrichMuxWithFindZone(mux, "/v2/zones/domain.com", "domain.com")
				enrichMuxWithAddZoneRecord(mux,
					"/v2/zones/test.domain.com/_acme-challenge.my.test.domain.com/"+recordType)
				return cl, cancel
			},
			args: args{
				ctx:   context.Background(),
				fqdn:  "_acme-challenge.my.test.domain.com.",
				value: "acme",
				ttl:   10,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl, cancel := tt.clientGetter()
			defer cancel()
			if err := cl.AddTXTRecord(tt.args.ctx, tt.args.fqdn, tt.args.value, tt.args.ttl); (err != nil) != tt.wantErr {
				t.Errorf("AddTXTRecord() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_requestUrl(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name   string
		client *Client
		args   args
		want   string
	}{
		{
			name:   "no trim",
			client: &Client{BaseURL: defaultBaseURL},
			args: args{
				path: "path",
			},
			want: defaultBaseURL + "/path",
		},
		{
			name:   "base url trim",
			client: &Client{BaseURL: "http/"},
			args: args{
				path: "path",
			},
			want: "http/path",
		},
		{
			name:   "booth trim",
			client: &Client{BaseURL: "http/"},
			args: args{
				path: "/path",
			},
			want: "http/path",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.client
			if got := c.requestURL(tt.args.path); got != tt.want {
				t.Errorf("requestUrl() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_RemoveTXTRecord(t *testing.T) {
	type args struct {
		ctx  context.Context
		fqdn string
		txt  string
	}
	tests := []struct {
		name         string
		clientGetter func() (*Client, func())
		args         args
		wantErr      bool
	}{
		{
			name: "success",
			clientGetter: func() (*Client, func()) {
				mux, cl, cancel := clientForTest()
				enrichMuxWithFindZone(mux, "/v2/zones/test.domain.com", "test.domain.com")
				mux.HandleFunc(
					"/v2/zones/test.domain.com/_acme-challenge.my.test.domain.com/"+recordType,
					func(w http.ResponseWriter, r *http.Request) {
						if !validRequest(w, r, http.MethodDelete, nil) {
							return
						}
					})
				return cl, cancel
			},
			args: args{
				ctx:  context.Background(),
				fqdn: "_acme-challenge.my.test.domain.com.",
				txt:  "acme",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl, cancel := tt.clientGetter()
			defer cancel()
			if err := cl.RemoveTXTRecord(tt.args.ctx, tt.args.fqdn, tt.args.txt); (err != nil) != tt.wantErr {
				t.Errorf("RemoveTXTRecord() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
