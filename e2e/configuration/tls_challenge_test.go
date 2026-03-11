package configuration

import (
	"path/filepath"
	"testing"

	"github.com/go-acme/lego/v5/e2e/loader"
)

func TestChallengeTLS_Run_simple(t *testing.T) {
	loader.CleanLegoFiles(t.Context())

	err := load.RunLego(t.Context(),
		"--config", filepath.Join("fixtures", "lego_tls-simple.yml"),
		"--log.level", "debug",
	)
	if err != nil {
		t.Fatal(err)
	}
}
