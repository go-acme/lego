package e2e

import (
	"flag"
	"fmt"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(testMain(m))
}

func TestHelp(t *testing.T) {
	output, err := runLego("-h")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", output)
		t.Fatal(err)
	}

	fmt.Fprintf(os.Stdout, "%s\n", output)
}

func TestChallengeHTTP(t *testing.T) {
	cleanLegoFiles()

	output, err := runLego(
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
	cleanLegoFiles()

	output, err := runLego(
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
