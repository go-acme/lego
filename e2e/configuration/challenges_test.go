package configuration

import (
	"context"
	"os"
	"testing"

	"github.com/go-acme/lego/v5/e2e/loader"
)

const caDirectory = "https://localhost:17000/dir"

var load = loader.EnvLoader{
	PebbleOptions: &loader.CmdOption{
		HealthCheckURL: caDirectory,
		Args:           []string{"-strict", "-config", "fixtures/pebble-config-file.json", "-dnsserver", "localhost:8853"},
		Env:            []string{"PEBBLE_VA_NOSLEEP=1", "PEBBLE_WFE_NONCEREJECT=20"},
		Dir:            "../",
	},
	LegoOptions: []string{
		"LEGO_CA_CERTIFICATES=../fixtures/certs/pebble.minica.pem",
		"LEGO_DEBUG_ACME_HTTP_CLIENT=1",
	},
	ChallSrv: &loader.CmdOption{
		Args: []string{"-dnsserver", ":8853", "-http01", ":5019", "-tlsalpn01", ":5018", "-management", ":8855"},
	},
}

func TestMain(m *testing.M) {
	os.Exit(load.MainTest(context.Background(), m))
}
