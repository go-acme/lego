package e2e

import (
	"fmt"
	"os"
	"testing"

	"github.com/xenolf/lego/e2e/loader"
)

var load = loader.EnvLoader{
	PebbleOptions: &loader.CmdOption{
		Args: []string{"-strict", "-config", "fixtures/pebble-config.json"},
		Env:  []string{"PEBBLE_VA_NOSLEEP=1", "PEBBLE_WFE_NONCEREJECT=20"},
	},
	LegoOptions: []string{
		"LEGO_CA_CERTIFICATES=./fixtures/certs/pebble.minica.pem",
	},
}

func init() {
	os.Setenv("LEGO_E2E_TESTS", "LEGO_E2E_TESTS")
}

func TestMain(m *testing.M) {
	os.Exit(load.MainTest(m))
}

func TestHelp(t *testing.T) {
	output, err := load.RunLego("-h")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", output)
		t.Fatal(err)
	}

	fmt.Fprintf(os.Stdout, "%s\n", output)
}

func TestChallengeHTTP(t *testing.T) {
	loader.CleanLegoFiles()

	output, err := load.RunLego(
		"-m", "hubert@hubert.com",
		"-a",
		"-x", "dns-01",
		"-x", "tls-alpn-01",
		"-s", "https://localhost:14000/dir",
		"-d", "acme.wtf",
		"--http", ":5002",
		"--tls", ":5001",
		"run")

	if len(output) > 0 {
		fmt.Fprintf(os.Stdout, "%s\n", output)
	}
	if err != nil {
		t.Fatal(err)
	}
}

func TestChallengeTLS(t *testing.T) {
	loader.CleanLegoFiles()

	output, err := load.RunLego(
		"-m", "hubert@hubert.com",
		"-a",
		"-x", "dns-01",
		"-x", "http-01",
		"-s", "https://localhost:14000/dir",
		"-d", "acme.wtf",
		"--http", ":5002",
		"--tls", ":5001",
		"run")

	if len(output) > 0 {
		fmt.Fprintf(os.Stdout, "%s\n", output)
	}
	if err != nil {
		t.Fatal(err)
	}
}
