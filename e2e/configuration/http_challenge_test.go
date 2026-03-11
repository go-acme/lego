package configuration

import (
	"path/filepath"
	"testing"

	"github.com/go-acme/lego/v5/e2e/loader"
)

func TestChallengeHTTP_Run_simple(t *testing.T) {
	loader.CleanLegoFiles(t.Context())

	err := load.RunLego(t.Context(),
		"--config", filepath.Join("fixtures", "lego_http-simple.yml"),
		"--log.level", "debug",
	)
	if err != nil {
		t.Fatal(err)
	}
}

func TestChallengeHTTP_Run_file_server(t *testing.T) {
	loader.CleanLegoFiles(t.Context())

	err := load.RunLego(t.Context(),
		"--config", filepath.Join("fixtures", "lego_http-server.yml"),
		"--log.level", "debug",
	)
	if err != nil {
		t.Fatal(err)
	}
}
