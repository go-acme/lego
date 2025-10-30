package liquidweb

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-acme/lego/v4/platform/tester/servermock"
	"github.com/liquidweb/liquidweb-go/network"
	"github.com/liquidweb/liquidweb-go/types"
)

func mockProvider(t *testing.T, initRecs ...network.DNSRecord) *DNSProvider {
	t.Helper()

	recs := make(map[int]network.DNSRecord)

	for _, rec := range initRecs {
		recs[int(rec.ID)] = rec
	}

	return servermock.NewBuilder(
		func(server *httptest.Server) (*DNSProvider, error) {
			config := NewDefaultConfig()
			config.Username = "blars"
			config.Password = "tacoman"
			config.BaseURL = server.URL

			return NewDNSProviderConfig(config)
		},
		servermock.CheckHeader().
			WithBasicAuth("blars", "tacoman"),
	).
		Route("/v1/Network/DNS/Record/delete", mockAPIDelete(recs)).
		Route("/v1/Network/DNS/Record/create", mockAPICreate(recs)).
		Route("/v1/Network/DNS/Zone/list", mockAPIListZones()).
		Route("/bleed/Network/DNS/Record/delete", mockAPIDelete(recs)).
		Route("/bleed/Network/DNS/Record/create", mockAPICreate(recs)).
		Route("/bleed/Network/DNS/Zone/list", mockAPIListZones()).
		Build(t)
}

func mockAPICreate(recs map[int]network.DNSRecord) http.HandlerFunc {
	_, mockAPIServerZones := makeMockZones()

	return func(rw http.ResponseWriter, req *http.Request) {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(rw, "invalid request", http.StatusInternalServerError)
			return
		}

		payload := struct {
			Params network.DNSRecord `json:"params"`
		}{}

		if err = json.Unmarshal(body, &payload); err != nil {
			http.Error(rw, makeEncodingError(body), http.StatusBadRequest)
			return
		}

		payload.Params.ID = types.FlexInt(rand.Intn(10000000))
		payload.Params.ZoneID = types.FlexInt(mockAPIServerZones[payload.Params.Name])

		if _, exists := recs[int(payload.Params.ID)]; exists {
			http.Error(rw, "dns record already exists", http.StatusTeapot)
			return
		}

		recs[int(payload.Params.ID)] = payload.Params

		resp, err := json.Marshal(payload.Params)
		if err != nil {
			http.Error(rw, "", http.StatusInternalServerError)
			return
		}

		http.Error(rw, string(resp), http.StatusOK)
	}
}

func mockAPIDelete(recs map[int]network.DNSRecord) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(rw, "invalid request", http.StatusInternalServerError)
			return
		}

		payload := struct {
			Params struct {
				Name string `json:"name"`
				ID   int    `json:"id"`
			} `json:"params"`
		}{}

		if err := json.Unmarshal(body, &payload); err != nil {
			http.Error(rw, makeEncodingError(body), http.StatusBadRequest)
			return
		}

		if payload.Params.ID == 0 {
			http.Error(rw, `{"error":"","error_class":"LW::Exception::Input::Multiple","errors":[{"error":"","error_class":"LW::Exception::Input::Required","field":"id","full_message":"The required field 'id' was missing a value.","position":null}],"field":["id"],"full_message":"The following input errors occurred:\nThe required field 'id' was missing a value.","type":null}`, http.StatusOK)
			return
		}

		if _, ok := recs[payload.Params.ID]; !ok {
			http.Error(rw, fmt.Sprintf(`{"error":"","error_class":"LW::Exception::RecordNotFound","field":"network_dns_rr","full_message":"Record 'network_dns_rr: %d' not found","input":"%d","public_message":null}`, payload.Params.ID, payload.Params.ID), http.StatusOK)
			return
		}

		delete(recs, payload.Params.ID)
		http.Error(rw, fmt.Sprintf("{\"deleted\":%d}", payload.Params.ID), http.StatusOK)
	}
}

func mockAPIListZones() http.HandlerFunc {
	mockZones, mockAPIServerZones := makeMockZones()

	return func(rw http.ResponseWriter, req *http.Request) {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(rw, "invalid request", http.StatusInternalServerError)
			return
		}

		payload := struct {
			Params struct {
				PageNum int `json:"page_num"`
			} `json:"params"`
		}{}

		if err = json.Unmarshal(body, &payload); err != nil {
			http.Error(rw, makeEncodingError(body), http.StatusBadRequest)
			return
		}

		switch {
		case payload.Params.PageNum < 1:
			payload.Params.PageNum = 1
		case payload.Params.PageNum > len(mockZones):
			payload.Params.PageNum = len(mockZones)
		}

		resp := mockZones[payload.Params.PageNum]
		resp.ItemTotal = types.FlexInt(len(mockAPIServerZones))
		resp.PageNum = types.FlexInt(payload.Params.PageNum)
		resp.PageSize = 5
		resp.PageTotal = types.FlexInt(len(mockZones))

		var respBody []byte
		if respBody, err = json.Marshal(resp); err == nil {
			http.Error(rw, string(respBody), http.StatusOK)
			return
		}

		http.Error(rw, "", http.StatusInternalServerError)
	}
}

func makeEncodingError(buf []byte) string {
	return fmt.Sprintf(`{"data":"%q","encoding":"JSON","error":"unexpected end of string while parsing JSON string, at character offset 32 (before \"(end of string)\") at /usr/local/lp/libs/perl/LW/Base/Role/Serializer.pm line 16.\n","error_class":"LW::Exception::Deserialize","full_message":"Could not deserialize \"%q\" from JSON: unexpected end of string while parsing JSON string, at character offset 32 (before \"(end of string)\") at /usr/local/lp/libs/perl/LW/Base/Role/Serializer.pm line 16.\n"}⏎`, string(buf), string(buf))
}

func makeMockZones() (map[int]network.DNSZoneList, map[string]int) {
	mockZones := map[int]network.DNSZoneList{
		1: {
			Items: []network.DNSZone{
				{
					ID:                1,
					Name:              "blars.com",
					Active:            1,
					DelegationStatus:  "CORRECT",
					PrimaryNameserver: "ns.liquidweb.com",
				},
				{
					ID:                2,
					Name:              "tacoman.com",
					Active:            1,
					DelegationStatus:  "CORRECT",
					PrimaryNameserver: "ns.liquidweb.com",
				},
				{
					ID:                3,
					Name:              "storm.com",
					Active:            1,
					DelegationStatus:  "CORRECT",
					PrimaryNameserver: "ns.liquidweb.com",
				},
				{
					ID:                4,
					Name:              "not-apple.com",
					Active:            1,
					DelegationStatus:  "BAD_NAMESERVERS",
					PrimaryNameserver: "ns.liquidweb.com",
				},
				{
					ID:                5,
					Name:              "example.com",
					Active:            1,
					DelegationStatus:  "BAD_NAMESERVERS",
					PrimaryNameserver: "ns.liquidweb.com",
				},
			},
		},
		2: {
			Items: []network.DNSZone{
				{
					ID:                6,
					Name:              "banana.com",
					Active:            1,
					DelegationStatus:  "NXDOMAIN",
					PrimaryNameserver: "ns.liquidweb.com",
				},
				{
					ID:                7,
					Name:              "cherry.com",
					Active:            1,
					DelegationStatus:  "SERVFAIL",
					PrimaryNameserver: "ns.liquidweb.com",
				},
				{
					ID:                8,
					Name:              "dates.com",
					Active:            1,
					DelegationStatus:  "SERVFAIL",
					PrimaryNameserver: "ns.liquidweb.com",
				},
				{
					ID:                9,
					Name:              "eggplant.com",
					Active:            1,
					DelegationStatus:  "SERVFAIL",
					PrimaryNameserver: "ns.liquidweb.com",
				},
				{
					ID:                10,
					Name:              "fig.com",
					Active:            1,
					DelegationStatus:  "UNKNOWN",
					PrimaryNameserver: "ns.liquidweb.com",
				},
			},
		},
		3: {
			Items: []network.DNSZone{
				{
					ID:                11,
					Name:              "grapes.com",
					Active:            1,
					DelegationStatus:  "UNKNOWN",
					PrimaryNameserver: "ns.liquidweb.com",
				},
				{
					ID:                12,
					Name:              "money.banana.com",
					Active:            1,
					DelegationStatus:  "UNKNOWN",
					PrimaryNameserver: "ns.liquidweb.com",
				},
				{
					ID:                13,
					Name:              "money.stand.banana.com",
					Active:            1,
					DelegationStatus:  "UNKNOWN",
					PrimaryNameserver: "ns.liquidweb.com",
				},
				{
					ID:                14,
					Name:              "stand.banana.com",
					Active:            1,
					DelegationStatus:  "UNKNOWN",
					PrimaryNameserver: "ns.liquidweb.com",
				},
			},
		},
	}

	mockAPIServerZones := make(map[string]int)

	for _, page := range mockZones {
		for _, zone := range page.Items {
			mockAPIServerZones[zone.Name] = int(zone.ID)
		}
	}

	return mockZones, mockAPIServerZones
}
