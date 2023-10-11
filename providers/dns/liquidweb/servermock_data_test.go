package liquidweb

import (
	"fmt"

	"github.com/liquidweb/liquidweb-go/network"
)

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
