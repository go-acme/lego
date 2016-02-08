package dns_provider

import "testing"

func TestProviderResistry(t *testing.T) {
	entries := Registry.Entries()
	if len(entries) == 0 {
		t.Fatal("expected to have entries, had 0")
	}
}
