package checkdomain

import (
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

func setupTestProvider(t *testing.T) (*DNSProvider, *http.ServeMux) {
	t.Helper()

	handler := http.NewServeMux()
	svr := httptest.NewServer(handler)

	t.Cleanup(svr.Close)

	config := NewDefaultConfig()
	config.Endpoint, _ = url.Parse(svr.URL)
	config.Token = "secret"

	prd, err := NewDNSProviderConfig(config)
	require.NoError(t, err)

	return prd, handler
}

func Test_getDomainIDByName(t *testing.T) {
	prd, handler := setupTestProvider(t)

	handler.HandleFunc("/v1/domains", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(rw, "invalid method: "+req.Method, http.StatusBadRequest)
			return
		}

		domainList := DomainListingResponse{
			Embedded: EmbeddedDomainList{Domains: []*Domain{
				{ID: 1, Name: "test.com"},
				{ID: 2, Name: "test.org"},
			}},
		}

		err := json.NewEncoder(rw).Encode(domainList)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	id, err := prd.getDomainIDByName("test.com")
	require.NoError(t, err)

	assert.Equal(t, 1, id)
}

func Test_checkNameservers(t *testing.T) {
	prd, handler := setupTestProvider(t)

	handler.HandleFunc("/v1/domains/1/nameservers", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(rw, "invalid method: "+req.Method, http.StatusBadRequest)
			return
		}

		nsResp := NameserverResponse{
			Nameservers: []*Nameserver{
				{Name: ns1},
				{Name: ns2},
				// {Name: "ns.fake.de"},
			},
		}

		err := json.NewEncoder(rw).Encode(nsResp)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	err := prd.checkNameservers(1)
	require.NoError(t, err)
}

func Test_createRecord(t *testing.T) {
	prd, handler := setupTestProvider(t)

	handler.HandleFunc("/v1/domains/1/nameservers/records", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(rw, "invalid method: "+req.Method, http.StatusBadRequest)
			return
		}

		content, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		if string(content) != `{"name":"test.com","value":"value","ttl":300,"priority":0,"type":"TXT"}` {
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

	err := prd.createRecord(1, record)
	require.NoError(t, err)
}

func Test_deleteTXTRecord(t *testing.T) {
	prd, handler := setupTestProvider(t)

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

	handler.HandleFunc("/v1/domains/1", func(rw http.ResponseWriter, req *http.Request) {
		resp := DomainResponse{
			ID:   1,
			Name: domainName,
		}

		err := json.NewEncoder(rw).Encode(resp)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	handler.HandleFunc("/v1/domains/1/nameservers", func(rw http.ResponseWriter, req *http.Request) {
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

	handler.HandleFunc("/v1/domains/1/nameservers/records", func(rw http.ResponseWriter, req *http.Request) {
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

	fqdn, _ := dns01.GetRecord(domainName, "abc")
	err := prd.deleteTXTRecord(1, fqdn, recordValue)
	require.NoError(t, err)
}
