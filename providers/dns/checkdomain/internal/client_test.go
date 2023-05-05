package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T) (*Client, *http.ServeMux) {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := NewClient(OAuthStaticAccessToken(server.Client(), "secret"))
	client.BaseURL, _ = url.Parse(server.URL)

	return client, mux
}

func checkAuthorizationHeader(req *http.Request) error {
	val := req.Header.Get("Authorization")
	if val != "Bearer secret" {
		return fmt.Errorf("invalid header value, got: %s want %s", val, "Bearer secret")
	}
	return nil
}

func TestClient_GetDomainIDByName(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/v1/domains", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(rw, "invalid method: "+req.Method, http.StatusBadRequest)
			return
		}

		err := checkAuthorizationHeader(req)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusUnauthorized)
			return
		}

		domainList := DomainListingResponse{
			Embedded: EmbeddedDomainList{Domains: []*Domain{
				{ID: 1, Name: "test.com"},
				{ID: 2, Name: "test.org"},
			}},
		}

		err = json.NewEncoder(rw).Encode(domainList)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	id, err := client.GetDomainIDByName(context.Background(), "test.com")
	require.NoError(t, err)

	assert.Equal(t, 1, id)
}

func TestClient_CheckNameservers(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/v1/domains/1/nameservers", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(rw, "invalid method: "+req.Method, http.StatusBadRequest)
			return
		}

		err := checkAuthorizationHeader(req)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusUnauthorized)
			return
		}

		nsResp := NameserverResponse{
			Nameservers: []*Nameserver{
				{Name: ns1},
				{Name: ns2},
				// {Name: "ns.fake.de"},
			},
		}

		err = json.NewEncoder(rw).Encode(nsResp)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	err := client.CheckNameservers(context.Background(), 1)
	require.NoError(t, err)
}

func TestClient_CreateRecord(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("/v1/domains/1/nameservers/records", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(rw, "invalid method: "+req.Method, http.StatusBadRequest)
			return
		}

		err := checkAuthorizationHeader(req)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusUnauthorized)
			return
		}

		content, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		if string(bytes.TrimSpace(content)) != `{"name":"test.com","value":"value","ttl":300,"priority":0,"type":"TXT"}` {
			http.Error(rw, "invalid request body: "+string(content), http.StatusBadRequest)
			return
		}
	})

	record := &Record{
		Name:  "test.com",
		TTL:   300,
		Type:  "TXT",
		Value: "value",
	}

	err := client.CreateRecord(context.Background(), 1, record)
	require.NoError(t, err)
}

func TestClient_DeleteTXTRecord(t *testing.T) {
	client, mux := setupTest(t)

	domainName := "lego.test"
	recordValue := "test"

	records := []*Record{
		{
			Name:  "_acme-challenge",
			Value: recordValue,
			Type:  "TXT",
		},
		{
			Name:  "_acme-challenge",
			Value: recordValue,
			Type:  "A",
		},
		{
			Name:  "foobar",
			Value: recordValue,
			Type:  "TXT",
		},
	}

	expectedRecords := []*Record{
		{
			Name:  "_acme-challenge",
			Value: recordValue,
			Type:  "A",
		},
		{
			Name:  "foobar",
			Value: recordValue,
			Type:  "TXT",
		},
	}

	mux.HandleFunc("/v1/domains/1", func(rw http.ResponseWriter, req *http.Request) {
		err := checkAuthorizationHeader(req)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusUnauthorized)
			return
		}

		resp := DomainResponse{
			ID:   1,
			Name: domainName,
		}

		err = json.NewEncoder(rw).Encode(resp)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	mux.HandleFunc("/v1/domains/1/nameservers", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(rw, "invalid method: "+req.Method, http.StatusBadRequest)
			return
		}

		nsResp := NameserverResponse{
			Nameservers: []*Nameserver{{Name: ns1}, {Name: ns2}},
		}

		err := json.NewEncoder(rw).Encode(nsResp)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	mux.HandleFunc("/v1/domains/1/nameservers/records", func(rw http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodGet:
			resp := RecordListingResponse{
				Embedded: EmbeddedRecordList{
					Records: records,
				},
			}

			err := json.NewEncoder(rw).Encode(resp)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				return
			}

		case http.MethodPut:
			var records []*Record
			err := json.NewDecoder(req.Body).Decode(&records)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusBadRequest)
				return
			}

			if len(records) == 0 {
				http.Error(rw, "empty request body", http.StatusBadRequest)
				return
			}

			if !reflect.DeepEqual(expectedRecords, records) {
				http.Error(rw, fmt.Sprintf("invalid records: %v", records), http.StatusBadRequest)
				return
			}
		default:
			http.Error(rw, "invalid method: "+req.Method, http.StatusBadRequest)
		}
	})

	info := dns01.GetChallengeInfo(domainName, "abc")
	err := client.DeleteTXTRecord(context.Background(), 1, info.EffectiveFQDN, recordValue)
	require.NoError(t, err)
}
