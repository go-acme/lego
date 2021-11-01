package sakuracloud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/sacloud/libsacloud/api"
	"github.com/sacloud/libsacloud/sacloud"
	"github.com/stretchr/testify/require"
)

type simpleResponse struct {
	*sacloud.DNS `json:"CommonServiceItem,omitempty"`
}

type apiQuery struct {
	Filter struct {
		Name          string `json:"Name"`
		ProviderClass string `json:"Provider.Class"`
	} `json:"Filter"`
}

func setupTest(t *testing.T, handler http.HandlerFunc) {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	mux.HandleFunc("/is1a/api/cloud/1.1/commonserviceitem/", handler)

	backup := api.SakuraCloudAPIRoot
	t.Cleanup(func() {
		api.SakuraCloudAPIRoot = backup
	})
	api.SakuraCloudAPIRoot = server.URL
}

func TestDNSProvider_addTXTRecord(t *testing.T) {
	searchResp := &api.SearchDNSResponse{}

	handler := func(rw http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodGet:
			if len(searchResp.CommonServiceDNSItems) == 0 {
				q := &apiQuery{}
				if err := json.Unmarshal([]byte(req.URL.RawQuery), q); err != nil {
					http.Error(rw, err.Error(), http.StatusServiceUnavailable)
				}

				fakeZone := sacloud.CreateNewDNS(q.Filter.Name)
				fakeZone.ID = 123456789012
				searchResp = &api.SearchDNSResponse{CommonServiceDNSItems: []sacloud.DNS{*fakeZone}}
			}

			if err := json.NewEncoder(rw).Encode(searchResp); err != nil {
				http.Error(rw, err.Error(), http.StatusServiceUnavailable)
			}
		case http.MethodPut: // Update
			resp := &simpleResponse{}
			if err := json.NewDecoder(req.Body).Decode(resp); err != nil {
				http.Error(rw, err.Error(), http.StatusServiceUnavailable)
			}

			var items []sacloud.DNS
			for _, v := range searchResp.CommonServiceDNSItems {
				if resp.Name == v.Name {
					items = append(items, *resp.DNS)
				} else {
					items = append(items, v)
				}
			}
			searchResp.CommonServiceDNSItems = items

			if err := json.NewEncoder(rw).Encode(searchResp); err != nil {
				http.Error(rw, err.Error(), http.StatusServiceUnavailable)
			}
		default:
			http.Error(rw, "OOPS", http.StatusServiceUnavailable)
		}
	}

	setupTest(t, handler)

	config := NewDefaultConfig()
	config.Token = "token1"
	config.Secret = "secret1"

	p, err := NewDNSProviderConfig(config)
	require.NoError(t, err)

	err = p.addTXTRecord("test.example.com", "example.com", "dummyValue", 10)
	require.NoError(t, err)

	updZone, err := p.getHostedZone("example.com")
	require.NoError(t, err)
	require.NotNil(t, updZone)

	require.Len(t, updZone.Settings.DNS.ResourceRecordSets, 1)
}

func TestDNSProvider_cleanupTXTRecord(t *testing.T) {
	searchResp := &api.SearchDNSResponse{}

	handler := func(rw http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodGet:
			if len(searchResp.CommonServiceDNSItems) == 0 {
				q := &apiQuery{}
				if err := json.Unmarshal([]byte(req.URL.RawQuery), q); err != nil {
					http.Error(rw, err.Error(), http.StatusServiceUnavailable)
				}

				fakeZone := sacloud.CreateNewDNS(q.Filter.Name)
				fakeZone.ID = 123456789012
				fakeZone.CreateNewRecord("test", "TXT", "dummyValue", 10)
				searchResp = &api.SearchDNSResponse{CommonServiceDNSItems: []sacloud.DNS{*fakeZone}}
			}

			if err := json.NewEncoder(rw).Encode(searchResp); err != nil {
				http.Error(rw, err.Error(), http.StatusServiceUnavailable)
			}
		case http.MethodPut: // Update
			resp := &simpleResponse{}
			if err := json.NewDecoder(req.Body).Decode(resp); err != nil {
				http.Error(rw, err.Error(), http.StatusServiceUnavailable)
			}

			var items []sacloud.DNS
			for _, v := range searchResp.CommonServiceDNSItems {
				if resp.Name == v.Name {
					items = append(items, *resp.DNS)
				} else {
					items = append(items, v)
				}
			}
			searchResp.CommonServiceDNSItems = items

			if err := json.NewEncoder(rw).Encode(searchResp); err != nil {
				http.Error(rw, err.Error(), http.StatusServiceUnavailable)
			}
		default:
			http.Error(rw, "OOPS", http.StatusServiceUnavailable)
		}
	}

	setupTest(t, handler)

	config := NewDefaultConfig()
	config.Token = "token2"
	config.Secret = "secret2"

	p, err := NewDNSProviderConfig(config)
	require.NoError(t, err)

	err = p.cleanupTXTRecord("test.example.com", "example.com")
	require.NoError(t, err)

	updZone, err := p.getHostedZone("example.com")
	require.NoError(t, err)
	require.NotNil(t, updZone)

	require.Len(t, updZone.Settings.DNS.ResourceRecordSets, 0)
}

func TestDNSProvider_addTXTRecord_concurrent(t *testing.T) {
	searchResp := &api.SearchDNSResponse{}

	handler := func(rw http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodGet:
			if len(searchResp.CommonServiceDNSItems) == 0 {
				q := &apiQuery{}
				if err := json.Unmarshal([]byte(req.URL.RawQuery), q); err != nil {
					http.Error(rw, err.Error(), http.StatusServiceUnavailable)
				}

				fakeZone := sacloud.CreateNewDNS(q.Filter.Name)
				fakeZone.ID = 123456789012
				searchResp = &api.SearchDNSResponse{CommonServiceDNSItems: []sacloud.DNS{*fakeZone}}
			}

			if err := json.NewEncoder(rw).Encode(searchResp); err != nil {
				http.Error(rw, err.Error(), http.StatusServiceUnavailable)
			}
		case http.MethodPut: // Update
			resp := &simpleResponse{}
			if err := json.NewDecoder(req.Body).Decode(resp); err != nil {
				http.Error(rw, err.Error(), http.StatusServiceUnavailable)
			}

			var items []sacloud.DNS
			for _, v := range searchResp.CommonServiceDNSItems {
				if resp.Name == v.Name {
					items = append(items, *resp.DNS)
				} else {
					items = append(items, v)
				}
			}
			searchResp.CommonServiceDNSItems = items

			if err := json.NewEncoder(rw).Encode(searchResp); err != nil {
				http.Error(rw, err.Error(), http.StatusServiceUnavailable)
			}
		default:
			http.Error(rw, "OOPS", http.StatusServiceUnavailable)
		}
	}

	setupTest(t, handler)

	dummyRecordCount := 10

	var providers []*DNSProvider
	for i := 0; i < dummyRecordCount; i++ {
		config := NewDefaultConfig()
		config.Token = "token3"
		config.Secret = "secret3"

		p, err := NewDNSProviderConfig(config)
		require.NoError(t, err)

		providers = append(providers, p)
	}

	var wg sync.WaitGroup
	wg.Add(len(providers))

	for i, p := range providers {
		go func(fqdn string, client *DNSProvider) {
			err := client.addTXTRecord(fqdn, "example.com", "dummyValue", 10)
			require.NoError(t, err)
			wg.Done()
		}(fmt.Sprintf("test%d.example.com", i), p)
	}

	wg.Wait()

	updZone, err := providers[0].getHostedZone("example.com")
	require.NoError(t, err)
	require.NotNil(t, updZone)

	require.Len(t, updZone.Settings.DNS.ResourceRecordSets, dummyRecordCount)
}

func TestDNSProvider_cleanupTXTRecord_concurrent(t *testing.T) {
	dummyRecordCount := 10

	baseFakeZone := sacloud.CreateNewDNS("example.com")
	baseFakeZone.ID = 123456789012
	for i := 0; i < dummyRecordCount; i++ {
		baseFakeZone.AddRecord(baseFakeZone.CreateNewRecord(fmt.Sprintf("test%d", i), "TXT", "dummyValue", 10))
	}

	searchResp := &api.SearchDNSResponse{CommonServiceDNSItems: []sacloud.DNS{*baseFakeZone}}

	handler := func(rw http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodGet:
			if err := json.NewEncoder(rw).Encode(searchResp); err != nil {
				http.Error(rw, err.Error(), http.StatusServiceUnavailable)
			}
		case http.MethodPut: // Update
			resp := &simpleResponse{}
			if err := json.NewDecoder(req.Body).Decode(resp); err != nil {
				http.Error(rw, err.Error(), http.StatusServiceUnavailable)
			}

			var items []sacloud.DNS
			for _, v := range searchResp.CommonServiceDNSItems {
				if resp.Name == v.Name {
					items = append(items, *resp.DNS)
				} else {
					items = append(items, v)
				}
			}
			searchResp.CommonServiceDNSItems = items

			if err := json.NewEncoder(rw).Encode(searchResp); err != nil {
				http.Error(rw, err.Error(), http.StatusServiceUnavailable)
			}
		default:
			http.Error(rw, "OOPS", http.StatusServiceUnavailable)
		}
	}

	setupTest(t, handler)

	fakeZone := sacloud.CreateNewDNS("example.com")
	fakeZone.ID = 123456789012
	for i := 0; i < dummyRecordCount; i++ {
		fakeZone.AddRecord(fakeZone.CreateNewRecord(fmt.Sprintf("test%d", i), "TXT", "dummyValue", 10))
	}

	var providers []*DNSProvider
	for i := 0; i < dummyRecordCount; i++ {
		config := NewDefaultConfig()
		config.Token = "token4"
		config.Secret = "secret4"

		p, err := NewDNSProviderConfig(config)
		require.NoError(t, err)
		providers = append(providers, p)
	}

	var wg sync.WaitGroup
	wg.Add(len(providers))

	for i, p := range providers {
		go func(fqdn string, client *DNSProvider) {
			err := client.cleanupTXTRecord(fqdn, "example.com")
			require.NoError(t, err)
			wg.Done()
		}(fmt.Sprintf("test%d.example.com", i), p)
	}

	wg.Wait()

	updZone, err := providers[0].getHostedZone("example.com")
	require.NoError(t, err)
	require.NotNil(t, updZone)

	require.Len(t, updZone.Settings.DNS.ResourceRecordSets, 0)
}
