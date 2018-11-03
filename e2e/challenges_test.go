package e2e

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/xenolf/lego/log"
)

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(testMain(m))
}

func TestHelp(t *testing.T) {
	cmd := exec.Command(lego, "-h")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", output)
		t.Fatal(err)
	}

	fmt.Fprintf(os.Stdout, "%s\n", output)
}

func TestChallengeHTTP(t *testing.T) {
	cleanLegoFiles()

	cmd := exec.Command(lego,
		"-m", "hubert@hubert.com",
		"-a",
		"-x", "dns-01",
		"-x", "tls-alpn-01",
		"-s", "https://localhost:14000/dir",
		"-d", "acme.wtf",
		"--http", ":5002",
		"--tls", ":5001",
		"run")

	cmd.Env = []string{"LEGO_CA_CERTIFICATES=./fixtures/certs/pebble.minica.pem"}

	output, err := cmd.CombinedOutput()
	if len(output) > 0 {
		log.Println(string(output))
	}
	if err != nil {
		t.Fatal(err)
	}
}

func TestChallengeTLS(t *testing.T) {
	cleanLegoFiles()

	cmd := exec.Command(lego,
		"-m", "hubert@hubert.com",
		"-a",
		"-x", "dns-01",
		"-x", "http-01",
		"-s", "https://localhost:14000/dir",
		"-d", "acme.wtf",
		"--http", ":5002",
		"--tls", ":5001",
		"run")

	cmd.Env = []string{"LEGO_CA_CERTIFICATES=./fixtures/certs/pebble.minica.pem"}

	output, err := cmd.CombinedOutput()
	if len(output) > 0 {
		log.Println(string(output))
	}
	if err != nil {
		t.Fatal(err)
	}
}
