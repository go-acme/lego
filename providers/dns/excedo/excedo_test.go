package excedo

import (
	"encoding/json"
	"testing"
)

func TestNewDNSProviderConfigValidation(t *testing.T) {
	if _, err := NewDNSProviderConfig(nil); err == nil {
		t.Fatal("expected error for nil config")
	}

	config := NewDefaultConfig()
	if _, err := NewDNSProviderConfig(config); err == nil {
		t.Fatal("expected error for missing API key")
	}

	config = NewDefaultConfig()
	config.APIKey = "token"
	config.APIURL = ""
	if _, err := NewDNSProviderConfig(config); err == nil {
		t.Fatal("expected error for missing API URL")
	}

	config = NewDefaultConfig()
	config.APIKey = "token"
	config.HTTPClient = nil
	if _, err := NewDNSProviderConfig(config); err == nil {
		t.Fatal("expected error for missing HTTP client")
	}
}

func TestNewDNSProviderFromEnv(t *testing.T) {
	t.Setenv(envAPIKey, "token")
	t.Setenv(envAPIURL, "https://example.com")

	if _, err := NewDNSProvider(); err != nil {
		t.Fatalf("expected provider, got error: %v", err)
	}
}

func TestAPIResponseDNSArray(t *testing.T) {
	raw := []byte(`{"code":1000,"desc":"ok","dns":[]}`)

	var payload apiResponse
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("expected array dns to decode, got error: %v", err)
	}

	if len(payload.DNS) != 0 {
		t.Fatalf("expected empty dns map, got %d entries", len(payload.DNS))
	}
}

func TestAPIResponseDNSMixedEntries(t *testing.T) {
	raw := []byte(`{
		"code":1000,
		"desc":"ok",
		"dns":{
			"example.com":{"records":[{"recordid":"1","name":"_acme-challenge.example.com","type":"TXT","content":"token"}]},
			"available-types":[{"type":"A"}],
			"available-ttl":[{"ttl":60}]
		}
	}`)

	var payload apiResponse
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("expected mixed dns entries to decode, got error: %v", err)
	}

	var zone dnsZone
	if err := json.Unmarshal(payload.DNS["example.com"], &zone); err != nil {
		t.Fatalf("expected zone entry to decode, got error: %v", err)
	}

	if len(zone.Records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(zone.Records))
	}
}

func TestRecordNameMatches(t *testing.T) {
	zone := "excedodnssec.se"
	challenge := "_acme-challenge"

	if !recordNameMatches(zone, challenge, "_acme-challenge") {
		t.Fatal("expected relative challenge name to match")
	}

	if !recordNameMatches(zone, challenge, "_acme-challenge.excedodnssec.se") {
		t.Fatal("expected fqdn challenge name to match")
	}

	if recordNameMatches(zone, challenge, "_acme-challenge.other.example") {
		t.Fatal("expected unrelated name not to match")
	}
}
