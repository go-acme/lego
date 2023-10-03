package liquidweb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/liquidweb/liquidweb-go/network"
	"github.com/liquidweb/liquidweb-go/types"
)

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

func mockApiCreate(recs map[int]network.DNSRecord) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "invalid request", http.StatusInternalServerError)
			return
		}

		req := struct {
			Params network.DNSRecord `json:"params"`
		}{}

		if err := json.Unmarshal(body, &req); err != nil {
			resp := jsonEncodingError
			resp.Data = string(body)
			resp.FullMessage = fmt.Sprintf(resp.FullMessage, string(body))
			json.NewEncoder(w).Encode(resp)
		}
		req.Params.ZoneID = 1

		if _, exists := recs[int(req.Params.ID)]; exists {
			http.Error(w, "dns record already exists", http.StatusTeapot)
			return
		}

		recs[int(req.Params.ID)] = req.Params
		recsNameToID[req.Params.Name] = int(req.Params.ID)

		resp, err := json.Marshal(req.Params)
		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		w.Write(resp)
		return
	}
}

func mockApiDelete(recs map[int]network.DNSRecord) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "invalid request", http.StatusInternalServerError)
			return
		}

		req := struct {
			Params struct {
				Name string `json:"name"`
				ID   int    `json:"id"`
			} `json:"params"`
		}{}

		if err := json.Unmarshal(body, &req); err != nil {
			resp := jsonEncodingError
			resp.Data = string(body)
			resp.FullMessage = fmt.Sprintf(resp.FullMessage, string(body))
			json.NewEncoder(w).Encode(resp)
		}

		if req.Params.ID == 0 && req.Params.Name == "" {
			http.Error(w, `{"error":"","error_class":"LW::Exception::Input::Multiple","errors":[{"error":"","error_class":"LW::Exception::Input::Required","field":"id","full_message":"The required field 'id' was missing a value.","position":null}],"field":["id"],"full_message":"The following input errors occurred:\nThe required field 'id' was missing a value.","type":null}`, http.StatusOK)
			return
		}

		if req.Params.ID == 0 {
			if name, ok := recsNameToID[req.Params.Name]; ok {
				req.Params.ID = name
			}
		}

		if _, ok := recs[req.Params.ID]; ok {
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
		resp.ItemTotal = types.FlexInt(len(mockApiServerZones))
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

var recsNameToID map[string]int

func mockApiServer(t *testing.T, initRecs ...network.DNSRecord) string {
	t.Helper()

	recs := make(map[int]network.DNSRecord)
	recsNameToID = make(map[string]int)

	for _, rec := range initRecs {
		recs[int(rec.ID)] = rec
		recsNameToID[rec.Name] = int(rec.ID)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/v1/Network/DNS/Record/delete", mockApiDelete(recs))
	mux.HandleFunc("/v1/Network/DNS/Record/create", mockApiCreate(recs))
	mux.HandleFunc("/v1/Network/DNS/Zone/list", mockApiListZones())
	handler := http.HandlerFunc(requireJson(mux))
	handler = http.HandlerFunc(requireBasicAuth(handler))

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return server.URL
}
