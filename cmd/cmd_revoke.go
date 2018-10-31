package cmd

import (
	"io/ioutil"
	"path/filepath"

	"github.com/urfave/cli"
	"github.com/xenolf/lego/log"
)

func CreateRevoke() cli.Command {
	return cli.Command{
		Name:   "revoke",
		Usage:  "Revoke a certificate",
		Action: revoke,
	}
}

func revoke(c *cli.Context) error {
	conf, acc, client := setup(c)
	if acc.Registration == nil {
		log.Fatalf("Account %s is not registered. Use 'run' to register a new account.\n", acc.Email)
	}

	if err := checkFolder(conf.CertPath()); err != nil {
		log.Fatalf("Could not check/create path: %v", err)
	}

	for _, domain := range c.GlobalStringSlice("domains") {
		log.Printf("Trying to revoke certificate for domain %s", domain)

		certPath := filepath.Join(conf.CertPath(), domain+".crt")
		certBytes, err := ioutil.ReadFile(certPath)
		if err != nil {
			log.Println(err)
		}

		err = client.Certificate.RevokeCertificate(certBytes)
		if err != nil {
			log.Fatalf("Error while revoking the certificate for domain %s\n\t%v", domain, err)
		} else {
			log.Println("Certificate was revoked.")
		}
	}

	return nil
}
