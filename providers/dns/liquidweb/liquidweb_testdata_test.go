package liquidweb

import (
	"github.com/liquidweb/liquidweb-go/network"
)

var mockApiServerZones = make(map[string]int)

func init() {
	for _, page := range mockZones {
		for _, zone := range page.Items {
			mockApiServerZones[zone.Name] = int(zone.ID)
		}
	}
}

var jsonEncodingError = struct {
	Data        string `json:"data"`
	Encoding    string `json:"encoding"`
	Err         string `json:"error"`
	ErrClass    string `json:"error_class"`
	FullMessage string `json:"full_message"`
}{
	Encoding:    "JSON",
	ErrClass:    "LW::Exception::Deserialize",
	Err:         `unexpected end of string while parsing JSON string, at character offset 32 (before \"(end of string)\") at /usr/local/lp/libs/perl/LW/Base/Role/Serializer.pm line 16.\n`,
	FullMessage: `Could not deserialize "%s" from JSON: unexpected end of string while parsing JSON string, at character offset 32 (before "(end of string)") at /usr/local/lp/libs/perl/LW/Base/Role/Serializer.pm line 16.\n`,
}

var mockZones = map[int]network.DNSZoneList{
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

var testNewDNSProvider_testdata = []struct {
	desc     string
	envVars  map[string]string
	expected string
}{
	{
		desc: "minimum-success",
		envVars: map[string]string{
			EnvPrefix + EnvUsername: "blars",
			EnvPrefix + EnvPassword: "tacoman",
		},
	},
	{
		desc: "set-everything",
		envVars: map[string]string{
			EnvPrefix + EnvURL:      "https://storm.com",
			EnvPrefix + EnvUsername: "blars",
			EnvPrefix + EnvPassword: "tacoman",
			EnvPrefix + EnvZone:     "blars.com",
		},
	},
	{
		desc:     "missing credentials",
		envVars:  map[string]string{},
		expected: "liquidweb: username and password are missing, set LWAPI_USERNAME and LWAPI_PASSWORD",
	},
	{
		desc: "missing username",
		envVars: map[string]string{
			EnvPrefix + EnvPassword: "tacoman",
			EnvPrefix + EnvZone:     "blars.com",
		},
		expected: "liquidweb: username is missing, set LWAPI_USERNAME",
	},
	{
		desc: "missing password",
		envVars: map[string]string{
			EnvPrefix + EnvUsername: "blars",
			EnvPrefix + EnvZone:     "blars.com",
		},
		expected: "liquidweb: password is missing, set LWAPI_PASSWORD",
	},
}

var testIntegration_testdata = map[string]struct {
	envVars       map[string]string
	initRecs      []network.DNSRecord
	domain        string
	token         string
	keyauth       string
	present       bool
	cleanup       bool
	expPresentErr string
	expCleanupErr string
}{
	"expected successful": {
		envVars: map[string]string{
			"LWAPI_USERNAME": "blars",
			"LWAPI_PASSWORD": "tacoman",
		},
		domain:  "tacoman.com",
		token:   "123",
		keyauth: "456",
		present: true,
		cleanup: true,
	},
	"other successful": {
		envVars: map[string]string{
			"LWAPI_USERNAME": "blars",
			"LWAPI_PASSWORD": "tacoman",
		},
		domain:  "banana.com",
		token:   "123",
		keyauth: "456",
		present: true,
		cleanup: true,
	},
	"zone not on account": {
		envVars: map[string]string{
			"LWAPI_USERNAME": "blars",
			"LWAPI_PASSWORD": "tacoman",
		},
		domain:        "huckleberry.com",
		token:         "123",
		keyauth:       "456",
		present:       true,
		cleanup:       false,
		expPresentErr: "no valid zone in account for certificate _acme-challenge.huckleberry.com",
	},
	"ssl for domain": {
		envVars: map[string]string{
			"LWAPI_USERNAME": "blars",
			"LWAPI_PASSWORD": "tacoman",
		},
		domain:        "sundae.cherry.com",
		token:         "5847953",
		keyauth:       "34872934",
		present:       true,
		cleanup:       true,
		expPresentErr: "",
	},
	"complicated domain": {
		envVars: map[string]string{
			"LWAPI_USERNAME": "blars",
			"LWAPI_PASSWORD": "tacoman",
		},
		domain:        "always.money.stand.banana.com",
		token:         "5847953",
		keyauth:       "theres always money in the banana stand",
		present:       true,
		cleanup:       true,
		expPresentErr: "",
	},
}
