package configuration

import (
	"path/filepath"
	"testing"

	"github.com/go-acme/lego/v5/e2e/loader"
)

func TestChallengeDNS_Run_simple(t *testing.T) {
	loader.CleanLegoFiles(t.Context())

	err := load.RunLego(t.Context(),
		"--config", filepath.Join("fixtures", "lego_dns-simple.yml"),
		"--log.level", "debug",
	)
	if err != nil {
		t.Fatal(err)
	}
}

func TestChallengeDNS_Run_explicit_challenge(t *testing.T) {
	loader.CleanLegoFiles(t.Context())

	err := load.RunLego(t.Context(),
		"--config", filepath.Join("fixtures", "lego_dns-explicit.yml"),
		"--log.level", "debug",
	)
	if err != nil {
		t.Fatal(err)
	}
}
