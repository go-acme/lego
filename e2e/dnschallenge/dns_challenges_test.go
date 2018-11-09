package dnschallenge

import (
	"fmt"
	"os"
	"testing"

	"github.com/xenolf/lego/e2e/loader"
)

var load = loader.EnvLoader{
	PebbleOptions: &loader.CmdOption{
		Args: []string{"-strict", "-config", "fixtures/pebble-config-dns.json", "-dnsserver", "localhost:8053"},
		Env:  []string{"PEBBLE_VA_NOSLEEP=1", "PEBBLE_WFE_NONCEREJECT=20"},
		Dir:  "../",
	},
	LegoOptions: []string{
		"LEGO_CA_CERTIFICATES=../fixtures/certs/pebble.minica.pem",
		"EXEC_PATH=../fixtures/update-dns.sh",
	},
	ChallSrv: &loader.CmdOption{
		Args: []string{"-http01", ":5012", "-tlsalpn01", ":5011"},
	},
}

func TestMain(m *testing.M) {
	os.Exit(load.MainTest(m))
}

func TestDNSHelp(t *testing.T) {
	output, err := load.RunLego("dnshelp")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", output)
		t.Fatal(err)
	}

	fmt.Fprintf(os.Stdout, "%s\n", output)
}

func TestChallengeDNS(t *testing.T) {
	loader.CleanLegoFiles()

	output, err := load.RunLego(
		"-m", "hubert@hubert.com",
		"-a",
		"-x", "http-01",
		"-x", "tls-alpn-01",
		"--disable-cp",
		"--dns-resolvers", ":8053",
		"--dns", "exec",
		"-s", "https://localhost:15000/dir",
		"-d", "*.lego.acme",
		"-d", "lego.acme",
		"--http", ":5004",
		"--tls", ":5003",
		"run")

	if len(output) > 0 {
		fmt.Fprintf(os.Stdout, "%s\n", output)
	}
	if err != nil {
		t.Fatal(err)
	}
}
