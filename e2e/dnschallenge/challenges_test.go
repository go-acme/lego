package dnschallenge

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/go-acme/lego/v5/e2e/loader"
)

const caDirectory = "https://localhost:15000/dir"

var load = loader.EnvLoader{
	PebbleOptions: &loader.CmdOption{
		HealthCheckURL: caDirectory,
		Args:           []string{"-strict", "-config", "fixtures/pebble-config-dns.json", "-dnsserver", "localhost:8553"},
		Env:            []string{"PEBBLE_VA_NOSLEEP=1", "PEBBLE_WFE_NONCEREJECT=20"},
		Dir:            "../",
	},
	LegoOptions: []string{
		"LEGO_CA_CERTIFICATES=../fixtures/certs/pebble.minica.pem",
		"LEGO_DEBUG_ACME_HTTP_CLIENT=1",
	},
	ChallSrv: &loader.CmdOption{
		Args: []string{"-dnsserver", ":8553", "-http01", ":5012", "-tlsalpn01", ":5011", "-management", ":8555"},
	},
}

func TestMain(m *testing.M) {
	os.Exit(load.MainTest(context.Background(), m))
}

func TestDNSHelp(t *testing.T) {
	output, err := load.RunLegoCombinedOutput(t.Context(), "dnshelp")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", output)
		t.Fatal(err)
	}

	fmt.Fprintf(os.Stdout, "%s\n", output)
}
