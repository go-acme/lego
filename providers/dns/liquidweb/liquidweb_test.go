package liquidweb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/liquidweb/liquidweb-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const envDomain = EnvPrefix + "DOMAIN"

var envTest = tester.NewEnvTest(
	EnvPrefix+EnvURL,
	EnvPrefix+EnvUsername,
	EnvPrefix+EnvPassword,
	EnvPrefix+EnvZone).
	WithDomain(envDomain)

func requireBasicAuth(child http.Handler) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if ok && username == "blars" && password == "tacoman" {
			child.ServeHTTP(w, r)
			return
		}
		http.Error(w, "invalid auth", http.StatusForbidden)
	}
}

func requireJson(child http.Handler) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		buf := &bytes.Buffer{}
		_, err := buf.ReadFrom(r.Body)
		if err != nil {
			http.Error(w, "malformed request - json required", http.StatusBadRequest)
			return
		}
		r.Body = io.NopCloser(buf)
		child.ServeHTTP(w, r)
	}
}

func mockApiCreate(recs map[int]string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "invalid request", http.StatusInternalServerError)
			return
		}

		req := struct {
			Params struct {
				Name   string `json:"name"`
				Rdata  string `json:"rdata"`
				Type   string `json:"type"`
				ID     int    `json:"ID"`
				Zone   int    `json:"zone"`
				ZoneID int    `json:"zone_id"`
			} `json:"params"`
		}{}
		req.Params.ZoneID = 1

		if err := json.Unmarshal(body, &req); err != nil {
			resp := jsonEncodingError
			resp.Data = string(body)
			resp.FullMessage = fmt.Sprintf(resp.FullMessage, string(body))
			json.NewEncoder(w).Encode(resp)
		}

		if val, ok := mockApiServerRecords[req.Params.Name]; ok {
			recs[val] = req.Params.Name
			req.Params.ID = val
			resp, err := json.Marshal(req.Params)
			if err != nil {
				http.Error(w, "", http.StatusInternalServerError)
			}
			w.Write(resp)
			return
		}
		http.Error(w, "record not defined for tests", http.StatusInternalServerError)
		return
	}
}

func mockApiDelete(recs map[int]string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "invalid request", http.StatusInternalServerError)
			return
		}

		req := struct {
			Params struct {
				ID int `json:"id"`
			} `json:"params"`
		}{}

		if err := json.Unmarshal(body, &req); err != nil {
			resp := jsonEncodingError
			resp.Data = string(body)
			resp.FullMessage = fmt.Sprintf(resp.FullMessage, string(body))
			json.NewEncoder(w).Encode(resp)
		}

		if val, ok := recs[req.Params.ID]; ok && val != "" {
			delete(recs, req.Params.ID)
			w.Write([]byte(fmt.Sprintf("{\"deleted\":%d}", req.Params.ID)))
			return
		}
		http.Error(w, fmt.Sprintf(`{"error":"","error_class":"LW::Exception::RecordNotFound","field":"network_dns_rr","full_message":"Record 'network_dns_rr: %d' not found","input":"%d","public_message":null}`, req.Params.ID, req.Params.ID), http.StatusOK)
		return
	}
}

func mockApiListZones() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "invalid request", http.StatusInternalServerError)
			return
		}

		req := struct {
			Params struct {
				PageNum int `json:"page_num"`
			} `json:"params"`
		}{}

		if err := json.Unmarshal(body, &req); err != nil {
			resp := jsonEncodingError
			resp.Data = string(body)
			resp.FullMessage = fmt.Sprintf(resp.FullMessage, string(body))
			json.NewEncoder(w).Encode(resp)
		}

		switch {
		case req.Params.PageNum < 1:
			req.Params.PageNum = 1
		case req.Params.PageNum > len(mockZones):
			req.Params.PageNum = len(mockZones)
		}
		resp := mockZones[req.Params.PageNum]
		resp.ItemTotal = types.FlexInt(len(mockApiServerRecords))
		resp.PageNum = types.FlexInt(req.Params.PageNum)
		resp.PageSize = 5
		resp.PageTotal = types.FlexInt(len(mockZones))

		if respBody, err := json.Marshal(resp); err == nil {
			w.Write(respBody)
			return
		}
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
}

func testApiServer(t *testing.T) {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	recs := map[int]string{}
	mux.HandleFunc("/Network/DNS/Record/delete", mockApiDelete(t, recs))
	mux.HandleFunc("/Network/DNS/Record/create", mockApiCreate(t, recs))
	mux.HandleFunc("/Network/DNS/Zone/list", mockApiListZones(t))

}

func setupTest(t *testing.T) (*DNSProvider, *http.ServeMux) {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	envTest.Apply(map[string]string{
		EnvPrefix + EnvUsername: "blars",
		EnvPrefix + EnvPassword: "tacoman",
		EnvPrefix + EnvURL:      server.URL,
		EnvPrefix + EnvZone:     "tacoman.com", // this needs to be removed from test?
	})

	provider, err := NewDNSProvider()
	require.NoError(t, err)

	return provider, mux
}

func TestNewDNSProvider(t *testing.T) {
	for _, test := range testNewDNSProvider_testdata {
		t.Run(test.desc, func(t *testing.T) {
			defer envTest.RestoreEnv()
			envTest.ClearEnv()

			envTest.Apply(test.envVars)

			p, err := NewDNSProvider()

			if test.expected == "" {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.config)
				require.NotNil(t, p.client)
				require.NotNil(t, p.recordIDs)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

/*
func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc     string
		username string
		password string
		zone     string
		expected string
	}{
		{
			desc:     "success",
			username: "acme",
			password: "secret",
			zone:     "example.com",
		},
		{
			desc:     "missing credentials",
			username: "",
			password: "",
			zone:     "",
			expected: "liquidweb: zone is missing",
		},
		{
			desc:     "missing username",
			username: "",
			password: "secret",
			zone:     "example.com",
			expected: "liquidweb: username is missing",
		},
		{
			desc:     "missing password",
			username: "acme",
			password: "",
			zone:     "example.com",
			expected: "liquidweb: password is missing",
		},
		{
			desc:     "missing zone",
			username: "acme",
			password: "secret",
			zone:     "",
			expected: "liquidweb: zone is missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.Username = test.username
			config.Password = test.password
			config.Zone = test.zone

			p, err := NewDNSProviderConfig(config)

			if test.expected == "" {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.config)
				require.NotNil(t, p.client)
				require.NotNil(t, p.recordIDs)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}
*/

func TestDNSProvider_Present(t *testing.T) {
	provider, mux := setupTest(t)

	mux.HandleFunc("/v1/Network/DNS/Record/create", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)

		username, password, ok := r.BasicAuth()
		assert.Equal(t, "blars", username)
		assert.Equal(t, "tacoman", password)
		assert.True(t, ok)

		reqBody, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		expectedReqBody := `
			{
				"params": {
					"name": "_acme-challenge.tacoman.com",
					"rdata": "\"47DEQpj8HBSa-_TImW-5JCeuQeRkm5NMpJWZG3hSuFU\"",
					"ttl": 300,
					"type": "TXT",
					"zone": "tacoman.com"
				}
			}`
		assert.JSONEq(t, expectedReqBody, string(reqBody))

		w.WriteHeader(http.StatusOK)
		_, err = fmt.Fprintf(w, `{
			"type": "TXT",
			"name": "_acme-challenge.tacoman.com",
			"rdata": "\"47DEQpj8HBSa-_TImW-5JCeuQeRkm5NMpJWZG3hSuFU\"",
			"ttl": 300,
			"id": 1234567,
			"prio": null
		}`)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	err := provider.Present("tacoman.com", "", "")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider, mux := setupTest(t)

	mux.HandleFunc("/v1/Network/DNS/Record/delete", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)

		username, password, ok := r.BasicAuth()
		assert.Equal(t, "blars", username)
		assert.Equal(t, "tacoman", password)
		assert.True(t, ok)

		_, err := fmt.Fprintf(w, `{"deleted": "123"}`)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	provider.recordIDs["123"] = 1234567

	err := provider.CleanUp("tacoman.com.", "123", "")
	require.NoError(t, err, "fail to remove TXT record")
}

func TestLivePresent(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()

	provider, err := NewDNSProvider()
	require.NoError(t, err)

	err = provider.Present(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}

func TestLiveCleanUp(t *testing.T) {
	if !envTest.IsLiveTest() {
		t.Skip("skipping live test")
	}

	envTest.RestoreEnv()

	provider, err := NewDNSProvider()
	require.NoError(t, err)

	time.Sleep(2 * time.Second)

	err = provider.CleanUp(envTest.GetDomain(), "", "123d==")
	require.NoError(t, err)
}
